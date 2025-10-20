package routes


import (
	"encoding/json"
	"net/http"
	"time"
)

// SystemStats represents general app stats
type SystemStats struct {
	TotalSessions       int     `json:"total_sessions"`
	TotalWords          int     `json:"total_words"`
	AvgWordsPerSession  float64 `json:"avg_words_per_session"`
}

// Session represents one meeting session
type Session struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	StartTime   time.Time `json:"start_time"`
	Duration    int       `json:"duration"`
	Status      string    `json:"status"` // "active" or "completed"
	TotalWords  int       `json:"total_words"`
	KeyPoints   []string  `json:"key_points"`
	ActionItems []string  `json:"action_items"`
	Text        string    `json:"text,omitempty"`
}

// Global in-memory (temporary) store for sessions
var allSessions []Session
var currentSession *Session

// getSystemStatus - returns system stats + current session
func getSystemStatus(w http.ResponseWriter, r *http.Request) {
	stats := SystemStats{
		TotalSessions:      len(allSessions),
		TotalWords:         0,
		AvgWordsPerSession: 0,
	}

	// Calculate stats if sessions exist
	if len(allSessions) > 0 {
		totalWords := 0
		for _, s := range allSessions {
			totalWords += s.TotalWords
		}
		stats.TotalWords = totalWords
		stats.AvgWordsPerSession = float64(totalWords) / float64(len(allSessions))
	}

	resp := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"is_recording":   currentSession != nil,
			"current_session": currentSession,
			"system_stats":   stats,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
