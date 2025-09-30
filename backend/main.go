// to run used fo build -o meetnote

package main

import (
	"bufio"
	f "fmt"
	l "log"
	"os"
	"strings"
	"time"

)


const version = "3.0.0"


func main(){
	f.Println("===========================================")
	f.Println("🎙️ MeetNote Version 3 - Live Transcription")
	f.Println("===========================================")
	f.Println()
	}

	//Initialize the system
	err := initializeTranscriptionSystem()
	if err != nil {
		l.Fatalf("❌ Failed to initialize system: %v", err)
	}

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
		handleStartWebServer()
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

session, err := StartTranscriptionSession(titlte)
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
func hanldStopTranscription(){
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
	f.Printf("   Segments: %d\n", currentSession.Segmentcount)
}else {
	f.Println("🚫 No active transcription session")
}

//Show recent sessions
db, err := LoadNotesDB()
if err != nil {
	f.Printf("❌ Error loading database: %v\n", err)
	return
}

f.Printf("\n📊 Total Sessions: 5d\n", len(db.Sessions))
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
		f.Println("")
	}
}



func showUsage(){
	f.Println("Usage: meetnote <command> [options]")
	f.Println()
	f.Println("Commands:")
	f.Println(" add           Add a new note")
	f.Println(" list          List all notes or notes in a category")
	f.Println(" delete        Delete a note by ID")
	f.Println(" search        Search notes to file")
	f.Println(" export        Export notes to file")
	f.Println(" categories    List all categories")
	f.Println(" help          show detailed help")
	f.Println(" version       show version")
	f.Println()
	f.Println("Run 'meetnote help' for detailed usage examples.")
}



func showHelp() {
	f.Printf("meetnote v%s - Meeting Notes Assistant\n\n", version)

	f.Println("USAGE:")
	f.Println("   meetnote <command> [options]")
	f.Println()

	f.Println("COMMANDS:")
	f.Println("   add <text> [--category==<name>] [--tags=<tag1, tag2>]")
	f.Println("   		Add a new note with optional category and tags")
	f.Println()

	f.Println("   list [--category=<name>]")
	f.Println("   		List all notes or filter by category")
	f.Println()

	f.Println("   delete <note-id>")
	f.Println("			Delete a specific note by its ID")
	f.Println()
	f.Println("  search <keyword>")
	f.Println("			Search for notes containing the keyword")
	f.Println()
	f.Println("  export <format> <filename> [--category=<name>]")
	f.Println("			Export notes to file (formats: txt, md)")
	f.Println()
	f.Println("  categories")
	f.Println("			List all available categories")
	f.Println()


	f.Println("EXAMPLES:")
	f.Println("  meetnote add \"Discussed Q1 goals\" --category==team-standup ")
	f.Println("  meetnote add \"Fix login bug\" --tags=urgent, backend")
	f.Println("  meetnote list --category=team-standup")
	f.Println("  meetnote search \"goals\" ")
	f.Println("  meetnote export md team_notes.md  --category=team-standup")
}





