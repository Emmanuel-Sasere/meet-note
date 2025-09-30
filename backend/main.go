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





