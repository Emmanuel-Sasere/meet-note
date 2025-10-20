package routes

import (
	"strings"
	"noted/internal/summarization"
	"fmt"
)

// splitWords counts words in the transcription
func splitWords(text string) []string {
	return strings.Fields(text)
}

// extractKeyPoints uses Gemini to pull out key ideas
func extractKeyPoints(text string) []string {
	prompt := "Extract 3–5 bullet points summarizing the key ideas from this meeting transcript:\n\n" + text

	summary, err := summarization.SummarizeWithGemini(prompt)
	if err != nil {
		fmt.Println("Error extracting key points:", err)
		return []string{}
	}

	// split into lines (Gemini often returns bullets or numbered points)
	points := strings.Split(summary, "\n")
	var cleaned []string
	for _, line := range points {
		line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}

	return cleaned
}

// extractActionItems uses Gemini to get to-dos or tasks
func extractActionItems(text string) []string {
	prompt := "From this meeting transcript, extract clear action items or next steps:\n\n" + text

	summary, err := summarization.SummarizeWithGemini(prompt)
	if err != nil {
		fmt.Println("Error extracting action items:", err)
		return []string{}
	}

	lines := strings.Split(summary, "\n")
	var items []string
	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
		if line != "" {
			items = append(items, line)
		}
	}

	return items
}
