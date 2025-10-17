package config

import (
	"os"
	"log"
)

func GetGeminiKey() string {
	key := os.Getenv("GEMINI_API_KEY")

	if key == "" {
		log.Fatal("GEMINI_API_KEY environment variable is not set")
	}
	return key
}

