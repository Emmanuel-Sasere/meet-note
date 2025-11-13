package summarization

import (
	"fmt"
	"strings"
)

// TranscribeWithFallback tries multiple APIs until one works
func TranscribeWithFallback(audioFilePath string) (string, error) {
	var lastError error

	// Try Groq first (fastest + free)
	fmt.Println("🔄 Trying Groq API...")
	transcript, err := TranscribeWithGroq(audioFilePath)
	if err == nil {
		return transcript, nil
	}
	if !strings.Contains(err.Error(), "RATE_LIMIT") && !strings.Contains(err.Error(), "not set") {
		return "", err // Real error, not rate limit
	}
	lastError = err
	fmt.Printf("⚠️  Groq failed: %v\n", err)

	// Try Gemini second
	fmt.Println("🔄 Trying Gemini API...")
	transcript, err = TranscribeWithGemini(audioFilePath)
	if err == nil {
		return transcript, nil
	}
	if !strings.Contains(err.Error(), "limit") && !strings.Contains(err.Error(), "not set") {
		return "", err
	}
	lastError = err
	fmt.Printf("⚠️  Gemini failed: %v\n", err)

	// Add more APIs here...

	return "", fmt.Errorf("all transcription services failed. Last error: %v", lastError)
}

// SummarizeWithFallback tries multiple APIs for summarization
func SummarizeWithFallback(text string) (string, error) {
	// Try Groq first (same API that transcribed)
	fmt.Println("🔄 Trying Groq for summary...")
	summary, err := SummarizeWithGroq(text)
	if err == nil {
		return summary, nil
	}
	fmt.Printf("⚠️ Groq summary failed: %v\n", err)

	// Fallback to Gemini
	fmt.Println("🔄 Trying Gemini for summary...")
	summary, err = SummarizeWithGemini(text)
	if err == nil {
		return summary, nil
	}
	fmt.Printf("⚠️ Gemini summary failed: %v\n", err)

	return "", fmt.Errorf("all summarization services failed")
}

