// to run used fo build -o meetnote

package main

import (
	"bufio"
	f "fmt"
	l "log"
	"os"
	"strings"
	"time"

	vosk "github.com/alphacep/vosk-api/go"
)


const version = "3.0.0"

var voskModel *vosk.VoskModel

// loadSpeechModel loads the Vosk speech recognition model from the specified path
func loadSpeechModel() error {
	var err error
	voskModel, err = vosk.NewModel("models/vosk-model-small-en-us-0.15")
	if err != nil {
		return err
	}
	return nil
}

func main(){
	f.Println("===========================================")
	f.Println("🎙️ MeetNote Version 3 - Live Transcription")
	f.Println("===========================================")
	f.Println()
	

	//Initialize the system
	err := initializeTranscriptionSystem()
	if err != nil {
		l.Fatalf("❌ Failed to initialize system: %v", err)
	}

	//Load speech recognition model
err = loadSpeechModel()
	if err != nil {
		l.Fatalf("Error loading model: %v", err)
	}
	defer voskModel.Free()

	l.Println("✅ Model loaded successfully!")



	//CHECK COMMAND LINE ARGUMENT
	if len(os.Args) < 2 {
		showMainMenu()
		return
	}

	command := os.Args[1]


	//Handle commands
	switch command {
	case "start":
		handleStartTranscription()
	case "stop":
		handleStopTranscription()
	case "status":
		handleShowStatus()
	case "sessions":
		handleListSessions()
	case "summary":
		handleShowSummary()
	case "export":
		handleExportSession()
	case "interactive", "menu":
		runInteractiveMode()
	case "server":
		if err := StartWebServer(); err != nil {
        l.Fatalf("❌ Failed to start web server: %v", err)
    	}
	case "help", "--help", "-h":
		showHelp()
	case "version", "--version", "-v":
		f.Printf("MeetNote version %s\n", version)
	default:
		f.Printf("Unknown command: %s\n\n", command)
		showMainMenu()
		os.Exit(1)
	}
}

//INITIALIZE TRANSCRIPTION SYSTEM
func initializeTranscriptionSystem() error {
	l.Println("🔧 Initializing transcription system...")

	//check if notes database exists, create if not 
	_, err := LoadNotesDB()
	if err != nil {
		l.Println("📁 Creating new notes database...")
		//Create empty database
		emptyDB := &NotesDB {
			Notes:   []Note{},
			Sessions: []MeetingSession{},
			Segments:  []TranscriptSegment{},
			Config: AudioConfig{
				SampleRate:    16000,
				Channels:    1,
				BitDepth:   16,
				ChunkDuration:  2 * time.Second,
				Language:   "en",
				MinConfidence: 0.6,
				MaxSegmentLength:  200,
				EnableNoiseSuppression: true,
			},
			LastTranscription: time.Now(),
		}

		err = SaveNotesDB(emptyDB)
		if err != nil{
			return f.Errorf("failed to create database: %w", err)
		}
	}
	l.Println("✅ System initialized successfully")
	return nil
}

//HANDLE START TRANSCRIPTION
func handleStartTranscription(){
	var title string

	//Get meeting title from command line or prompt user
	if len(os.Args) >= 3 {
		title = strings.Join(os.Args[2:], " ")

	}else {
		f.Print("Enter meeting title: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan(){
			title = scanner.Text()
		}
	}

	if title == ""{
		title = f.Sprintf("Meeting %s", time.Now().Format("2006-02-02 15:04"))
	}
	f.Printf("🎙️ Starting transcription session: %s\n", title) 

session, err := StartTranscriptionSession(title)
if err != nil {
	l.Fatalf("❌ Failed to start transcription: %v", err)
}

f.Printf("✅ Session started successfully!\n")
f.Printf("  Session ID: %s\n", session.ID)
f.Printf("  Title: %s\n", session.Title)
f.Printf("  Started: %s\n", session.StartTime.Format("15:04:05"))
f.Println()
f.Println("🛑 Recording in progress...")
f.Println("   - Speak naturally, the system will capture everything")
f.Println("  - use 'meetnote stop' to end the session")
f.Println("  - use 'meetnote status' to check current status")
f.Println()
f.Println("Press ctrl+C to stop the program (transcription will continue in background)")
}


