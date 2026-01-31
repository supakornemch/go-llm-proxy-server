package models

import (
	"time"

	"github.com/google/uuid"
)

type Connection struct {
	ID             string    `gorm:"primaryKey" bson:"_id" json:"id"`
	Name           string    `gorm:"uniqueIndex" bson:"name" json:"name"`
	Provider       string    `bson:"provider" json:"provider"` // e.g., openai, azure, anthropic
	Endpoint       string    `bson:"endpoint" json:"endpoint"`
	APIKey         string    `bson:"api_key" json:"api_key"`
	Model          string    `bson:"model" json:"model"`
	DeploymentName string    `bson:"deployment_name" json:"deployment_name"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}

type VirtualKey struct {
	ID              string    `gorm:"primaryKey" bson:"_id" json:"id"`
	Name            string    `gorm:"uniqueIndex" bson:"name" json:"name"`
	Key             string    `gorm:"uniqueIndex" bson:"key" json:"key"`
	ConnectionID    string    `gorm:"index" bson:"connection_id" json:"connection_id"`
	RateLimitTPS    float64   `bson:"rate_limit_tps" json:"rate_limit_tps"`
	RateLimitTokens int64     `bson:"rate_limit_tokens" json:"rate_limit_tokens"`
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time `bson:"updated_at" json:"updated_at"`
}

func NewID() string {
	return uuid.New().String()
}
