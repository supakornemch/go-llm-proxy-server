package cmd

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/supakornemchananon/go-llm-proxy-server/internal/db"
	"github.com/supakornemchananon/go-llm-proxy-server/internal/server"
)

var port int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the LLM proxy server",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.InitDB(dbType, dsn)
		if err != nil {
			return err
		}

		// Auto-seed database from environment variables (Cloud-friendly)
		db.AutoSeed(database)

		return server.Start(database, port)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	defaultPort := 8132
	if pStr, ok := os.LookupEnv("PORT"); ok {
		if p, err := strconv.Atoi(pStr); err == nil {
			defaultPort = p
		}
	}

	serveCmd.Flags().IntVarP(&port, "port", "p", defaultPort, "Port to listen on")
}
