package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"noted/internal/summarization"
	"os"

	"github.com/gorilla/mux"
)

// Response structure for API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RegisterRoutes sets up all HTTP endpoints
func RegisterRoutes() *mux.Router {
	router := mux.NewRouter()

	// Enable CORS for frontend
	router.Use(corsMiddleware)

	// Route 1: Health check
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to Noted API")
	}).Methods("GET")

	// Route 2: Upload and transcribe audio (handles both recording and file upload)
	router.HandleFunc("/transcribe", handleTranscribe).Methods("POST")

	return router
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper function to send JSON responses
func sendJSON(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Handle transcription and summarization
func handleTranscribe(w http.ResponseWriter, r *http.Request) {
	// Parse the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Failed to read file from request",
		})
		return
	}
	defer file.Close()

	// Get file type (audio or video)
	fileType := r.FormValue("type")
	if fileType == "" {
		fileType = "audio" // default
	}

	fmt.Printf("📁 Processing %s file: %s (size: %d bytes)\n", fileType, header.Filename, header.Size)

	// Create temp file in system temp directory
	var tempFile *os.File
	if fileType == "video" {
		tempFile, err = os.CreateTemp("", "video_*.mp4")
	} else {
		tempFile, err = os.CreateTemp("", "audio_*.webm")
	}
	
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create temp file: %v", err),
		})
		return
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // Clean up temp file

	// Copy uploaded file to temp location
	_, err = io.Copy(tempFile, file)
	tempFile.Close()
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to save uploaded file: %v", err),
		})
		return
	}

	fmt.Printf("✅ File saved: %s\n", tempPath)

	// If video, extract audio first
	var audioPath string
	if fileType == "video" {
		fmt.Println("🎬 Extracting audio from video...")
		audioPath, err = summarization.ExtractAudioFromVideo(tempPath)
		if err != nil {
			sendJSON(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to extract audio from video: %v", err),
			})
			return
		}
		defer os.Remove(audioPath) // Clean up extracted audio
		fmt.Println("✅ Audio extracted successfully")
	} else {
		audioPath = tempPath
	}

	// Step 1: Transcribe audio with Gemini
	fmt.Println("🎤 Transcribing audio...")
	transcript, err := summarization.TranscribeWithGemini(audioPath)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Transcription failed: %v", err),
		})
		return
	}

	// Step 2: Summarize the transcript with Gemini
	fmt.Println("📝 Summarizing transcript...")
	summary, err := summarization.SummarizeWithGemini(transcript)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Summarization failed: %v", err),
		})
		return
	}

	// Return both transcript and summary
	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]string{
			"transcript": transcript,
			"summary":    summary,
		},
	})
}