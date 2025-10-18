package routes

import (
	"fmt"
	"io"
	"net/http"
	"noted/internal/summarization"
	"os"

	"github.com/gorilla/mux"
)

//rEGISTERrOUTES SETS UP ALL http ENDPOINTS

func RegisterRoutes() *mux.Router {
	router := mux.NewRouter()

	//Route 1: Home endpoint (test route)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to Noted API")

	} ).Methods("GET")

	//Route 2: upload and transcribe audio
	router.HandleFunc("/transcribe", handleTranscribe).Methods("POST")

	router.HandleFunc("/summarize", handleSummarize).Methods("POST")
	router.HandleFunc("/status", sessionsHandler).Methods("GET")
	



	return router
}


func handleSummarize(w http.ResponseWriter, r *http.Request) {
	body, err  := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	output, err := summarization.SummarizeWithGemini(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(output))
}


func handleTranscribe(w http.ResponseWriter, r *http.Request){
	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "failed to read audio file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tempPath := "temp_" + header.Filename
	tempFile, err := os.Create(tempPath)
	if err != nil {
		http.Error(w, "failed to create temp file", http.StatusInternalServerError)
		return
	}

	defer os.Remove(tempPath)
	io.Copy(tempFile, file)
	tempFile.Close()


	output, err := summarization.TranscribeWithGemini(tempPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(output))
}