//HANDLE STOP TRANSCRIPTION 
func handleStopTranscription(){
	f.Println("⏹️ Stopping transcription session...")

	err := StopTranscriptionSession()
	if err != nil {
		l.Fatalf("❌ Failed to stop transcription: %v", err)
	}

f.Println("✅ Transcription session stopped successfully!")
f.Println("📈 Generating meeting summary...")


// Show brief summary
db, err := LoadNotesDB()
if err == nil && len(db.Sessions) > 0 {
	lastSession := db.Sessions[len(db.Sessions)-1]
	f.Printf("\n📝 Session Summary:\n")
	f.Printf("  Title: %s\n", lastSession.Title)
	f.Printf("  Duration: %v\n", lastSession.TotalWords)
	f.Printf("  Words transcribed: %d\n",)
	f.Printf("  Key points: %d\n", len(lastSession.KeyPoints))
	f.Printf("  Action items: %d\n", len(lastSession.ActionItems))
	f.Println()
	f.Printf("Use 'meetnote summary %s' for detailed summary\n", lastSession.ID)
}



}


//HANDLE SHOW STATUS
func handleShowStatus() {
	f.Println("📈 Meetnote status")
	f.Println("==================")


	//Check if transcription is active
if currentSession != nil && currentSession.Status == "active" {
	f.Printf("🛑 Recording: %s\n", currentSession.Title)
	f.Printf("    Session ID: %s\n", currentSession.ID)
	f.Printf("    Started: %s\n", currentSession.StartTime.Format("15:04:05"))
	f.Printf("    Duration; %v\n", time.Since(currentSession.StartTime))
	f.Printf("    Words captured: %d\n", currentSession.TotalWords)
	f.Printf("   Segments: %d\n", currentSession.SegmentCount)
}else {
	f.Println("🚫 No active transcription session")
}

//Show recent sessions
db, err := LoadNotesDB()
if err != nil {
	f.Printf("❌ Error loading database: %v\n", err)
	return
}

f.Printf("\n📊 Total Sessions: %d\n", len(db.Sessions))
f.Printf("📝 Total Notes: %d\n", len(db.Notes))

if len(db.Sessions) > 0 {
	f.Println("\n⏲️ Recent Sessions:")
	//Show last 3 sessions
	start := len(db.Sessions) -3
	if start < 0 {
		start = 0
	}

	for i := len(db.Sessions) - 1; i >= start; i-- {
		session := db.Sessions[i]
		f.Printf("  %s - %s (%d words)\n", session.StartTime.Format("Jan 2 15:04"),
	session.Title,
	session.TotalWords)
	}

}
}


//HANDLE LIST SESSIONS
func handleListSessions(){
	db, err := LoadNotesDB()
	if err != nil {
		l.Fatalf("❌ Error loading database: %v", err)
	}
	if len(db.Sessions) == 0 {
		f.Println("📭 No meeting sessions found")
		f.Println("Use 'meetnote start \"Meeting Title\"' to begin transcription")
		return
	}
	
	f.Printf("📚 Meeting Sessions (%d total)\n", len(db.Sessions))
	f.Println(strings.Repeat("=", 60))
	
	for i := len(db.Sessions) - 1; i >= 0; i-- {
		session := db.Sessions[i]
		
		f.Printf("\n🎯 %s\n", session.Title)
		f.Printf("   ID: %s\n", session.ID)
		f.Printf("   Date: %s\n", session.StartTime.Format("January 2, 2006 at 15:04"))
		f.Printf("   Status: %s\n", session.Status)
		
		if session.EndTime != nil {
			f.Printf("   Duration: %v\n", session.Duration)
		} else {
			f.Printf("   Duration: %v (ongoing)\n", time.Since(session.StartTime))
		}
		
		f.Printf("   Words: %d | Segments: %d\n", session.TotalWords, session.SegmentCount)
		
		if len(session.KeyPoints) > 0 {
			f.Printf("   Key Points: %d\n", len(session.KeyPoints))
		}
		if len(session.ActionItems) > 0 {
			f.Printf("   Action Items: %d\n", len(session.ActionItems))
		}
		if len(session.Participants) > 0 {
			f.Printf("   Participants: %s\n", strings.Join(session.Participants, ", "))
		}
	}
	
	f.Println("\nUse 'meetnote summary <session-id>' to view detailed summary")
}

