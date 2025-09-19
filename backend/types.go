package main

import (
	"time"
)



type Note struct {
	ID   			string  	`json:"id"`
	Text 			string  	`json:"text"`
	Timestamp time.Time `json:"timestamp"`
	Category 	string    `json:"category"`
	Tags			[]string	`json:"tags"` 
	SessionID string    `json:"session_id"`
	IsTranscript bool `json:"is_transcript"`
	Confidence  float64  `json:"confidence"`
	Speaker    string  `json:"speaker"`
	WordCount int     `json:"word_count"`
}



type MeetingSession struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	StartTime time.Time `json:"start_time"`
	EndTime  *time.Time `json:"end_time"`
	Status   string   `json:"status"` 

	//TRANSCRIPTION SETTINGS
	Language    string `json:"language"` //Speech recognition languge ("en", "jp")


	//STATISTICS
	Duration   time.Duration `json:"duration"`
	TotalWords   int   `json:"total_words"`
	SegmentCount  int     `json:segment_count"`



	//SUMMARY DATA
	Summary  string  `json:"summary"`
	KeyPoints   []string  `json:key_points"`
	ActionItems  []string `json:action_items"`
	Participants  []string  `json:"participants"`
}




type TranscriptSegment struct {
	ID    string   `json:"id"`
	Text   string    `json:"text"`
	SessionId string  `json:"session_id"`
	Timestamp time.Time  `json:"timestamp"`



	//RECOGNITION  DATA
	Confidence  float64  `json:"confidence"`
	Language  string  `json:"language`
	Speaker   string   `json:"speaker"`

	//PROCESSING FLAGS
	IsProcessed bool  `json:"is_processed"`
	IsSummarized bool `json:"is_summarized"`


}


type AudioConfig struct {
	//AUDIO CAPTURE SETTING
	SampleRate  int  `json:"sample_rate"` //audio quality (usually 1600 Hz for speech)
	Channels    int   `json:"channels"`  //mono (1) or stereo (2)
	BitDepth   int   `json:"bit_depth"`  //audio precision (usually 16)

	//PROCESSING SETTING
	ChunkDuration   time.Duration   `json:"chunkDuration"`
	Language        string          `json:"Language"`
	EnableNoiseSuppression bool     `json:"noise_suppression"`

	//TRANSCRIPTION SETTINGS
	MinConfidence   float64  `json:"min_confidence"`
	MaxSegmentLength  int   `json:"max_segment_length"`

}




type TranscriptionStatus  struct {
	IsRecording  bool   `json:"is_recording"`
	IsProcessing bool   `json:"is_processing"`
	CurrentSession  string `json:"current-session"`
	LastUpdate     time.Time  `json:"last_update"`


	//REAL_TIME STATS
	WordsPerMinute    float64  `json:"words_per_minute"`
	AudioLevel   float64  `json:"audio_level"`
	QueuedSegments  int `json:"queued_segments"`
	ProcessedSegments  int `json:"processed_segments"`
}






//  My jsondb at the moment
type NotesDB struct {
	Notes  		[]Note  `json:"notes"`
	Sessions  []MeetingSession  `json:"sessions"`
	Segments  []TranscriptSegment `json:"segments"`
	Config    AudioConfig          `json:"config"`
	LastTranscription  time.Time  `json:"last_transcription"`
}


type ExportFormat string


const (
	FormatTXT      ExportFormat = "txt"
	FormatMarkdown ExportFormat = "md"
)