package main

import (
	"fmt"
	"log"
	"net/http"
	"noted/internal/config"
	"noted/internal/routes"

	"github.com/joho/godotenv"
)


func main() {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found, using system environment variables.")
	}

	//Load config (like API Key)
	if config.GetAPIKeys().Gemini == "" {
		log.Fatal("❌GEMINI_API_KEY is not set. Please set it in your environment variables.")
	} 
	
	if config.GetAPIKeys().Groq == "" {
		log.Fatal("❌ GROQ_API_KEY is not set. Please set it in your environment variables.")
	}


	//Register all routes from internal/routes
	router := routes.RegisterRoutes()




		//Define server Port
	port := ":8080"

	//Start Http server
	fmt.Println(" Server is running on port " + port)
	if err := http.ListenAndServe(
	port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	

}