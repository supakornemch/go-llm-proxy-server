package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/supakornemchananon/go-llm-proxy-server/cmd"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cmd.Execute()
}
