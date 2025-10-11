package main

import (
	"encoding/json"
	f "fmt"
	"io"
	l "log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	vosk "github.com/alphacep/vosk-api/go"
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
	RecentSessions  []MeetingSession  `json:"recent_sessions"`
	TranscriptionStatus TranscriptionStatus  `json:"transcription_status"`
	SystemStats       SystemStats   `json:"system_stats"`
}




//SYSTEM STATISTICS
type SystemStats struct {
	TotalSessions  int   `json:"total_sessions"`
	TotalNotes   int   `json:"total_notes" `
	TotalWords int   `json:"total_words"`
	AverageWordsPerSession  float64 `json:"avg_words_per_session"`
	LastActivity time.Time  `json:"last_activity"`
}




// START WEB SERVER
func StartWebServer() error {
	l.Printf("🌐 Starting MeetNote web server on port %s", serverPort)

	// Create new ServeMux instead of using DefaultServeMux
	mux := http.NewServeMux()

	// API ROUTES
	mux.HandleFunc("/api/status", handleAPIStatus)
	mux.HandleFunc("/api/sessions", handleAPISessions)
	mux.HandleFunc("/api/session/", handleAPISession)
	mux.HandleFunc("/api/transcription/start", handleAPIStart)
	mux.HandleFunc("/api/transcription/stop", handleAPIStop)
	mux.HandleFunc("/api/notes", handleAPINotes)
	mux.HandleFunc("/api/export", handleAPIExport)
	mux.HandleFunc("/api/download/", handleAPIDownload)
	mux.HandleFunc("/api/transcribe", transcribeHandler)

	// REAL-TIME ROUTE
	mux.HandleFunc("/api/live", handleAPILive)

	// STATIC FILE ROUTES
	mux.HandleFunc("/", handleStaticFiles)

	// Wrap mux with CORS middleware
	handler := corsMiddleware(mux)

	// Start server
	address := ":" + serverPort
	l.Printf("✅ Web dashboard available at: http://localhost:%s", serverPort)
	l.Printf("🛜 API endpoints available at http://localhost:%s/api/", serverPort)

	return http.ListenAndServe(address, handler)
}



//API ENDPOINTS

