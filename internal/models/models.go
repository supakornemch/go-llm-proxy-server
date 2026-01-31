package models

import (
	"time"

	"github.com/google/uuid"
)

type Connection struct {
	ID        string    `gorm:"primaryKey" bson:"_id" json:"id"`
	Name      string    `gorm:"uniqueIndex" bson:"name" json:"name"`
	Provider  string    `bson:"provider" json:"provider"` // e.g., openai, azure, google
	Endpoint  string    `bson:"endpoint" json:"endpoint"`
	APIKey    string    `bson:"api_key" json:"api_key"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type ProviderModel struct {
	ID             string    `gorm:"primaryKey" bson:"_id" json:"id"`
	ConnectionID   string    `gorm:"index" bson:"connection_id" json:"connection_id"`
	Name           string    `bson:"name" json:"name"`                       // Name used by provider or general identifier
	RemoteModel    string    `bson:"remote_model" json:"remote_model"`       // Internal model ID
	DeploymentName string    `bson:"deployment_name" json:"deployment_name"` // For Azure
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}

type VirtualKey struct {
	ID        string    `gorm:"primaryKey" bson:"_id" json:"id"`
	Name      string    `gorm:"uniqueIndex" bson:"name" json:"name"`
	Key       string    `gorm:"index" bson:"key" json:"key"`          // Encrypted
	KeyHash   string    `gorm:"uniqueIndex" bson:"key_hash" json:"-"` // SHA-256 hash for lookup
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type VirtualKeyAssignment struct {
	ID              string    `gorm:"primaryKey" bson:"_id" json:"id"`
	VirtualKeyID    string    `gorm:"index" bson:"virtual_key_id" json:"virtual_key_id"`
	ProviderModelID string    `gorm:"index" bson:"provider_model_id" json:"provider_model_id"`
	ModelAlias      string    `bson:"model_alias" json:"model_alias"` // The model name the user sends in request
	RateLimitTPS    float64   `bson:"rate_limit_tps" json:"rate_limit_tps"`
	RateLimitTokens int64     `bson:"rate_limit_tokens" json:"rate_limit_tokens"`
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time `bson:"updated_at" json:"updated_at"`
}

func NewID() string {
	return uuid.New().String()
}
