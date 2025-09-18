// to run used fo build -o meetnote

package main

import (
	f "fmt"
	"os"
	"strings"
)


const version = "2.0.0"


func main(){
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	command := os.Args[1]


	switch command {
	case "add":
		handleAdd()
	case "list":
		handleList()
	case "delete":
		handleDelete()
	case "search":
		handleSearch()
	case "export":
		handleExport()
	case "categories":
		handleCategories()
	case "help", "--help", "-h":
		showHelp()
	case "version", "--version", "-v":
		f.Printf("meetnote version %s\n", version)
	default:
		f.Printf("Unkown command: %s\n", command)
		showUsage()
		os.Exit(1)
	}
}


func handleAdd(){
	if len(os.Args) < 3 {
		f.Println("Usage: meetnote add \"note text\" [--category category] [--tags=<tag1, tag2>]")
		return
	}
	text := os.Args[2]
	category := ""
	var tags []string

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]

		if strings.HasPrefix(arg, "--category="){
			category = strings.TrimPrefix(arg, "--category=")
		}else if strings.HasPrefix(arg, "--tags="){
			tagStr := strings.TrimPrefix(arg, "--tags=")

			tags = strings.Split(tagStr, ",")

			for j, tag := range tags {
				tags[j] = strings.TrimSpace(tag)
			}
		}
	}

	err := AddNote(text, category, tags)
	if err != nil {
		f.Printf("Error adding note: %v\n", err)
		os.Exit(1)
	}
	f.Printf("\u2713 Note added successfully")
	if category != "" {
		f.Printf(" to category '%s'", category)
	}
	if len(tags) > 0 {
		f.Printf("with tags: %s", strings.Join(tags, ", "))
	}
	f.Println()
}

func handleList() {
	var notes []Note
	var err error

	if len(os.Args) >= 3 && strings.HasPrefix(os.Args[2], "--category=") {
		category := strings.TrimPrefix(os.Args[2], "--category=")
		notes, err = FilterNotesByCategory(category)
	} else {
		notes, err = GetAllNotes()
	}

	if err != nil {
		f.Printf("Error loading notes: %v\n", err)
		os.Exit(1)
	}
	if len(notes) == 0 {
		f.Println("No notes found.")
		return
	}
	f.Printf("Found %d note(s):\n\n", len(notes))
	for i, note := range notes {
		f.Printf("%d. [%s] %s\n", i+1, note.Timestamp.Format("2006-01-02 15:04"), note.Category)
		f.Printf("  ID: %s\n", note.ID)
		f.Printf("  %s\n", note.Text)

		if len(note.Tags) > 0 {
			f.Printf("  Tags: %s\n", strings.Join(note.Tags, ", "))
		}

		f.Println()
	}
}


func handleDelete(){
	if len(os.Args) < 3 {
		f.Println("Usage: meetnote delete <note-id>")
		f.Println("Use 'meetnite list'to see not IDs")
		return
	}

	noteID := os.Args[2]
	err := DeleteNoteByID(noteID)
	if err != nil {
		f.Printf("Error deleting note: %v\n", err)
		os.Exit(1)
	}

	f.Println("\u2713 Note deleted successfully")
}


func handleSearch(){
	if len(os.Args) < 3 {
		f.Println("Usage: meetnote search <search-term>")
		return
	}

	searchTerm := os.Args[2]
	notes, err := SearchNotes(searchTerm)
	if err != nil {
		f.Printf("Error searching notes: %v\n", err)
		os.Exit(1)
	}

	if len(notes)  == 0 {
		f.Printf("No matching notes found containing '%s'\n", searchTerm)
		return
	}

	f.Printf("Found %d note(s) containing '%s':\n\n", len(notes), searchTerm)
	for i, note := range notes {
		f.Printf("%d. [%s] %s\n", i+1, note.Timestamp.Format("2006-01-02 15:04"), note.Category)
		f.Printf("  ID: %s\n", note.ID)
		f.Printf("  %s\n", note.Text)
		f.Println()
	}
}


func handleExport() {
	if len(os.Args) < 4 {
		f.Println("Usage: meetnote export <format> <filename> [--category=<name>]")
		f.Println("Formats: txt, md")
		return
	}

	formatStr := os.Args[2]
	filename := os.Args[3]
	category := ""

	if len(os.Args) >= 5 && strings.HasPrefix(os.Args[4], "--category="){
		category = strings.TrimPrefix(os.Args[4], "--category=")
	}

	var format ExportFormat
	switch formatStr{
	case "txt":
		format = FormatTXT
	case "md":
		format = FormatMarkdown
	default:
		f.Printf("Unsupported format: %s\n", formatStr)
		f.Printf("Supported formats: txt, md")
		return
	}

	err := ExportNotes (category, format, filename)
	if err != nil {
		f.Printf("Error exporting notes: %v\n", err)
		os.Exit(1)
	}

	f.Printf("\u2713 Notes exported to %s\n", filename)
}


func handleCategories(){
	categories, err := GetCategories()
	if err != nil {
		f.Printf("Error loading categories: %v\n", err)
		os.Exit(1)
	}

	if len(categories) == 0 {
	f.Println("No categories found.")
	return
}
f.Printf("Available categories (%d):\n", len(categories))
for _, category := range categories {
	f.Printf("  - %s\n", category)
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
	f.Println("   		Delete a specific note by its ID")
	f.Println()
	f.Println("  search <keyword>")
	f.Println("				Search for notes containing the keyword")
	f.Println()
	f.Println("  export <format> <filename> [--category=<name>]")
	f.Println("				Export notes to file (formats: txt, md)")
	f.Println()
	f.Println("  categories")
	f.Println("				List all available categories")
	f.Println()


	f.Println("EXAMPLES:")
	f.Println("  meetnote add \"Discussed Q1 goals\" --category==team-standup ")
	f.Println("  meetnote add \"Fix login bug\" --tags=urgent, backend")
	f.Println("  meetnote list --category=team-standup")
	f.Println("  meetnote search \"goals\" ")
	f.Println("  meetnote export md team_notes.md  --category=team-standup")
}





