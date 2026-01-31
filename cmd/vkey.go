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
	vkName   string
	vkKey    string
	vkConnID string
	vkTPS    float64
	vkTokens int64
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
			ID:              models.NewID(),
			Name:            vkName,
			Key:             vkKey,
			ConnectionID:    vkConnID,
			RateLimitTPS:    vkTPS,
			RateLimitTokens: vkTokens,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err = database.SaveVirtualKey(context.Background(), vk)
		if err != nil {
			return err
		}

		fmt.Printf("Virtual key added successfully: %s (Key: %s)\n", vk.Name, vk.Key)
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

		fmt.Printf("%-36s %-15s %-20s %-10s %-10s\n", "ID", "Name", "Key", "TPS", "Tokens")
		for _, v := range vks {
			fmt.Printf("%-36s %-15s %-20s %-10.2f %-10d\n", v.ID, v.Name, v.Key, v.RateLimitTPS, v.RateLimitTokens)
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
	addVkeyCmd.Flags().StringVar(&vkConnID, "conn-id", "", "ID of the real connection to map to")
	addVkeyCmd.Flags().Float64Var(&vkTPS, "tps", 0, "Rate limit: Requests per second")
	addVkeyCmd.Flags().Int64Var(&vkTokens, "tokens", 0, "Rate limit: Tokens per minute (simulated)")

	addVkeyCmd.MarkFlagRequired("name")
	addVkeyCmd.MarkFlagRequired("key")
	addVkeyCmd.MarkFlagRequired("conn-id")
}
