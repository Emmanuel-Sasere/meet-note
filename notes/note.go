package note

import (
	"encoding/json"
	"time"
)


type Note struct {
	Title  		string  			`json:"title"`
	Content 	string 			`json:"content"`
	CreatedAt time.Time 	`json:"created_at"`
}