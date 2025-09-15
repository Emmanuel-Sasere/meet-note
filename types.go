package main

import (
	"time"
)



type Note struct {
	ID   			string  	`json:"id"`
	Text 			string  	`json:"text"`
	Timestamp time.Time `json:"timestamp"`
	Category 	string    `json:"category"`
	Tags			[]string	`json:"tags"` 
}


//  My jsondb at the moment
type NotesDB struct {
	Notes  []Note  `json:"notes"`
}


type ExportFormat string


const (
	FormatTXT      ExportFormat = "txt"
	FormatMarkdown ExportFormat = "md"
)