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
	"path/filepath"
	"strings"
	"time"
)

func SummarizeWithGemini(text string) (string, error) {
	// Prepare request body
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{
						Text: `Please provide a concise summary of this meeting transcript. Include:
1. Main topics discussed
2. Key decisions made
3. Action items (if any)
4. Important points raised

Keep the summary clear and well-organized.

Transcript:
` + text,
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

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=" + apiKey

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{Timeout: 60 * time.Second}
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
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter == "" {
				retryAfter = "15 minutes"
			}
			return "", fmt.Errorf("gemini API limit reached. Please wait %s before retrying", retryAfter)
		}
		return "", fmt.Errorf("gemini API error (%d): %s", resp.StatusCode, string(body))
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

func TranscribeWithGemini(audioFilePath string) (string, error) {
	// Load Gemini API Key
	apiKey := config.GetGeminiKey()
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	fmt.Println("📂 Reading audio file:", audioFilePath)

	// Read audio file
	audioData, err := os.ReadFile(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read audio file: %v", err)
	}

	fmt.Printf("📊 Audio file size: %d bytes\n", len(audioData))

	// Convert audio to base64
	encodedAudio := base64.StdEncoding.EncodeToString(audioData)

	// Detect MIME type from file extension
	mimeType := detectMimeType(audioFilePath)
	fmt.Println("🎵 Detected MIME type:", mimeType)

	// Prepare request body - CRITICAL: Text and InlineData in separate parts
	requestBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{
						Text: `Please transcribe this audio into clear, well-formatted notes. Focus on 
capturing what was discussed, including key points, decisions, and any 
important details mentioned.`,
					},
					{
						InlineData: &InlineData{
							MimeType: mimeType,
							Data:     encodedAudio,
						},
					},
				},
			},
		},
	}

	// Turn request body to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Debug: Print first 500 chars of JSON (without the huge base64 data)
	debugJSON := string(jsonData)
	if len(debugJSON) > 500 {
		debugJSON = debugJSON[:500] + "...[truncated]"
	}
	fmt.Println("📤 Request JSON structure:", debugJSON)

	// Send to Gemini API
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=" + apiKey
	
	fmt.Println("🌐 Sending request to Gemini API...")
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second} // Longer timeout for audio
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Printf("📥 Response status: %d\n", resp.StatusCode)

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		// Pretty print error for debugging
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
			fmt.Println("❌ API Error Response:", prettyJSON.String())
		}
		return "", fmt.Errorf("gemini API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the API response
	var geminiResp GeminiResponse
	if err = json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	// Extract transcript text
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		transcript := geminiResp.Candidates[0].Content.Parts[0].Text
		fmt.Printf("✅ Transcription successful! Length: %d characters\n", len(transcript))
		return transcript, nil
	}

	return "", fmt.Errorf("invalid response: missing text content")
}

// detectMimeType determines the MIME type based on file extension
func detectMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	mimeTypes := map[string]string{
		".mp3":  "audio/mp3",
		".wav":  "audio/wav",
		".webm": "audio/webm",
		".m4a":  "audio/mp4",
		".ogg":  "audio/ogg",
		".flac": "audio/flac",
		".aac":  "audio/aac",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}

	// Default to webm if unknown
	return "audio/webm"
}