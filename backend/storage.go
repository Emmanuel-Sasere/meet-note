
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
data, err := os.ReadFile(dataFile)
if err != nil {
	if os.IsNotExist(err){
		return &NotesDB{Notes: []Note{}}, nil
	}

	return nil, f.Errorf("failed to read notes file: %w", err)
}
var db NotesDB

err = json.Unmarshal(data, &db)
if err != nil {
	return nil, f.Errorf("failed to parse notes file: %w", err)
}
return &db, nil

}


// SaveNotes save the entire notes database to file
 
func SaveNotesDB(db *NotesDB) error {
	data, err := json.MarshalIndent(db, "", " ")
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