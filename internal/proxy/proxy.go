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
	if modelAlias == "" {
		pathParts := strings.Split(c.Request.URL.Path, "/")
		for i, part := range pathParts {
			if part == "models" && i+1 < len(pathParts) {
				modelAlias = strings.Split(pathParts[i+1], ":")[0]
				break
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

	// Rewrite model name in body if it differs from remote model
	if pm.RemoteModel != "" && pm.RemoteModel != modelAlias {
		bodyObj["model"] = pm.RemoteModel
	}

	body, _ = json.Marshal(bodyObj)

	targetURL := strings.TrimSuffix(conn.Endpoint, "/")
	targetPath := strings.TrimPrefix(c.Request.URL.Path, "/")

	// Provider-specific routing logic
	if conn.Provider == "azure" {
		// Map OpenAI-style path to Azure Foundry path if it matches
		if targetPath == "v1/chat/completions" {
			targetPath = "models/chat/completions"
		}
		// If api-version isn't in endpoint or query, add default
		if !strings.Contains(targetURL, "api-version=") && !strings.Contains(c.Request.URL.RawQuery, "api-version=") {
			if c.Request.URL.RawQuery == "" {
				c.Request.URL.RawQuery = "api-version=2024-05-01-preview"
			} else {
				c.Request.URL.RawQuery += "&api-version=2024-05-01-preview"
			}
		}
	}

	targetURL = targetURL + "/" + targetPath
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(body))
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

	if conn.Provider == "azure" {
		req.Header.Set("api-key", conn.APIKey)
	} else {
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
