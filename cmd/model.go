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
	pmName       string
	pmRemote     string
	pmDeployment string
	pmConnID     string

	asVKID    string
	asModelID string
	asAlias   string
	asTPS     float64
	asTokens  int64
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage provider models available on connections",
}

var addModelCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a model to a connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.InitDB(dbType, dsn)
		if err != nil {
			return err
		}
		pm := &models.ProviderModel{
			ID:             models.NewID(),
			ConnectionID:   pmConnID,
			Name:           pmName,
			RemoteModel:    pmRemote,
			DeploymentName: pmDeployment,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		err = database.SaveProviderModel(context.Background(), pm)
		if err != nil {
			return err
		}
		fmt.Printf("Model %s added to connection %s [ID: %s]\n", pm.Name, pm.ConnectionID, pm.ID)
		return nil
	},
}

var listModelCmd = &cobra.Command{
	Use:   "list",
	Short: "List all provider models",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.InitDB(dbType, dsn)
		if err != nil {
			return err
		}
		pms, err := database.ListProviderModels(context.Background(), pmConnID)
		if err != nil {
			return err
		}
		fmt.Printf("%-36s %-15s %-20s %-20s\n", "ID", "Name", "RemoteModel", "ConnID")
		for _, m := range pms {
			fmt.Printf("%-36s %-15s %-20s %-20s\n", m.ID, m.Name, m.RemoteModel, m.ConnectionID)
		}
		return nil
	},
}

var assignCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign a model to a virtual key with rate limits",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.InitDB(dbType, dsn)
		if err != nil {
			return err
		}
		as := &models.VirtualKeyAssignment{
			ID:              models.NewID(),
			VirtualKeyID:    asVKID,
			ProviderModelID: asModelID,
			ModelAlias:      asAlias,
			RateLimitTPS:    asTPS,
			RateLimitTokens: asTokens,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		err = database.SaveVirtualKeyAssignment(context.Background(), as)
		if err != nil {
			return err
		}
		fmt.Printf("Assigned model %s (alias: %s) to virtual key %s\n", as.ProviderModelID, as.ModelAlias, as.VirtualKeyID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(modelCmd)
	modelCmd.AddCommand(addModelCmd)
	modelCmd.AddCommand(listModelCmd)
	rootCmd.AddCommand(assignCmd)

	addModelCmd.Flags().StringVar(&pmName, "name", "", "Display name for the model")
	addModelCmd.Flags().StringVar(&pmRemote, "remote", "", "Remote model name (e.g. gpt-4)")
	addModelCmd.Flags().StringVar(&pmDeployment, "deployment", "", "Azure deployment name (optional)")
	addModelCmd.Flags().StringVar(&pmConnID, "conn-id", "", "Connection ID")
	addModelCmd.MarkFlagRequired("name")
	addModelCmd.MarkFlagRequired("remote")
	addModelCmd.MarkFlagRequired("conn-id")

	listModelCmd.Flags().StringVar(&pmConnID, "conn-id", "", "Filter by connection ID")

	assignCmd.Flags().StringVar(&asVKID, "vkey-id", "", "Virtual Key ID")
	assignCmd.Flags().StringVar(&asModelID, "model-id", "", "Provider Model ID")
	assignCmd.Flags().StringVar(&asAlias, "alias", "", "The model name client will use (e.g. 'gpt-4')")
	assignCmd.Flags().Float64Var(&asTPS, "tps", 1.0, "TPS limit")
	assignCmd.Flags().Int64Var(&asTokens, "tokens", 1000, "Token limit (simulated)")
	assignCmd.MarkFlagRequired("vkey-id")
	assignCmd.MarkFlagRequired("model-id")
	assignCmd.MarkFlagRequired("alias")
}
