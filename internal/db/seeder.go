package db

import (
	"context"
	"log"
	"os"
	"strconv"

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
	modelName := os.Getenv("MASTER_CONN_MODEL") // This will be used as both Model Name and Alias
	apiKey := os.Getenv("MASTER_CONN_API_KEY")

	if provider == "" || endpoint == "" || apiKey == "" {
		log.Println("⚠️ Auto-seeding skipped: Missing required MASTER_CONN variables")
		return
	}

	// 1. Connection
	conns, _ := database.ListConnections(ctx)
	var existingConn *models.Connection
	for _, c := range conns {
		if c.Name == name {
			existingConn = &c
			break
		}
	}

	var connID string
	if existingConn == nil {
		connID = models.NewID()
		conn := &models.Connection{
			ID:       connID,
			Name:     name,
			Provider: provider,
			Endpoint: endpoint,
			APIKey:   apiKey,
		}
		if err := database.SaveConnection(ctx, conn); err != nil {
			log.Printf("❌ Failed to auto-seed connection: %v\n", err)
			return
		}
		log.Printf("✅ Auto-seeded connection: %s\n", name)
	} else {
		connID = existingConn.ID
	}

	// 2. Provider Model
	var modelID string
	if modelName != "" {
		pms, _ := database.ListProviderModels(ctx, connID)
		var existingPM *models.ProviderModel
		for _, m := range pms {
			if m.Name == modelName {
				existingPM = &m
				break
			}
		}

		if existingPM == nil {
			modelID = models.NewID()
			pm := &models.ProviderModel{
				ID:           modelID,
				ConnectionID: connID,
				Name:         modelName,
				RemoteModel:  modelName,
			}
			database.SaveProviderModel(ctx, pm)
			log.Printf("✅ Auto-seeded model: %s\n", modelName)
		} else {
			modelID = existingPM.ID
		}
	}

	// 3. Virtual Key Seed
	vkeyName := os.Getenv("MASTER_VKEY_NAME")
	vkeyValue := os.Getenv("MASTER_VKEY_KEY")
	if vkeyName != "" && vkeyValue != "" {
		vks, _ := database.ListVirtualKeys(ctx)
		var existingVK *models.VirtualKey
		for _, v := range vks {
			if v.Key == vkeyValue {
				existingVK = &v
				break
			}
		}

		var vkID string
		if existingVK == nil {
			vkID = models.NewID()
			vk := &models.VirtualKey{
				ID:   vkID,
				Name: vkeyName,
				Key:  vkeyValue,
			}
			if err := database.SaveVirtualKey(ctx, vk); err != nil {
				log.Printf("❌ Failed to auto-seed virtual key: %v\n", err)
				return
			}
			log.Printf("✅ Auto-seeded virtual key: %s (Key: %s)\n", vkeyName, vkeyValue)
		} else {
			vkID = existingVK.ID
		}

		// 4. Assignment
		if modelID != "" && vkID != "" {
			assignments, _ := database.ListVirtualKeyAssignments(ctx, vkID)
			exists := false
			for _, a := range assignments {
				if a.ProviderModelID == modelID {
					exists = true
					break
				}
			}

			if !exists {
				tps, _ := strconv.ParseFloat(os.Getenv("MASTER_VKEY_TPS"), 64)
				tokensInt, _ := strconv.Atoi(os.Getenv("MASTER_VKEY_TOKENS"))
				if tps == 0 {
					tps = 10
				}
				if tokensInt == 0 {
					tokensInt = 50000
				}

				as := &models.VirtualKeyAssignment{
					ID:              models.NewID(),
					VirtualKeyID:    vkID,
					ProviderModelID: modelID,
					ModelAlias:      modelName,
					RateLimitTPS:    tps,
					RateLimitTokens: int64(tokensInt),
				}
				database.SaveVirtualKeyAssignment(ctx, as)
				log.Printf("✅ Auto-seeded assignment: %s -> %s\n", vkeyName, modelName)
			}
		}
	}
}
