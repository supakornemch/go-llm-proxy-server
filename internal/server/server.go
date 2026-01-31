package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/supakornemchananon/go-llm-proxy-server/internal/db"
	"github.com/supakornemchananon/go-llm-proxy-server/internal/proxy"
)

func Start(database db.DB, port int) error {
	r := gin.Default()

	p := proxy.NewProxy(database)

	r.NoRoute(p.HandleProxy)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting LLM Proxy Server on port %d...\n", port)
	return r.Run(addr)
}