// HANDLE SHOW SUMMARY  
func handleShowSummary() {
	var sessionID string
	
	if len(os.Args) >= 3 {
		sessionID = os.Args[2]
	} else {
		// Show latest session summary
		db, err := LoadNotesDB()
		if err != nil || len(db.Sessions) == 0 {
			f.Println("❌ No sessions found")
			return
		}
		sessionID = db.Sessions[len(db.Sessions)-1].ID
	}
	
	// Find and display session
	db, err := LoadNotesDB()
	if err != nil {
		l.Fatalf("❌ Error loading database: %v", err)
	}
	
	var session *MeetingSession
	for _, s := range db.Sessions {
		if s.ID == sessionID || strings.Contains(s.Title, sessionID) {
			session = &s
			break
		}
	}
	
	if session == nil {
		f.Printf("❌ Session not found: %s\n", sessionID)
		f.Println("Use 'meetnote sessions' to see available sessions")
		return
	}
	
	// Display detailed summary
	f.Printf("\n📋 Meeting Summary: %s\n", session.Title)
	f.Println(strings.Repeat("=", len(session.Title)+17))
	f.Printf("📅 Date: %s\n", session.StartTime.Format("January 2, 2006"))
	f.Printf("⏰ Time: %s", session.StartTime.Format("3:04 PM"))
	if session.EndTime != nil {
		f.Printf(" - %s", session.EndTime.Format("3:04 PM"))
	}
	f.Printf(" (%v)\n", session.Duration)
	f.Printf("📊 Status: %s\n", strings.ToUpper(session.Status))
	f.Printf("💬 Words: %d | Segments: %d\n\n", session.TotalWords, session.SegmentCount)
	
	if session.Summary != "" {
		f.Println("📝 Summary:")
		f.Println(session.Summary)
		f.Println()
	}
	
	if len(session.KeyPoints) > 0 {
		f.Println("🎯 Key Points:")
		for i, point := range session.KeyPoints {
			f.Printf("%d. %s\n", i+1, point)
		}
		f.Println()
	}
	
	if len(session.ActionItems) > 0 {
		f.Println("✅ Action Items:")
		for _, item := range session.ActionItems {
			f.Printf("• %s\n", item)
		}
		f.Println()
	}
	
	if len(session.Participants) > 0 {
		f.Printf("👥 Participants: %s\n\n", strings.Join(session.Participants, ", "))
	}
	
	f.Printf("Use 'meetnote export %s summary meeting-summary.md' to export\n", session.ID)
}

// HANDLE EXPORT SESSION
func handleExportSession() {
	if len(os.Args) < 5 {
		f.Println("Usage: meetnote export <session-id> <format> <filename>")
		f.Println("Formats: summary, transcript, markdown")
		f.Println("Example: meetnote export session_123 summary my-meeting.md")
		return
	}
	
	sessionID := os.Args[2]
	format := ExportFormat(os.Args[3])
	filename := os.Args[4]
	
	f.Printf("📤 Exporting session to %s...\n", filename)
	
	err := exportSessionSummary(sessionID, format, filename)
	if err != nil {
		l.Fatalf("❌ Export failed: %v", err)
	}
	
	f.Printf("✅ Session exported successfully to %s\n", filename)
}