// Transcribe handler
func transcribeHandler(w http.ResponseWriter, r *http.Request) {
    // Parse uploaded file
    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Failed to read audio file", http.StatusBadRequest)
        return
    }
    defer file.Close()
		f.Printf("🎧 Received upload: %s (%d bytes)\n", header.Filename, header.Size)

    // Create temp files
    inputFile, err := os.CreateTemp("", "input-*.webm")
    if err != nil {
        http.Error(w, "Failed to create temp input file", http.StatusInternalServerError)
        return
    }
    defer os.Remove(inputFile.Name())

    _, err = io.Copy(inputFile, file)
    inputFile.Close()
    if err != nil {
        http.Error(w, "Failed to save uploaded audio", http.StatusInternalServerError)
        return
    }
		f.Printf("✅ Saved uploaded file to: %s\n", inputFile.Name())

    // Convert WebM → WAV (mono, 16kHz)
    outputFile, err := os.CreateTemp("", "output-*.wav")
    if err != nil {
        http.Error(w, "Failed to create temp output file", http.StatusInternalServerError)
        return
    }
    defer os.Remove(outputFile.Name())
    outputFile.Close()
		

   cmd := exec.Command(
    "ffmpeg",
    "-y",
    "-i", inputFile.Name(),
    "-ar", "16000",
    "-ac", "1",
    "-f", "wav",
    outputFile.Name(),
)
output, err := cmd.CombinedOutput()
if err != nil {
    f.Printf("ffmpeg error: %s\n", string(output))
    sendAPIError(w, f.Sprintf("ffmpeg failed: %v - %s", err, string(output)), http.StatusInternalServerError)
    return
}
 f.Printf("✅ ffmpeg successfully converted to: %s\n", outputFile.Name())

    // Read the converted WAV file
    wavData, err := os.ReadFile(outputFile.Name())
    if err != nil {
        http.Error(w, "Failed to read converted audio", http.StatusInternalServerError)
        return
    }

    // Feed it to Vosk
    rec, _ := vosk.NewRecognizer(voskModel, 16000.0)
    defer rec.Free()

    rec.AcceptWaveform(wavData)
    result := rec.FinalResult()
    f.Printf("📝 Transcription done for %s\n", header.Filename)
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(result))

		f.Printf("ffmpeg error: %s\n", string(output))
}


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
func handleAPISessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// ✅ Existing code to list sessions
		db, err := LoadNotesDB()
		if err != nil {
			sendAPIError(w, "Failed to load sessions", http.StatusInternalServerError)
			return
		}

		// Reverse order (newest first)
		sessions := make([]MeetingSession, len(db.Sessions))
		for i, j := 0, len(db.Sessions)-1; i <= j; i, j = i+1, j-1 {
			sessions[i], sessions[j] = db.Sessions[j], db.Sessions[i]
		}
		sendAPISuccess(w, "Sessions retrieved successfully", sessions)

	case "POST":
		// ✅ New code to create a new session
		var req struct {
			Title string `json:"title"`
			Text  string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendAPIError(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		db, err := LoadNotesDB()
		if err != nil {
			sendAPIError(w, "Failed to load DB", http.StatusInternalServerError)
			return
		}

		newSession := MeetingSession{
			ID:          generateSessionID(), // you can use uuid.NewString() if available
			Title:       req.Title,
			StartTime:   time.Now(),
			Duration:    0,
			Text:       req.Text,
			Status:      "completed",
			TotalWords:  len(strings.Fields(req.Text)),
			KeyPoints:   []string{}, // can be processed later
			ActionItems: []string{},
		}

		// Append to DB and save
		db.Sessions = append(db.Sessions, newSession)
		if err := SaveNotesDB(db); err != nil {
			sendAPIError(w, "Failed to save session", http.StatusInternalServerError)
			return
		}

		sendAPISuccess(w, "Session saved successfully", newSession)

	default:
		sendAPIError(w, "Method not allowed. Use GET or POST.", http.StatusMethodNotAllowed)
	}
}




// GET SPECIFIC SESSION DETAILS
func handleAPISession(w http.ResponseWriter, r *http.Request){
	if r.Method != "GET"{
		sendAPIError(w, "Method not allowed, Please use GET Method", http.StatusMethodNotAllowed)
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
			l.Printf("DEBUG Session found: %+v\n", session)
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
	"id":       session.ID,
	"title":    session.Title,
	"text":     session.Text, 
	"notes":    sessionNotes,
	"stats": map[string]interface{}{
		"total_notes":       len(sessionNotes),
		"transcript_notes":  countTranscriptNotes(sessionNotes),
		"manual_notes":      len(sessionNotes) - countTranscriptNotes(sessionNotes),
	},
}

sendAPISuccess(w, "Session details retrieved successfully", sessionDetails)
}

//STARAT TRANSCRIPTION VIA API
func handleAPIStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendAPIError(w, "Method not allowed, Please use POST Method", http.StatusMethodNotAllowed)
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
		sendAPIError(w, "Method not allowed, Please use POST Method", http.StatusMethodNotAllowed)
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
		sendAPIError(w, "Method not allowed,Please use GET Method", http.StatusMethodNotAllowed)
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
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit < len(notes){
			notes = notes[len(notes)-limit:] // gET MOST RECENT NOTES
		}
	}

	sendAPISuccess(w, "Notes retrieved successfully", notes)
}

//EXPORT SESSION VIA API

func handleAPIExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendAPIError(w, "Method not allowed, Please use POST Method", http.StatusMethodNotAllowed)
	}



	var request struct {
		SessionID string  `json:"session_id"`
		Format  ExportFormat `json:"format"`
		Filename string   `json:"filename"`
	}


	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendAPIError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	//Validate request
	if request.SessionID == "" || request.Format == "" || request.Filename == ""{
		sendAPIError(w, "Session ID, format, and filename are required", http.StatusBadRequest)
		return
	}

	//Export session
	err = exportSessionSummary(request.SessionID, request.Format, request.Filename)
	if err != nil {
		sendAPIError(w, f.Sprintf("Export failed: %v", err), http.StatusInternalServerError)
		return
	}

	sendAPISuccess(w, "Session exported successfully", map[string]string{
		"filename": request.Filename,
	"format": string(request.Format),
	})

}

