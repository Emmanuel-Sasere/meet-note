package main

import (
	"encoding/json"
	f "fmt"
	"os"
)





type Note struct {
	Text string  `json:"text"`
}


func SaveNotes(notes []Note) error {
	data, err := json.MarshalIndent(notes, "" , " ")
	if err != nil {
		f.Println("fail to marshal notes: %w", err)
	}
	return os.WriteFile("notes.json", data, 0644)
}

func LoadNotes() ([]Note, error) {
	data, err := os.ReadFile("notes.json")

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
	return notes, nil
}

func AddNote(text string) error {
	notes, err := LoadNotes()
	if err != nil {
		return err
	}
	notes = append(notes, Note{Text: text})
	return SaveNotes(notes)
}

func DeleteNote(index int) error {
	notes, err := LoadNotes()
	if err != nil {
		return err
	}
	if index < 0 || index>= len(notes) {
		return f.Errorf("invalid note index")
	}
	notes = append(notes[:index], notes[index+1:]...)
	return SaveNotes(notes)
}