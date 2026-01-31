package cmd

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	dbType string
	dsn    string
)

var rootCmd = &cobra.Command{
	Use:   "llm-proxy",
	Short: "A proxy server for LLM connections with virtual keys and rate limiting",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Load .env here to ensure it's available for flag defaults if needed,
	// though PersistentFlags defaults are set at init time.
	_ = godotenv.Load()

	rootCmd.PersistentFlags().StringVar(&dbType, "db-type", getEnv("DB_TYPE", "sqlite"), "Database type (sqlite, postgres, mssql, mongodb)")
	rootCmd.PersistentFlags().StringVar(&dsn, "dsn", getEnv("DB_DSN", "llm_proxy.db"), "Database connection string")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
