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
		err := AddNotes(note)
		if err != nil {
			f.Println ("Error adding note: %w", err)
		}else{
			f.Println("Note added successfully.", note)
		}

	case "list":
		

	}
		
}