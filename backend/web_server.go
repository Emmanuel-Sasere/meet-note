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
//Return current transcription status and recent activity
func handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	//only allow GET request
	if r.Method != "GET" {
		sendAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}


	//Load database to get system statictics
	db, err := LoadNotesDB()
	if err != nil {
		sendAPIError(w, " Failed to load system data", http.StatusInternalServerError)
		return
	}

	//Calculate system statistics
	stats := SystemStats{
		TotalSessions: len(db.Sessions),
		TotalNotes:    len(db.Notes),
		TotalWords:    calculateTotalWords(db.Sessions),
		LastActivity: db.LastTranscription,
	}

	if stats.TotalSessions > 0 {
		stats.AverageWordsPerSession = float64(stats.TotalWords) / float64(stats.TotalSessions)
	}

	//Get recent session (last 5)
	recentSessions := getRecentSessions(db.Sessions, 5)


	//Build transcription status
	transcriptionStatus := TranscriptionStatus {
		IsRecording:    isRecording,
		IsProcessing:   isProcessing,
		CurrentSession:  "",
		LastUpdate:      time.Now(),
		WordsPerMinute:  calculateCurrentWPM(),
		AudioLevel:    0.75,
		QueuedSegments:  len(audioChannel),
		ProcessedSegments: 0,  
	}

	if currentSession != nil {
		transcriptionStatus.CurrentSession = currentSession.ID
	}

	//BUILD COMPLETE STATUS RESPONSE
	status := WebStatus{
		IsRecording:  isRecording,
		CurrentSession:    currentSession,
		RecentSessions:    recentSessions,
		TranscriptionStatus:   transcriptionStatus,
		SystemStats:       stats,
	}

	sendAPISuccess(w, "Status retrieved successfully", status)
}


//GET ALL SESSIONS
func handleAPISessions(w http.ResponseWriter, r *http.Request)  {
	if r.Methid != "GET" {
		sendAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	db, err := LoadNotesDB()
	if err != nil {
		sendAPIError(w, "Failed to load sessions", http.StatusInternalServerError)
		return
	}


	//Return sessions in reverse chronological order (newest first)
	sessions := make([]MeetingSession, len(db.Sessions))
	for i,j := 0, len(db.Sessions)-1; i <= j; i,j = i+1, j-1 {
		sessions[i], sessions[j] = db.Sessions[j], db.Sessions[i]
	}
	sendAPISuccess(w, "Sessions retrieved successfully", sessions)
}

//GET SPECIFIC SESSION DETAILS
func handleAPISession(w http.RespoonseWriter, r *http.Request) {
	if r.Method != "GET"{
		sendAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := LoadNotesDB()
	if err != nil {
		sendAPIError(w, "Failed to load sessions", http.StatusInternalServerError)
		return
	}

	

	//RETURN SESSIONS IN REVERSE CHRONOLOGICAL ORDER (newest first)
	sessions := make([]MeetingSession, len(db.Sessions))
	for i,j := 0, len(db.Sessions)-1; i <= j; i, j = i+1, j-1{
		sessions[i], sessions[j] = db.Sessions[j], db.Sessions[i]
	} 

	sendAPISuccess(w, "Sesssion retrieved successfully", sessions)
}


// GET SPECIFIC SESSION DETAILS
func handleAPISession(w http.ResponseWriter, r *http.Request){
	if r.Method != "GET"{
		sendAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}


	//EXTRACT SESSION ID FROM URL PATH

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		sendAPIError(w, "Session ID required", http.StatusBadRequest)
		return
	}
	sessionID := pathParts[3]


	//Find session in database
	db, err := LoadNotesDB()
	if err != nil {
		sendAPIError(w, "Failed to load sessions", http.StatusInternalServerError)
		return
	}

	var session *MeetingSession
	for _, s := range db.Sessions{
		if s.ID == sessionID {
			session = &s
			break
		}
	}

	if session == nil {
		sendAPIError(w, "Session not found", http.StatusNotFound)
		return
	}

	//Get session notes
	sessionNotes, err := getNotesBySessionID(sessionID)
	if err != nil {
		l.Printf("Warning: Failed to get session notes: %v", err)
		sessionNotes = []Note{}
	}

	//Build detailed session response
	sessionDetails := map[string]interface{}{
		"session": session,
		"notes": sessionNotes,
		"stats": map[string]interface{}{
			"total_notes":  len(sessionNotes),
			"transcript_notes": countTranscriptNotes(sessionNotes),
			"manual_notes":   len(sessionNotes) - countTranscriptNotes(sessionNotes),
		},
	
	}
	sendAPISuccess(w, "Session details retrieved successfully", sessionDetails)
}

//STARAT TRANSCRIPTION VIA API
func handleAPIStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Parse JSON request body
	var request struct {
		Title string `json:"title"`
		Language string `json:"language"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendAPIError(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	//validate title
	if request.Title == ""{
		request.Title = f.Sprintf("Meeting %s", time.Now().Format("2006-01-02 15:04"))
	}

	//Start transcription session
	session, err := StartTranscriptionSession(request.Title)
	if err != nil{
	sendAPIError(w, f.Sprintf("Failed to start transcription: %v", err), http.StatusInternalServerError)
	return
}

l.Printf("🎙️ Web client started transcription: %s", request.Title)
sendAPISuccess(w, "Transcription started successfully", session)
} 

//STOP TRANSCRIPTION VIA API
func handleAPIStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := StopTranscriptionSession()
	if err != nil {
		sendAPIError(w, f.Sprintf("Failed to stop transcription: %v", err), http.StatusInternalServerError)
		return
	}
	l.Printf("⏹️ Web client stopped transcription")
	sendAPISuccess(w, "Transcription stopped successfully", nil)
}


//GET NOTES
func handleAPINotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET"{
		sendAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Parse query parameters
	sessionID := r.URL.Query().Get("session")
	limitStr := r.URL.Query().Get("limit")


	var notes []Note
	var err error

	if sessionID != ""{
		//Get notes for specific session
		notes, err = getNotesBySessionID(sessionID)
	}else {
		//Get all notes
		notes, err = GetAllNotes()
	}

	if err != nil {
		sendAPIError(w, "Failed to load notes", http.StatusInternalServerError)
		return
	}


	//APPLY LIMIT IF SPECIFIED
	if  limitStr != ""{
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && linit < len(notes){
			notes = notes[len(notes)-limit:] // gET MOST RECENT NOTES
		}
	}

	sendAPISuccess(w, "Notes retrieved successfully", notes)
}

//EXPORT SESSION VIA API
	