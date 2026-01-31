package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/supakornemchananon/go-llm-proxy-server/internal/db"
	"github.com/supakornemchananon/go-llm-proxy-server/internal/ratelimit"
)

type Proxy struct {
	db               db.DB
	ratelimitManager *ratelimit.Manager
}

func NewProxy(database db.DB) *Proxy {
	return &Proxy{
		db:               database,
		ratelimitManager: ratelimit.NewManager(),
	}
}

func (p *Proxy) HandleProxy(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid authorization header"})
		return
	}
	vkey := strings.TrimPrefix(authHeader, "Bearer ")

	vk, err := p.db.GetVirtualKey(c.Request.Context(), vkey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid virtual key"})
		return
	}

	// Read body to identify requested model alias
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	var bodyObj map[string]interface{}
	json.Unmarshal(body, &bodyObj)
	modelAlias, _ := bodyObj["model"].(string)

	// Fallback: Try to extract model from URL if not in body (common in Gemini SDKs)
	// Example path: /v1/models/gemini-1.5-flash:generateContent
	// Example Vertex path: /v1/projects/.../locations/.../publishers/google/models/gemini-1.5-flash:streamGenerateContent
	if modelAlias == "" {
		pathParts := strings.Split(c.Request.URL.Path, "/")
		for i, part := range pathParts {
			if part == "models" && i+1 < len(pathParts) {
				modelAlias = strings.Split(pathParts[i+1], ":")[0]
				// Check if it's the specific part we want (avoiding mid-path matching if possible)
				if modelAlias != "" {
					break
				}
			}
		}
	}

	if modelAlias == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'model' in request body or URL path"})
		return
	}

	// Get assignment for this virtual key and model alias
	vka, err := p.db.GetVirtualKeyAssignment(c.Request.Context(), vk.ID, modelAlias)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Virtual key not authorized for model: " + modelAlias})
		return
	}

	// Get the actual provider model
	pm, err := p.db.GetProviderModel(c.Request.Context(), vka.ProviderModelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Target model not found"})
		return
	}

	// Get credentials
	conn, err := p.db.GetConnection(c.Request.Context(), pm.ConnectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Provider connection not found"})
		return
	}

	// Rate limiting (per key per model)
	limiter := p.ratelimitManager.GetLimiter(vk.Key+":"+modelAlias, vka.RateLimitTPS, vka.RateLimitTokens)
	if !limiter.AllowTPS() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "TPS limit exceeded"})
		return
	}

	if !limiter.AllowTokens(1) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Token limit exceeded"})
		return
	}

	// Prepare target path and check for model replacement in URL
	targetURLStr := strings.TrimSuffix(conn.Endpoint, "/")
	targetPath := strings.TrimPrefix(c.Request.URL.Path, "/")

	if pm.RemoteModel != "" && pm.RemoteModel != modelAlias {
		// 1. Rewrite in body ONLY if it existed (OpenAI style)
		if _, exists := bodyObj["model"]; exists {
			bodyObj["model"] = pm.RemoteModel
		}

		// 2. Rewrite in URL path (Native Gemini/Vertex style)
		targetPath = strings.Replace(targetPath, modelAlias, pm.RemoteModel, 1)
	}

	body, _ = json.Marshal(bodyObj)

	// Build raw query. We merge existing endpoint query with request query.
	rawQuery := c.Request.URL.RawQuery

	// Provider-specific routing logic
	switch conn.Provider {
	case "aws":
		// AWS Bedrock (Claude) often expects /model/MODEL_ID/invoke or similar
		// If it's a direct endpoint to Bedrock Runtime, we might need to map it.
		// For OpenAI compatibility, we assume the user might be using a proxy that maps it,
		// but if we are hitting Bedrock directly, we need to handle the path.
		if targetPath == "v1/chat/completions" || targetPath == "chat/completions" {
			targetPath = "model/" + pm.RemoteModel + "/invoke"
		}
	case "azure":
		// Map OpenAI-style path to Azure Foundry path if it matches
		if targetPath == "v1/chat/completions" {
			targetPath = "models/chat/completions"
		}
		// If api-version isn't in endpoint or query, add default
		if !strings.Contains(targetURLStr, "api-version=") && !strings.Contains(rawQuery, "api-version=") {
			if rawQuery == "" {
				rawQuery = "api-version=2024-05-01-preview"
			} else {
				rawQuery += "&api-version=2024-05-01-preview"
			}
		}
	case "google":
		// Handle path mapping:
		// Convert Vertex-style path to AI Studio-style if the endpoint is AI Studio.
		isVertex := strings.Contains(targetURLStr, "aiplatform.googleapis.com")
		if !isVertex {
			targetPath = strings.Replace(targetPath, "publishers/google/", "", 1)
		}

		// Google AI Studio OpenAI-compatible endpoint doesn't want the /v1/ prefix
		if strings.HasSuffix(targetURLStr, "/openai") && strings.HasPrefix(targetPath, "v1/") {
			targetPath = strings.TrimPrefix(targetPath, "v1/")
		}

		// Strip client's 'key' param and inject our own
		params := strings.Split(rawQuery, "&")
		var newParams []string
		for _, p := range params {
			if !strings.HasPrefix(p, "key=") && p != "" {
				newParams = append(newParams, p)
			}
		}
		newParams = append(newParams, "key="+conn.APIKey)
		rawQuery = strings.Join(newParams, "&")
	}

	// Final URL Construction
	finalURL := targetURLStr + "/" + targetPath
	if rawQuery != "" {
		if strings.Contains(finalURL, "?") {
			finalURL += "&" + rawQuery
		} else {
			finalURL += "?" + rawQuery
		}
	}

	req, err := http.NewRequest(c.Request.Method, finalURL, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	for k, v := range c.Request.Header {
		lowerK := strings.ToLower(k)
		if lowerK == "authorization" || lowerK == "host" || lowerK == "api-key" {
			continue
		}
		req.Header[k] = v
	}

	switch conn.Provider {
	case "azure":
		req.Header.Set("api-key", conn.APIKey)
		req.Header.Set("Authorization", "Bearer "+conn.APIKey)
	case "google":
		req.Header.Set("x-goog-api-key", conn.APIKey)
		// Only use Bearer auth if it's an OAuth token (starts with ya29).
		// API Keys (like Vertex API keys starting with AQ.) should not use Bearer.
		if strings.HasPrefix(conn.APIKey, "ya29.") {
			req.Header.Set("Authorization", "Bearer "+conn.APIKey)
		}
	default:
		req.Header.Set("Authorization", "Bearer "+conn.APIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to call LLM provider", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		c.Writer.Header()[k] = v
	}
	c.Writer.WriteHeader(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
