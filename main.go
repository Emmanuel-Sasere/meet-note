package main

import (

f	"fmt"
	"os"
)

func main () {
	if len(os.Args) < 2 {
		f.Println("Usage: go run main.go [add | list | delete] [note]")
		return
	}

	command := os.Args[1]

	switch command {
	case "add":
		if len(os.Args) < 3 {
			f.Println("Please provide a note to add.")
			return
		}

		note := os.Args[2]
		err := AddNote(note)
		if err != nil {
			f.Println ("Error adding note: %w", err)
		}else{
			f.Println("Note added successfully.", note)
		}

	case "list":
		notes, err := LoadNotes()
		if err != nil {
			f.Println("Error loading notes: %w", err)
			return
		}
		if len(notes) == 0 {
			f.Println("No notes found")
			return
		}

		f.Println("Notes:")
		for i, n := range notes {
			f.Printf("%d. %s\n", i+1, n.Text)
		}

	case "delete":
		if len(os.Args) < 3 {
			f.Println("Please provide the number of the note to delete.")
			return
		}
		var index int
		f.Sscanf(os.Args[2], "%d", &index)
		
		err := DeleteNote(index - 1)
		 
		if err != nil {
			f.Println("Error deleting note:", err)
		}else{
			f.Println("Note deleted successfully.")
		}

	default: 
	f.Println("Invalid command.", command)

	}
		
}