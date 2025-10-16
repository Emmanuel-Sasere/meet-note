package summarization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"noted/internal/config"
	"os"
	"time"
)

	

		func SummarizeWithGemini(text string) (string, error) {
			// Prepare request body
		reqBody := GeminiRequest {
			Contents: []GeminiContent{
				{
					Parts: []GeminiPart{
						{
							Text: "Summarize this transcript: " + text,
						},
					},
				},
			},
		}
			
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return "", fmt.Errorf("failed to encode request body: %v", err)
		}

		req, err := http.NewRequest("POST", "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent", bytes.NewBuffer(jsonData))

		if err != nil {
			return "", fmt.Errorf("failed to create request: %v", err)
		}

		// Set HEADERS
		apiKey := config.GetGeminiKey()
		if apiKey == "" {
			return "", fmt.Errorf("GEMINI_API_KEY environment variable is not set")
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		//Make the request
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("request failed: %v", err)
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusTooManyRequests {
				//Try to get wait time from headers
				retryAfter := resp.Header.Get("Retry-After")
				if retryAfter != "" {
					retryAfter = "15 minutes" // fallback if no header
				}
				fmt.Errorf("Gemini API limit reached. Please wait %s before retrying.", retryAfter)
			}
			return "", fmt.Errorf("Gemini API error (%d): %s",resp.StatusCode, string(body))
		}

		var result GeminiResponse
		err = json.Unmarshal(body, &result)
		if err != nil {
			return "", fmt.Errorf("failed to decode response: %v", err)
		}

		if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
			return "", fmt.Errorf("invalid response: missing text content")
		}

		output := result.Candidates[0].Content.Parts[0].Text 

			return output, nil
		}



		func TranscribeWithGemini(rawText string) (string, error){
			//Prepare request body
			//Send to Gemini API
			//Parse response
			//Return cleaned transcript
		}