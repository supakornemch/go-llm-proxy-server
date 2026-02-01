package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/supakornemchananon/go-llm-proxy-server/internal/db"
	"github.com/supakornemchananon/go-llm-proxy-server/internal/models"
	"github.com/supakornemchananon/go-llm-proxy-server/pkg/cryptoutil"
)

var (
	vkName    string
	vkKey     string
	vkConnID  string
	vkModelID string
	vkTPS     float64
	vkTokens  int64
)

var vkeyCmd = &cobra.Command{
	Use:   "vkey",
	Short: "Manage virtual keys",
}

var addVkeyCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new virtual key",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.InitDB(dbType, dsn)
		if err != nil {
			return err
		}

		vk := &models.VirtualKey{
			ID:        models.NewID(),
			Name:      vkName,
			Key:       vkKey,
			KeyHash:   cryptoutil.HashKey(vkKey),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = database.SaveVirtualKey(context.Background(), vk)
		if err != nil {
			return err
		}

		fmt.Printf("Virtual key added successfully: %s (Key: %s) [ID: %s]\n", vk.Name, vk.Key, vk.ID)

		// Case 1: If model-id is provided, assign only that specific model
		if vkModelID != "" {
			pm, err := database.GetProviderModel(context.Background(), vkModelID)
			if err != nil {
				return fmt.Errorf("failed to find model %s: %v", vkModelID, err)
			}
			as := &models.VirtualKeyAssignment{
				ID:              models.NewID(),
				VirtualKeyID:    vk.ID,
				ProviderModelID: pm.ID,
				ModelAlias:      pm.Name,
				RateLimitTPS:    vkTPS,
				RateLimitTokens: vkTokens,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			if err := database.SaveVirtualKeyAssignment(context.Background(), as); err != nil {
				return fmt.Errorf("failed to assign model: %v", err)
			}
			fmt.Printf("Assigned model: %s (alias: %s)\n", pm.Name, pm.Name)
		}

		// Case 2: If conn-id is provided, automatically assign all models in that connection to this key
		if vkConnID != "" {
			pms, err := database.ListProviderModels(context.Background(), vkConnID)
			if err != nil {
				return fmt.Errorf("failed to list models for connection %s: %v", vkConnID, err)
			}
			for _, pm := range pms {
				as := &models.VirtualKeyAssignment{
					ID:              models.NewID(),
					VirtualKeyID:    vk.ID,
					ProviderModelID: pm.ID,
					ModelAlias:      pm.Name,
					RateLimitTPS:    vkTPS,
					RateLimitTokens: vkTokens,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}
				if err := database.SaveVirtualKeyAssignment(context.Background(), as); err != nil {
					fmt.Printf("Warning: failed to auto-assign model %s: %v\n", pm.Name, err)
				} else {
					fmt.Printf("Auto-assigned model: %s (alias: %s)\n", pm.Name, pm.Name)
				}
			}
		}

		return nil
	},
}

var listVkeyCmd = &cobra.Command{
	Use:   "list",
	Short: "List all virtual keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.InitDB(dbType, dsn)
		if err != nil {
			return err
		}

		vks, err := database.ListVirtualKeys(context.Background())
		if err != nil {
			return err
		}

		fmt.Printf("%-36s %-15s %-20s\n", "ID", "Name", "Key")
		for _, v := range vks {
			fmt.Printf("%-36s %-15s %-20s\n", v.ID, v.Name, v.Key)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(vkeyCmd)
	vkeyCmd.AddCommand(addVkeyCmd)
	vkeyCmd.AddCommand(listVkeyCmd)

	addVkeyCmd.Flags().StringVar(&vkName, "name", "", "Name of the virtual key")
	addVkeyCmd.Flags().StringVar(&vkKey, "key", "", "Actual virtual key value for users")
	addVkeyCmd.Flags().StringVar(&vkConnID, "conn-id", "", "Connection ID to auto-assign all models")
	addVkeyCmd.Flags().StringVar(&vkModelID, "model-id", "", "Model ID to assign a single model")
	addVkeyCmd.Flags().Float64Var(&vkTPS, "tps", 10.0, "TPS limit for the assigned models")
	addVkeyCmd.Flags().Int64Var(&vkTokens, "tokens", 100000, "Token limit for the assigned models")

	addVkeyCmd.MarkFlagRequired("name")
	addVkeyCmd.MarkFlagRequired("key")
}
