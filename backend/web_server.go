package main

import (
	"encoding/json"
	f "fmt"
	l "log"
	"net/http"
	"strconv"
	"strings"
	"time"
)




const (
	serverPort = "8080"
	staticDir = "web/"
)


type APIResponse struct {
	Success   bool   `json:"success"`
	Message  string  `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}



//REAL-TIME STATUS FOR WEB DASHBOARD
type WebStatus struct {
	IsRecording   bool     `json:"is_recording"`
	CurrentSession *MeetingSession   `json:"current_session"`
	RecentSession  []MeetingSession  `json:"recent_sessions"`
	TranscriptionStatus TranscriptionStatus  `json:"transcription_status"`
	SystemStats       SystemStats   `json:""system_stats`
}




//SYSTEM STATISTICS
type SystemStats struct {
	TotalSessions  int   `json:"total_sessions"`
	TotalNotes   int   `json:"total_notes" `
	TotalWords int   `json:"total_words"`
	AverageWordsPerSession  float64 `json:"avg_words_per_session"`
	LastActivity time.Time  `json:"last_activity"`
}



//STAT WEB SERVER
func StartWebServer() error {
	l.Printf("🌐 Starting MeetNote web server on port %s", serverPort)

	//API ROUTES (RETURN JSON DATA)
	http.HandleFunc("/api/status", handleAPIStatus)
	http.HandleFunc("/api/sessions", handleAPISessions)
	http.HandleFunc("/api/session/", handleAPISession)
	http.HandleFunc("/api/transcription/start", handleAPIStart)
	http.HandleFunc("/api/transcription/stop", handleAPIStop)
	http.HandleFunc("/api/notes", handleAPINotes)
	http.HandleFunc("/api/export", handleAPIExport)


	//REAL-TIME ROUTE
	http.HandleFunc("/api/live", handleAPILive)

	//STACTIC FILE ROUTES (serve NEXTJS FILES)
	http.HandleFunc("/", handleStaticFiles)

	//Enable CORS for all routes (allows browser access)
	http.HandleFunc("/", corsMiddleware(http.DefaultServeMux))



	//START THE SERVER
	address := ":" + serverPort
	l.Printf("✅ Web dashboard available at: http://localhost:%s", serverPort)
	l.Printf("🛜 API endpoints available at http://localhost:%s/api/", serverPort)

	return http.ListenAndServer(address, nil)
}


//API ENDPOINTS
//GET SYSTEM STATUS