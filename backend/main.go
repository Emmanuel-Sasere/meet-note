package main

import (
	"fmt"
	"noted/api"
	"noted/internal/config"
)




func main() {
	key := config.GetGeminiKey()
	fmt.Println("Gemini API Key:", key[:8]+"...") 
	api.StartServer()
}