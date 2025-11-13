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
						Text: `{
  "Text": "Please provide a concise and well-structured summary of this meeting transcript. The summary should include:\n1. Main topics discussed\n2. Key decisions made\n3. Action items (if any)\n4. Important points or insights raised\n\nEnsure the summary is clear, professionally written, and easy to read. Avoid unnecessary details or filler content — focus only on the essential information.\n\nTranscript:"
}
:
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
	apiKey := config.GetGeminiKey()
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	fmt.Println("📂 Reading audio file:", audioFilePath)
	audioData, err := os.ReadFile(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read audio file: %v", err)
	}

	fmt.Printf("📊 Audio file size: %.2f MB\n", float64(len(audioData))/(1024*1024))

	// Compress if too large
	if len(audioData) > 10*1024*1024 { // > 10MB
		fmt.Println("⚠️  Large file detected. This may take longer...")
	}

	fmt.Println("🔄 Encoding audio to base64...")
	encodedAudio := base64.StdEncoding.EncodeToString(audioData)
	
	fmt.Printf("📦 Encoded size: %.2f MB\n", float64(len(encodedAudio))/(1024*1024))

	mimeType := detectMimeType(audioFilePath)
	fmt.Println("🎵 MIME type:", mimeType)

	requestBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{
  Text: "Please transcribe this audio into clear, well-structured meeting notes. Correct grammar, punctuation, and spelling where necessary. Organize the content logically to reflect the flow of discussion, using clean formatting such as headings or bullet points only when appropriate. Include the actual names of speakers if mentioned, but do not use placeholders like 'Speaker 1' if names are unknown. Capture all key discussion points, decisions, important details, and action items. Maintain a professional and neutral tone suitable for corporate documentation, and remove filler words or irrelevant chatter. If no title is mentioned, assign one that reflects the meeting’s main focus. Return only the finalized, neatly formatted meeting notes without symbols, markdown marks, or commentary.",
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

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=" + apiKey
	
	fmt.Println("🌐 Sending to Gemini API...")
	startTime := time.Now()
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 300 * time.Second} // 5 minutes max
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("⏱️  Gemini response time: %.2f seconds\n", elapsed.Seconds())

	if resp.StatusCode != http.StatusOK {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
			fmt.Println("❌ API Error Response:", prettyJSON.String())
		}
		return "", fmt.Errorf("gemini API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var geminiResp GeminiResponse
	if err = json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		transcript := geminiResp.Candidates[0].Content.Parts[0].Text
		fmt.Printf("✅ Transcription successful! Length: %d characters (took %.1fs)\n", 
			len(transcript), elapsed.Seconds())
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