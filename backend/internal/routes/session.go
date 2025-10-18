package routes

import (
	"encoding/json"
	"net/http"
	"time"
)

func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getSessions(w, r)
	case "POST":
		createSession(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getSessions returns all stored sessions
func getSessions(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"success": true,
		"data":    allSessions,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// createSession saves a new session after transcription
func createSession(w http.ResponseWriter, r *http.Request) {
	var newSession struct {
		Title  string `json:"title"`
		Text   string `json:"text"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&newSession); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session := Session{
		ID:          time.Now().Format("20060102150405"), // simple unique ID
		Title:       newSession.Title,
		StartTime:   time.Now(),
		Duration:    0,
		Status:      newSession.Status,
		TotalWords:  len(splitWords(newSession.Text)),
		KeyPoints:   extractKeyPoints(newSession.Text),
		ActionItems: extractActionItems(newSession.Text),
		Text:        newSession.Text,
	}

	allSessions = append(allSessions, session)

	resp := map[string]interface{}{
		"success": true,
		"data":    session,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
