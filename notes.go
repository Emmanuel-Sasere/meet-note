package main

import (
	"encoding/json"
	f "fmt"
	"os"
	"time"
	"strconv"
	"strings"
)





type Note struct {
	ID int  						`json:"id"`
	Content string 			`json:"content"`
	Meeting string 			`json:"meeting"`
	CreatedAt time.Time `json:"created_at"`

}

const notesfile = "notes.json"


func SaveNotes(notes []Note) error {
	data, err := json.MarshalIndent(notes, "" , " ")
	if err != nil {
		f.Println("fail to marshal notes: %w", err)
	}
	return os.WriteFile(notesfile, data, 0644)
}

func LoadNotes() ([]Note, error) {
	data, err := os.ReadFile(notesfile)

	if err != nil {
		if os.IsNotExist(err) {
			return []Note{}, nil
		}
		return nil, err
	}
	var notes []Note
	err = json.Unmarshal(data, &notes)
	if err != nil {
		f.Println("fail to unmarshal notes: %w", err)
		
	}
	return notes, err
}



func AddNote(text string) error {
	notes, err := LoadNotes()
	if err != nil {
		return err
	}
	newNote := Note{
		ID: 		len(notes) + 1,
		Content: content,
		Meeting: meeting,
		CreatedAt: time.Now(),
	}
	notes = append(notes, newNote)
	return SaveNotes(notes)
}

func ListNotes() error {
	notes, err := LoadNotes()
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		f.Println("No notes found.")
		return nil
	}
	for _, note := range notes {
		f.Printf("[%d] %s | Meeting: %s | Created At: %s/n", note.ID, note.Content, note.Meeting, note.CreatedAt.Format("2006-01-02 15:04"))
	}
	return nil
}

func DeleteNote(idStr string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return f.Errorf("invalid ID")
	}
	notes, err := LoadNotes()
	if err != nil {
		return err
	}
	newNotes := []Note{}
	for _, n := range notes {
		if n.ID != id {
			newNotes =append(newNotes, n)
		}
	}
	return SaveNotes(notes)
}


func SearchNotes(keyword string) error {
	notes, err := LoadNotes()
	if err != nil {
		return err
	}
	found := false
	for _, n := range notes {
		if strings.Contains(strings.ToLower(n.Content), strings.ToLower(keyword)) || strings.Contains(strings.ToLower(n.Meeting), strings.ToLower(keyword)) {
			f.Printf("[%d] %s | Meeting: %s | Created At: %s\n", n.ID, n.Content, n.CreatedAt.Format("2006-01-02 15:04"))
			found = true
		}
	}
	if !found {
		f.Println("No matching notes found.")
	}
	return nil
}



func ExportNotes(meeting, format string) error {
	notes, err := LoadNotes()
	if err != nil {
		return err
	}
	var exported []Note
	for _, n := range notes {
		if strings.EqualFold(n.Meeting, meeting) {
			exported = append(exported, n)
		}
	}
	
}