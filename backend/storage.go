
package main 

import (
	f "fmt"
	"os"
	"time"
	"encoding/json"
)

const (
	dataFile = "notes.json"
)





func LoadNotesDB() (*NotesDB, error) {
    dataFile := "notes.json"

    data, err := os.ReadFile(dataFile)
    if err != nil {
        if os.IsNotExist(err) {
            // Create a new empty DB file
            defaultDB := &NotesDB{Sessions: []MeetingSession{}, Notes: []Note{}}
            SaveNotesDB(defaultDB)
            return defaultDB, nil
        }
        return nil, err
    }

    if len(data) == 0 {
        // Handle empty file
        defaultDB := &NotesDB{Sessions: []MeetingSession{}, Notes: []Note{}}
        SaveNotesDB(defaultDB)
        return defaultDB, nil
    }

    var db NotesDB
    if err := json.Unmarshal(data, &db); err != nil {
        // Handle corrupted or invalid JSON
        defaultDB := &NotesDB{Sessions: []MeetingSession{}, Notes: []Note{}}
        SaveNotesDB(defaultDB)
        return defaultDB, nil
    }

    return &db, nil
}



// SaveNotes save the entire notes database to file
 
func SaveNotesDB(db *NotesDB) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return f.Errorf("failed to write notes file: %w", err)
	}
err = os.WriteFile(dataFile, data, 0644)
if err != nil {
	return f.Errorf("failed to write notes file: %w", err)
}
return nil
}




// generate a new unique ID for a note

func generateID() string {
	return f.Sprintf("note_%d", time.Now().Unix())
}