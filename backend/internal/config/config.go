package config

import (
	"os"
	
)




type APIKeys struct {
	Gemini     string
	Groq       string

}



func GetAPIKeys() APIKeys {
	return APIKeys{
		Gemini:     os.Getenv("GEMINI_API_KEY"),
		Groq:       os.Getenv("GROQ_API_KEY"),

	}
}

func GetGeminiKey() string {
	return os.Getenv("GEMINI_API_KEY")
}

func GetGroqKey() string {
	return os.Getenv("GROQ_API_KEY")
}

