package main

import (
	f "fmt"
	"strings"
	"time"
)


func AddNote(text, category string, tags []string) error {
	db, err := LoadNotesDB()
	if err != nil {
		return err
	}

	note := Note{
		ID:    			generateID(),
		Text:  			text,
		Timestamp: 	time.Now(),
		Category:  	category,
		Tags: 			tags,
	}

	db.Notes = append(db.Notes, note)
	return SaveNotesDB(db)
}


//Delete Note By Id
func DeleteNoteByID(id string) error {
	db, err := LoadNotesDB()
	if err != nil {
		return err
}

index := -1 
for i, note := range db.Notes {
	if note.ID == id {
		index = i
		break
	}
}

if index == -1 {
	return f.Errorf("note with Id %s not found", id)
}
db.Notes = append(db.Notes[:index], db.Notes[index+1:]...)
return SaveNotesDB(db)
}



// Search Notes function

func SearchNotes(searchTerm string)([]Note, error){
	db, err := LoadNotesDB()
	if err != nil {
		return nil, err
	}

	var results []Note

	searchLower := strings.ToLower(searchTerm)

	for _, note := range db.Notes {
		textMatch := strings.Contains(strings.ToLower(note.Text), searchLower)
		categoryMatch := strings.Contains(strings.ToLower(note.Category), searchLower)

		if textMatch || categoryMatch {
			results = append(results, note)
		}
	}
	return results, nil
}



func FilterNotesByCategory(category string) ([]Note, error) {
	db, err := LoadNotesDB()
	if err != nil {
		return nil, err
	}

	var results []Note
	for _, note := range db.Notes {
		if strings.EqualFold(note.Category, category){
			results = append(results, note)
		}
	}

	return results, nil
}


//Get all notes 

func GetAllNotes()([]Note, error){
	db, err := LoadNotesDB()
	if err != nil {
		return nil, err
	}
	return db.Notes, nil
}



func GetCategories() ([]string, error) {
	db, err := LoadNotesDB()
	if err != nil {
		return nil, err
	}

	categoryMap := make(map[string]bool)


	for _, note := range db.Notes {
		if note.Category != ""{
		categoryMap[note.Category] = true
		}
	}

	var categories []string
	for category := range categoryMap {
		categories = append(categories, category)
	}
	return categories, nil
}