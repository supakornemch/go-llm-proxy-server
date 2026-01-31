package proxy

import (
	"bytes"
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

	limiter := p.ratelimitManager.GetLimiter(vk.Key, vk.RateLimitTPS, vk.RateLimitTokens)
	if !limiter.AllowTPS() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "TPS limit exceeded"})
		return
	}

	conn, err := p.db.GetConnection(c.Request.Context(), vk.ConnectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get real connection"})
		return
	}

	targetURL := strings.TrimSuffix(conn.Endpoint, "/")
	targetURL += "/" + strings.TrimPrefix(c.Request.URL.Path, "/")
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	if !limiter.AllowTokens(1) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Token limit exceeded"})
		return
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
	} else if conn.Provider == "google" {
		req.Header.Set("x-goog-api-key", conn.APIKey)
		q := req.URL.Query()
		q.Set("key", conn.APIKey)
		req.URL.RawQuery = q.Encode()
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
