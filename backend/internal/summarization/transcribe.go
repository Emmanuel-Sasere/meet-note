package summarization

import (
	"bytes"
	"errors"
	"fmt"
	
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const CHUNK_DURATION = 600 // 10 minutes in seconds

// -----------------------------
// Public convenience entrypoint
// -----------------------------
func TranscribeAndSummarize(audioFilePath string) (string, string, error) {
	// 1) Transcribe (chunked if needed)
	transcript, err := TranscribeWithFallback(audioFilePath)
	if err != nil {
		return "", "", fmt.Errorf("transcription failed: %w", err)
	}

	// 2) Summarize
	summary, err := SummarizeWithFallback(transcript)
	if err != nil {
		return transcript, "", fmt.Errorf("summarization failed: %w", err)
	}

	return transcript, summary, nil
}

// -----------------------------
// Transcription (with chunking)
// -----------------------------
func TranscribeWithFallback(audioFilePath string) (string, error) {
	fmt.Println("🔎 Determining audio duration...")
	duration, err := getAudioDuration(audioFilePath)
	if err != nil || duration <= 0 {
		fmt.Printf("⚠️ Could not get duration (%v) — processing entire file as single file\n", err)
		return transcribeSingleFile(audioFilePath)
	}

	fmt.Printf("🎵 Duration: %.0f sec (%.1f min)\n", duration, duration/60.0)

	if duration <= CHUNK_DURATION {
		return transcribeSingleFile(audioFilePath)
	}

	return transcribeInChunks(audioFilePath, duration)
}

func transcribeSingleFile(path string) (string, error) {
	fmt.Println("🎤 Transcribing single file — Groq first")
	txt, err := TranscribeWithGroq(path)
	if err == nil {
		return txt, nil
	}

	// If Groq error indicates rate limit/quota/missing key, fallback to Gemini
	if isGroqLimitError(err) {
		fmt.Printf("⚠️ Groq error indicates limit/quota/unauthorized: %v — falling back to Gemini\n", err)
	} else {
		// Non-fallbackable error — return directly
		return "", fmt.Errorf("groq transcription failed: %w", err)
	}

	txt, err = TranscribeWithGemini(path)
	if err == nil {
		return txt, nil
	}
	return "", fmt.Errorf("both groq and gemini transcription failed: %w", err)
}

func transcribeInChunks(audioPath string, totalDuration float64) (string, error) {
	numChunks := int(totalDuration/CHUNK_DURATION) // zero-based floor
	if totalDuration/CHUNK_DURATION > float64(numChunks) {
		numChunks++ // ceil
	}
	fmt.Printf("⏳ Splitting audio into %d chunks (~%d sec each)\n", numChunks, CHUNK_DURATION)

	tempDir := filepath.Dir(audioPath)
	var allTranscripts []string

	for i := 0; i < numChunks; i++ {
		startSec := i * CHUNK_DURATION
		chunkFile := filepath.Join(tempDir, fmt.Sprintf("chunk_%02d.mp3", i))

		fmt.Printf("\n🔹 Extracting chunk %d/%d (start %ds)...\n", i+1, numChunks, startSec)
		if err := extractAudioChunk(audioPath, chunkFile, startSec, CHUNK_DURATION); err != nil {
			fmt.Printf("⚠️ Failed to extract chunk %d: %v — skipping\n", i+1, err)
			continue
		}
		// Ensure we remove chunk after processing
		defer func(p string) {
			_ = os.Remove(p)
		}(chunkFile)

		// Transcribe chunk using Groq first and auto-fallback to Gemini
		fmt.Printf("🔤 Transcribing chunk %d with Groq (fallback to Gemini if needed)...\n", i+1)
		transcript, err := TranscribeWithGroq(chunkFile)
		if err != nil {
			if isGroqLimitError(err) {
				fmt.Printf("⚠️ Groq limit on chunk %d: %v — switching to Gemini\n", i+1, err)
				transcript, err = TranscribeWithGemini(chunkFile)
			}
		}

		if err != nil {
			fmt.Printf("❌ Failed to transcribe chunk %d with both providers: %v — skipping chunk\n", i+1, err)
			continue
		}

		fmt.Printf("✅ Chunk %d transcribed (%d chars)\n", i+1, len(transcript))
		allTranscripts = append(allTranscripts, transcript)
	}

	if len(allTranscripts) == 0 {
		return "", errors.New("no chunks were successfully transcribed")
	}

	combined := strings.Join(allTranscripts, "\n\n---\n\n")
	fmt.Printf("\n📦 Combined transcript length: %d chars (from %d chunks)\n", len(combined), len(allTranscripts))
	return combined, nil
}

// -----------------------------
// Summarization with fallback
// -----------------------------
func SummarizeWithFallback(text string) (string, error) {
	fmt.Println("📝 Summarize: trying Groq first...")
	sum, err := SummarizeWithGroq(text)
	if err == nil {
		return sum, nil
	}

	// If Groq indicates limit/quota, fallback to Gemini
	if isGroqLimitError(err) {
		fmt.Printf("⚠️ Groq summarization limit/quota/unauthorized: %v — switching to Gemini\n", err)
	} else {
		return "", fmt.Errorf("groq summarization failed: %w", err)
	}

	sum, err = SummarizeWithGemini(text)
	if err == nil {
		return sum, nil
	}
	return "", fmt.Errorf("both groq and gemini summarization failed: %w", err)
}

// -----------------------------
// Utilities: ffprobe / ffmpeg
// -----------------------------
func getAudioDuration(audioPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioPath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If ffprobe isn't available or fails, return error with stderr
		return 0, fmt.Errorf("ffprobe error: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}

	raw := strings.TrimSpace(stdout.String())
	var dur float64
	if _, err := fmt.Sscanf(raw, "%f", &dur); err != nil {
		return 0, fmt.Errorf("unable to parse duration from ffprobe output (%q): %w", raw, err)
	}
	return dur, nil
}

func extractAudioChunk(inputPath, outputPath string, startSec, durationSec int) error {
	// Use fast seek (-ss before -i) for better speed
	cmd := exec.Command("ffmpeg",
		"-ss", fmt.Sprintf("%d", startSec),
		"-t", fmt.Sprintf("%d", durationSec),
		"-i", inputPath,
		"-ac", "1",
		"-ar", "16000",
		"-acodec", "libmp3lame",
		"-y",
		outputPath,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %w output: %s", err, string(out))
	}
	return nil
}

// -----------------------------
// Error classification
// -----------------------------
func isGroqLimitError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())

	// common patterns indicating rate limit, quota, missing key, unauthorized, etc.
	return strings.Contains(msg, "rate") ||
		strings.Contains(msg, "limit") ||
		strings.Contains(msg, "quota") ||
		strings.Contains(msg, "unauthorized") ||
		strings.Contains(msg, "missing") ||
		strings.Contains(msg, "not set") ||
		strings.Contains(msg, "rate_limit") ||
		strings.Contains(msg, "rate-limit")
}

// -----------------------------
// NOTE: Your existing provider functions are used directly below.
// Ensure these functions exist in the same package (signatures must match).
// -----------------------------

// func TranscribeWithGroq(audioFilePath string) (string, error) { ... }
// func TranscribeWithGemini(audioFilePath string) (string, error) { ... }
// func SummarizeWithGroq(text string) (string, error) { ... }
// func SummarizeWithGemini(text string) (string, error) { ... }
