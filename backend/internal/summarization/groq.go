package summarization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"noted/internal/config"
	"os"
	"time"
)

type GroqTranscriptionResponse struct {
	Text string `json:"text"`
}

// TranscribeWithGroq - Step 1: Whisper transcribe, Step 2: Llama refine
func TranscribeWithGroq(audioFilePath string) (string, error) {
	apiKey := config.GetGroqKey()
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY not set")
	}

	// STEP 1: Whisper transcription
	fmt.Println("🚀 Step 1: Transcribing with Whisper...")
	rawTranscript, err := transcribeWithWhisper(apiKey, audioFilePath)
	if err != nil {
		return "", fmt.Errorf("whisper failed: %v", err)
	}
	fmt.Printf("✅ Raw transcript: %d chars\n", len(rawTranscript))

	// STEP 2: Refine with Llama3
	fmt.Println("🧠 Step 2: Refining with Llama3...")
	refinedTranscript, err := refineTranscriptWithLlama(apiKey, rawTranscript)
	if err != nil {
		fmt.Printf("⚠️ Refinement failed: %v. Returning raw transcript.\n", err)
		return rawTranscript, nil // Return raw if refinement fails
	}

	fmt.Printf("✅ Refined transcript: %d chars\n", len(refinedTranscript))
	return refinedTranscript, nil
}

// Step 1: Whisper transcription (raw audio → text)
func transcribeWithWhisper(apiKey, audioFilePath string) (string, error) {
	file, err := os.Open(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", audioFilePath)
	if err != nil {
		return "", err
	}
	io.Copy(part, file)

	writer.WriteField("model", "whisper-large-v3")
	writer.WriteField("language", "en")
	writer.Close()

	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/audio/transcriptions", body)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 429 {
		return "", fmt.Errorf("RATE_LIMIT")
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("whisper error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result GroqTranscriptionResponse
	json.Unmarshal(bodyBytes, &result)
	return result.Text, nil
}

// Step 2: Llama3 refinement (raw text → clean formatted text)
func refineTranscriptWithLlama(apiKey, rawTranscript string) (string, error) {
	payload := map[string]interface{}{
		"model": "llama-3.3-70b-versatile",
		"messages": []map[string]interface{}{
	{
  "role": "system",
  "content": "You are a professional corporate meeting transcript editor and recorder. Your task is to transform a raw meeting transcript into a clear, well-structured, and polished meeting document.\n\nInstructions:\n1. Correct all grammar, punctuation, and spelling errors.\n2. Organize the discussion into logical sections and paragraphs that follow the natural flow of the meeting.\n3. Apply clean, professional formatting — use headings, subheadings, and bullet points only where appropriate.\n4. Use the actual names of speakers if they are mentioned. If a speaker’s name is not available, do not insert placeholders like 'Speaker 1' — simply write their statements naturally.\n5. Include meeting details (e.g., title, date, time, participants, agenda) only if they appear in the transcript. If they are not provided, omit them.\n6. Maintain a professional, neutral tone suitable for corporate documentation.\n7. Preserve all meaningful details — do not summarize, shorten, or omit important information.\n8. Highlight key points, decisions, action items, and follow-ups if they are mentioned.\n9. Remove filler words, repeated phrases, or irrelevant chatter to ensure readability.\n10. If no clear title is provided, you may assign one based on the meeting’s content.\n\nDeliverable:\nReturn ONLY the finalized, neatly formatted meeting document — no symbols, markdown marks (#, *), or extra commentary.",
},


			{
				"role":    "user",
				"content": rawTranscript,
			},
		},
		"temperature": 0.3,
		"max_tokens":  8000,
	}

	jsonBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 429 {
		return "", fmt.Errorf("RATE_LIMIT")
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("llama error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	json.Unmarshal(bodyBytes, &result)

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no refined text returned")
	}

	return result.Choices[0].Message.Content, nil
}

// SummarizeWithGroq - Uses the REFINED transcript
func SummarizeWithGroq(refinedTranscript string) (string, error) {
	apiKey := config.GetGroqKey()
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY not set")
	}

	fmt.Println("📝 Summarizing refined transcript with Groq...")

	payload := map[string]interface{}{
		"model": "llama-3.3-70b-versatile",
		"messages": []map[string]interface{}{
			{
  "role": "system",
  "content": "You are a professional meeting summarizer. Your task is to create a clear and concise summary of the meeting transcript by including:\n\n1. **Main Topics Discussed** – capture the core subjects and themes of the conversation.\n2. **Key Decisions Made** – list any resolutions, agreements, or conclusions reached.\n3. **Action Items** – specify tasks assigned, responsible persons (if mentioned), and deadlines (if given).\n4. **Important Points** – highlight notable insights, concerns, or follow-ups.\n\nGuidelines:\n- Keep the summary concise but complete — do not omit critical information.\n- Use bullet points or short paragraphs for readability.\n- Maintain a professional tone suitable for official documentation.\n\nReturn ONLY the formatted summary — no explanations or commentary.",
},

			{
				"role":    "user",
				"content": "Summarize this transcript:\n\n" + refinedTranscript,
			},
		},
		"temperature": 0.5,
		"max_tokens":  1500,
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 429 {
		return "", fmt.Errorf("RATE_LIMIT")
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("groq error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	json.Unmarshal(bodyBytes, &result)

	if len(result.Choices) > 0 {
		summary := result.Choices[0].Message.Content
		fmt.Printf("✅ Summary complete: %d chars\n", len(summary))
		return summary, nil
	}

	return "", fmt.Errorf("no summary generated")
}