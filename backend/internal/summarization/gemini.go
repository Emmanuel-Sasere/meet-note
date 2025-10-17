package summarization

import (
	"bytes"
	"encoding/base64"
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
			
			apiKey := config.GetGeminiKey()
		if apiKey == "" {
			return "", fmt.Errorf("GEMINI_API_KEY environment variable is not set")
		}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return "", fmt.Errorf("failed to encode request body: %v", err)
		}

		url :=  "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

		if err != nil {
			return "", fmt.Errorf("failed to create request: %v", err)
		}

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
				return "", fmt.Errorf("gemini API limit reached. Please wait %s before retrying", retryAfter)
			}
			return "", fmt.Errorf("gemini API error (%d): %s",resp.StatusCode, string(body))
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



		func TranscribeWithGemini(audioFilePath string) (string, error){
			//Load Gemini API Key
			apiKey := config.GetGeminiKey()
			if apiKey == "" {
				return "", fmt.Errorf("GEMINI_API_KEY environment variable is not set")
			}
			//Read audio file
			audioData, err := os.ReadFile(audioFilePath)
			if err != nil {
				return "", fmt.Errorf("failed to read audio file: %v", err)
			}
			//Convert audio to base64
			encodedAudio := base64.StdEncoding.EncodeToString(audioData)
			//Prepare request body
			requestBody := GeminiRequest{
				Contents: []GeminiContent{
					{
						Parts: []GeminiPart{
							{Text: "Transcribe this audio into text, See if you can get name of speakers and label them accordingly, also ensure neat notes:"}, 
							{
								InlineData: &InlineData{
								MimeType: "audio/mp3", 
								Data: encodedAudio,
						},
						},
						},
					},
				},
			}

			//Turn request body to JSON
			jsonData, err := json.Marshal(requestBody)
			if err != nil {
				return "", fmt.Errorf("failed to marshal JSON: %v", err)
			}
			//Send to Gemini API
			url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				return "", fmt.Errorf("failed to send request: %v", err)
			}
			defer resp.Body.Close()

			// Read response body
			bodyBytes, _ := io.ReadAll(resp.Body)
			//Handle http errors
			if resp.StatusCode != http.StatusOK {
				return "", fmt.Errorf("gemini API error (%d): %s", resp.StatusCode, string(bodyBytes))
			}

			//Decode the API response
			var geminiResp GeminiResponse
		if 	err = json.Unmarshal(bodyBytes, &geminiResp); err != nil {
				return "", fmt.Errorf("failed to decode response: %v", err)
			}



			//Extract transcript text
			if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
				return geminiResp.Candidates[0].Content.Parts[0].Text, nil
			}

			//Return cleaned transcript
			return "", fmt.Errorf("invalid response: missing text content")
		}