package main

import (
	f "fmt"
	l "log"
	"time"
	"strings"
)




var (
	//CHANNELS
	audioChannel = make(chan []byte, 100)
	transcriptChan = make(chan TranscriptSegment, 50)

	//STATUS TRACKING
	currentSession  *MeetingSession = nil
	isRecording      bool = false
	isProcessing    bool = false
)



//START TRANSCRIPTION SESSION
func StartTranscriptionSession (title string) (*MeetingSession, error) {
 if currentSession != nil && currentSession.Status == "active" {
	return nil, f.Errorf("A transcription session is already active")
 }


 //Create New meeting session
 session := &MeetingSession{
	ID:    generateSessionID(),
	Title:  title,
	StartTime: time.Now(),
	Status:  "active",
	Language: "en",  //default to English



	//Initialize counters
	TotalWords: 0,
	Duration:  0,
	SegmentCount: 0,


	//Initial arrays
	KeyPoints: []string{},
	ActionItems: []string{},
	Participants: []string{},

 }

	//Set as current session
	currentSession = session


	//Start the background process(goroutines)
	go startAudioCapture()
	go startSpeechProcessor()
	go startNoteProcessor()



	//Save session to database

	err := saveSessionToDB(session)
	if err != nil {
		return nil, f.Errorf("failed to save session: %w", err)
	}

	l.Printf("🎙️ Started transcription session: %s", title)
	return session, nil
}


//Stop Transcripting session
func StopTranscriptionSession() error {
	if currentSession == nil {
		return f.Errorf("no active transcription session")
	}

	//Mark session as completed
	now := time.Now()
	currentSession.EndTime = &now
	currentSession.Status = "completed"
	currentSession.Duration = now.Sub(currentSession.StartTime)

	//Stop recording
	isRecording = false
	isProcessing = false

	//Generate final summary
	err := generateSessionSummary(currentSession)
	if err != nil {
		return f.Errorf("Warning: Failed to generate summary:%v,", err)
	}


	//Save final session state
	err = saveSessionToDB(currentSession)
	if err != nil {
		return f.Errorf("failed to save final session: %w", err)
	}
	l.Printf("⏹️ Stopped transcription session: %s (Duration:%v)", currentSession.Title, currentSession.Duration)

	//Clear current session
	currentSession = nil

	return nil
}


//GOROUTINE 1: AUDIO CAPTURE
//This function help in recording audio continuouly in the background

func startAudioCapture(){
	l.Println("🎤 Starting audio capture...")
	isRecording = true


	//AUDIO CONFIGURATION
	config := AudioConfig{
		SampleRate:   16000,  // 16kHz - good for speech recognition
		Channels:    1,  // Mono auido
		BitDepth:    16,  // 16-bit audio
		ChunkDuration: 2 * time.Second,  //Process audio every 2 second
		Language:   "en",  // English
		MinConfidence: 0.6, // Ignore low-confidence result
	}

	// SIMULATE AUDIO RECORDING
	 for isRecording {
		//generate fake audio, this is suppose to come from my mic, for now it is fake audio
		fakeAudioData := generateSimulateAudio(config.ChunkDuration)

		//Send audio data through the channel to the speech processor
		
		select{
		case audioChannel <- fakeAudioData:
			l.Printf("🛰️ Captured %.1f seconds of audio", config.ChunkDuration.Seconds())
		default: 
			l.Printf("⚠️ Audio buffer full - dropping audio chunk")
		}

		time.Sleep(config.ChunkDuration)
	 }
	l.Println("🛑 Audio capture stopped")
}


//GOROUTINE 2: SPEECH - TO - TEXT PROCESSOR
// This function runs in the background converting audio to text

func startSpeechProcessor(){
	l.Printf("🧠 Starting speech processor...")
	isProcessing = true

	for isProcessing {
		select{
		case audioData := <- audioChannel:
			l.Printf("'🔄 Processing %d bytes of audio data", len(audioData))


			segment := simulateSpeechToText(audioData)

			if segment.Confidence >= 0.6 && len(segment.Text) > 0 {
				select {
				case transcriptChan <- segment:
					l.Printf("✅ Transcribed: \"%s\" (confidence: %.2f)", truncateText(segment.Text, 50), segment.Confidence)
				default:
					l.Println("⚠️ Transcript buffer full")
				}
			}else {
				l.Printf("❌ Low confidence transcript ignored (%.2f)", segment.Confidence)
			}
		case <- time.After(5 * time.Second):
			if !isRecording {
				l.Println("🔇 No more audio to process")
				break
			}

		}
	}

	l.Println("🛑 Speech processor stopped")

}


//GOROUTINE 3: NOTE PROCESSOR
//This function converts transcript segments into notes and saves them

func startNoteProcessor(){
	l.Println("📝 Starting note processor...")

	for{
		select{
		case segment := <- transcriptChan:
			l.Printf("📑 Processing transcript segment: %s", segment.ID)


			note := Note{
				ID:			generateID(),
				Text:		segment.Text,
				Timestamp: segment.Timestamp,
				Category:  currentSession.Title,
				Tags:    []string{"transcript", "auto-generated"},
				SessionID: segment.SessionID,
				IsTranscript: true,
				Confidence: segment.Confidence,
				Speaker: segment.Speaker,
				WordCount:  countWords(segment.Text),
			}

			//save note to database
			err := AddNote(note.Text, note.Category, note.Tags)
			if err != nil {
				l.Printf("❌ Failed to save note: %v", err)
				continue
			}

			//Update session statistics
			if currentSession != nil {
				currentSession.TotalWords += note.WordCount
				currentSession.SegmentCount ++
			}

			//save transcript segment for detailes export
			err = saveSegmentToDB(segment)
			if err != nil {
				l.Printf("⚠️ Failed to save segment: %v", err)
			}

			l.Printf("✅ Saved note: \"%s\" (%d words)", truncateText(note.Text, 40), note.WordCount)
		case <- time.After(10 * time.Second):
			//No transcripts for 10 seconds - check if we should stop

			if !isProcessing {
				l.Println("📝 Note processor finished")
				return
			}
		}
	}
}


//HELPER FUNCTION

//Generate unique session ID
func generateSessionID() string {
	return f.Sprintf("session_%d", time.Now().Unix())
}


// Simulate audio recording (replace ith real audio capture)
func generateSimulateAudio(duration time.Duration) []byte {

	sampleRate := 16000
	sample := int(duration.Seconds() * float64(sampleRate))
	return make([]byte, sample*2)
}



func simulateSpeechToText(audioData []byte) TranscriptSegment {
	fakeTranscripts := []string {
		"Let's start today's meeting",
		"We need to discuss the quarterly results",
		"The project is on track for next month",
		"Can everyone share their updates?",
		"We should schedule a follow-up meeting",
		"Thank you everyone for your time",
	}

	//Pick a random transcript
	transcript := fakeTranscripts[time.Now().Second()%len(fakeTranscripts)]

	segment := TranscriptSegment {
		ID: 		generateID(),
		SessionID: currentSession.ID,
		Text:			transcript,
		Timestamp: time.Now(),
		Confidence: 0.7 + (float64(len(audioData)%30)/100),
		Language: "en",
		Speaker: "Speaker1",
		StartTime:  time.Since(currentSession.StartTime).Seconds(),
		Duration: 2.0,
	}

	segment.EndTime = segment.StartTime + segment.Duration

	return segment
}

//Count words in text

func countWords(text string) int {
	if text == " "{
		return 0
	}
	words := strings.Fields(text)
	return len(words)
}


//Truncate text dor logging
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}