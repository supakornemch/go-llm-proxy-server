package db

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/supakornemchananon/go-llm-proxy-server/internal/models"
)

// AutoSeed checks environment variables to automatically create a master connection and virtual key.
// This is useful for CI/CD or cloud deployments like Azure App Service.
func AutoSeed(database DB) {
	ctx := context.Background()

	name := os.Getenv("MASTER_CONN_NAME")
	if name == "" {
		return // No seeding requested
	}

	provider := os.Getenv("MASTER_CONN_PROVIDER")
	endpoint := os.Getenv("MASTER_CONN_ENDPOINT")
	model := os.Getenv("MASTER_CONN_MODEL")
	apiKey := os.Getenv("MASTER_CONN_API_KEY")

	if provider == "" || endpoint == "" || apiKey == "" {
		log.Println("⚠️ Auto-seeding skipped: Missing required MASTER_CONN variables")
		return
	}

	// Check if already exists
	conns, _ := database.ListConnections(ctx)
	var existingConn *models.Connection
	for _, c := range conns {
		if c.Name == name {
			existingConn = &c
			break
		}
	}

	connID := ""
	if existingConn == nil {
		connID = uuid.New().String()
		conn := &models.Connection{
			ID:       connID,
			Name:     name,
			Provider: provider,
			Endpoint: endpoint,
			Model:    model,
			APIKey:   apiKey,
		}
		if err := database.SaveConnection(ctx, conn); err != nil {
			log.Printf("❌ Failed to auto-seed connection: %v\n", err)
			return
		}
		log.Printf("✅ Auto-seeded connection: %s\n", name)
	} else {
		connID = existingConn.ID
		log.Printf("ℹ️ Connection %s already exists, skipping seed\n", name)
	}

	// Virtual Key Seed
	vkeyName := os.Getenv("MASTER_VKEY_NAME")
	vkeyValue := os.Getenv("MASTER_VKEY_KEY")
	if vkeyName != "" && vkeyValue != "" {
		vks, _ := database.ListVirtualKeys(ctx)
		exists := false
		for _, v := range vks {
			if v.Key == vkeyValue {
				exists = true
				break
			}
		}

		if !exists {
			tps, _ := strconv.ParseFloat(os.Getenv("MASTER_VKEY_TPS"), 64)
			tokensInt, _ := strconv.Atoi(os.Getenv("MASTER_VKEY_TOKENS"))
			if tps == 0 {
				tps = 5
			}
			if tokensInt == 0 {
				tokensInt = 10000
			}

			vk := &models.VirtualKey{
				ID:              uuid.New().String(),
				Name:            vkeyName,
				Key:             vkeyValue,
				ConnectionID:    connID,
				RateLimitTPS:    tps,
				RateLimitTokens: int64(tokensInt),
			}
			if err := database.SaveVirtualKey(ctx, vk); err != nil {
				log.Printf("❌ Failed to auto-seed virtual key: %v\n", err)
			} else {
				log.Printf("✅ Auto-seeded virtual key: %s (Key: %s)\n", vkeyName, vkeyValue)
			}
		}
	}
}
