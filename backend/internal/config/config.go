package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetGeminiKey() string {
	 err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }

	
	key := os.Getenv("GEMINI_API_KEY")

	if key == "" {
		log.Fatal("GEMINI_API_KEY environment variable is not set")
	}
	return key
}

