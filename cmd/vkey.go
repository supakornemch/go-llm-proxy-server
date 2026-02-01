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
			ID:        models.NewID(),
			Name:      vkName,
			Key:       vkKey,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = database.SaveVirtualKey(context.Background(), vk)
		if err != nil {
			return err
		}

		fmt.Printf("Virtual key added successfully: %s (Key: %s) [ID: %s]\n", vk.Name, vk.Key, vk.ID)
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
	// Note: --conn-id is not used directly in 'vkey add' anymore.
	// Use './llm-proxy assign' to link a Virtual Key to a Model/Connection.
	addVkeyCmd.Flags().StringVar(&vkConnID, "conn-id", "", "Connection ID (deprecated in vkey add, use assign instead)")

	addVkeyCmd.MarkFlagRequired("name")
	addVkeyCmd.MarkFlagRequired("key")
}