//LIVE UPDATES ENDPOINT
//This provides real-time updates for the web dashboard
func handleAPILive(w http.ResponseWriter, r *http.Request){
	if r.Method != "GET" {
		sendAPIError(w, "Method not allowed, Please use GET Method", http.StatusMethodNotAllowed)
		return
	}


//This is a simplified live update endpoint
//In a production system, you'd use WebSockets or Server-sent events

liveData := map[string]interface{}{
	"timestamp":    time.Now(),
	"is_recording":  isRecording,
	"currrent_session":  currentSession,
	"audio_level":  0.65,
	"words_per_minute":   calculateCurrentWPM(),
	"queued_segments":  len(audioChannel),
	"processed_segments": 0,
}

sendAPISuccess(w, "Live data retrieved", liveData)


}


// DOWNLOAD ENDPOINT - Serves exported files for download
func handleAPIDownload(w http.ResponseWriter, r *http.Request) {
    // Extract filename from URL
    parts := strings.Split(r.URL.Path, "/")
    filename := parts[len(parts)-1]
    
    // Security: prevent directory traversal
    if strings.Contains(filename, "..") {
        http.Error(w, "Invalid filename", http.StatusBadRequest)
        return
    }
    
    // Read file
    data, err := os.ReadFile(filename)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }
    
    // Set download headers
    w.Header().Set("Content-Disposition", f.Sprintf("attachment; filename=%s", filename))
    w.Header().Set("Content-Type", "application/octet-stream")
    w.Write(data)
}

//STATIC FILE HANDLER
// sERVES THE HTML,CSS AND JAVASCRIPT FILES FOR WEB DASHBOARD
func handleStaticFiles(w http.ResponseWriter, r *http.Request){
	// If requesting root path, serve index.html
	if r.URL.Path == "/"{
		http.ServeFile(w, r, staticDir+"index.html")
		return
	}


	//For other paths, serve files from static directory
	filepath := staticDir + strings.TrimPrefix(r.URL.Path, "/")

	//Security check:prevent directory traversal
	if strings.Contains(filepath, ".."){
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}


	//Serve the file
	http.ServeFile(w, r, filepath)
}



//CORS MIDDLEWARE
//Allows web browsers to access our API from any origin
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				f.Println("✅ CORS middleware applied for:", r.URL.Path)


        // Handle preflight requests
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        // Continue to actual handler
        next.ServeHTTP(w, r)
    })
}



//HELPER FUNCTIONS

//Send successful API response
func sendAPISuccess(w http.ResponseWriter, message string, data interface{}){
	w.Header().Set("Content-Type", "application/json")
	response := APIResponse{
		Success: true,
		Message: message,
		Data: data,
		Timestamp: time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}


//Send error API response
func sendAPIError(w http.ResponseWriter, message string, statusCode int){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success:  false,
		Message: message,
		Timestamp: time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

//Calculate total words across all sessions
func calculateTotalWords(sessions []MeetingSession) int {
	total := 0
	for _, session := range sessions {
		total += session.TotalWords
	}
	return total
}



//GET RECENT SESSIONS
func getRecentSessions(session []MeetingSession, limit int) []MeetingSession {
	if len(session) <= limit {
		return session
	}
	return session[len(session)-limit:]
}

//Count transcript notes vs manual notes
func countTranscriptNotes(notes []Note) int {
	count := 0 
	for _, note := range notes {
		if note.IsTranscript{
			count++
		}
	}
	return count
}


//Calculate current words per minute
func calculateCurrentWPM() float64 {
	if currentSession == nil {
		return 0.0
	}

	duration := time.Since(currentSession.StartTime).Minutes()
	if duration <= 0 {
		return 0.0
	}
	return float64(currentSession.TotalWords) / duration
}