// INTERACTIVE MODE
func runInteractiveMode() {
	f.Println("🎛️ Interactive Mode")
	f.Println("Type 'help' for commands, 'quit' to exit")
	f.Println()
	
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		f.Print("meetnote> ")
		
		if !scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		
		if input == "quit" || input == "exit" {
			f.Println("👋 Goodbye!")
			break
		}
		
		// Parse command
		parts := strings.Fields(input)
		
		switch parts[0] {
		case "start":
			if len(parts) > 1 {
				title := strings.Join(parts[1:], " ")
				_, err := StartTranscriptionSession(title)
				if err != nil {
					f.Printf("❌ Error: %v\n", err)
				} else {
					f.Println("✅ Transcription started!")
				}
			} else {
				f.Println("Usage: start <meeting title>")
			}
			
		case "stop":
			err := StopTranscriptionSession()
			if err != nil {
				f.Printf("❌ Error: %v\n", err)
			} else {
				f.Println("✅ Transcription stopped!")
			}
			
		case "status":
			handleShowStatus()
			
		case "sessions":
			handleListSessions()
			
		case "help":
			showInteractiveHelp()
			
		default:
			f.Printf("Unknown command: %s (type 'help' for available commands)\n", parts[0])
		}
		
		f.Println()
	}
}

// HELP FUNCTIONS

func showMainMenu() {
	f.Println("Usage: meetnote <command> [options]")
	f.Println()
	f.Println("🎙️ TRANSCRIPTION COMMANDS:")
	f.Println("  start [title]    Start live transcription session")
	f.Println("  stop             Stop current transcription session")
	f.Println("  status           Show current transcription status")
	f.Println()
	f.Println("📚 SESSION MANAGEMENT:")
	f.Println("  sessions         List all meeting sessions")
	f.Println("  summary [id]     Show detailed session summary")
	f.Println("  export <id> <format> <file>  Export session")
	f.Println()
	f.Println("🎛️ INTERFACE:")
	f.Println("  interactive      Run in interactive mode")
	f.Println("  server           Start web server + dashboard")
	f.Println("  help             Show detailed help")
	f.Println("  version          Show version")
	f.Println()
	f.Println("Examples:")
	f.Println("  meetnote start \"Team Standup\"")
	f.Println("  meetnote summary")
	f.Println("  meetnote export session_123 summary meeting.md")
}

func showHelp() {
	f.Printf("MeetNote v%s - Live Meeting Transcription\n\n", version)
	
	f.Println("OVERVIEW:")
	f.Println("MeetNote automatically transcribes your meetings in real-time,")
	f.Println("generates summaries, and extracts action items - all locally and free!")
	f.Println()
	
	
	showMainMenu()
	
	f.Println()
	f.Println("FEATURES:")
	f.Println("• 🎙️ Live audio transcription (offline speech-to-text)")
	f.Println("• 📝 Automatic note generation and organization")  
	f.Println("• 🧠 AI-powered summarization (key points & action items)")
	f.Println("• 📊 Session management (track multiple meetings)")
	f.Println("• 📤 Multiple export formats (Markdown, text, JSON)")
	f.Println("• 🌐 Web dashboard for easy management")
	f.Println()
}
// SHOW INTERACTIVE HELP
func showInteractiveHelp() {
	f.Println("📖 Interactive Mode Help")
	f.Println("========================")
	f.Println()
	f.Println("Available commands:")
	f.Println("  start <title>    Start a new transcription session")
	f.Println("  stop             Stop the current transcription session")
	f.Println("  status           Show current transcription status")
	f.Println("  sessions         List all past meeting sessions")
	f.Println("  help             Show this help menu")
	f.Println("  quit / exit      Leave interactive mode")
	f.Println()
	f.Println("Tips:")
	f.Println(" • You don’t need to type 'meetnote' before commands here.")
	f.Println(" • Example: just type 'start Team Meeting' to begin.")
}
