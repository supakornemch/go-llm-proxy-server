package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/supakornemchananon/go-llm-proxy-server/internal/db"
	"github.com/supakornemchananon/go-llm-proxy-server/internal/models"
)

var (
	connName   string
	provider   string
	endpoint   string
	apiKey     string
	model      string
	deployment string
)

var connCmd = &cobra.Command{
	Use:   "connection",
	Short: "Manage real LLM connections",
}

var addConnCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new LLM connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.InitDB(dbType, dsn)
		if err != nil {
			return err
		}

		conn := &models.Connection{
			ID:             models.NewID(),
			Name:           connName,
			Provider:       provider,
			Endpoint:       endpoint,
			APIKey:         apiKey,
			Model:          model,
			DeploymentName: deployment,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err = database.SaveConnection(context.Background(), conn)
		if err != nil {
			return err
		}

		fmt.Printf("Connection added successfully: %s (ID: %s)\n", conn.Name, conn.ID)
		return nil
	},
}

var listConnCmd = &cobra.Command{
	Use:   "list",
	Short: "List all LLM connections",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.InitDB(dbType, dsn)
		if err != nil {
			return err
		}

		conns, err := database.ListConnections(context.Background())
		if err != nil {
			return err
		}

		fmt.Printf("%-36s %-15s %-10s %-20s\n", "ID", "Name", "Provider", "Model")
		for _, c := range conns {
			fmt.Printf("%-36s %-15s %-10s %-20s\n", c.ID, c.Name, c.Provider, c.Model)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(connCmd)
	connCmd.AddCommand(addConnCmd)
	connCmd.AddCommand(listConnCmd)

	addConnCmd.Flags().StringVar(&connName, "name", "", "Name of the connection")
	addConnCmd.Flags().StringVar(&provider, "provider", "", "LLM Provider (openai, azure, etc.)")
	addConnCmd.Flags().StringVar(&endpoint, "endpoint", "", "Endpoint URL")
	addConnCmd.Flags().StringVar(&apiKey, "api-key", "", "API Key")
	addConnCmd.Flags().StringVar(&model, "model", "", "Model Name")
	addConnCmd.Flags().StringVar(&deployment, "deployment", "", "Deployment Name (optional)")

	addConnCmd.MarkFlagRequired("name")
	addConnCmd.MarkFlagRequired("provider")
	addConnCmd.MarkFlagRequired("endpoint")
	addConnCmd.MarkFlagRequired("api-key")
}
