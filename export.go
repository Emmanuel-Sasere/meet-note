package main
import (
	f "fmt"
	"os"
	"strings"
	"time"
)


// Function to export notes to specified format
func ExportNotes(category string, format ExportFormat, filename string) error {
	var notes []Note
	var err error

	if category == "" {
		notes, err = GetAllNotes()

	} else {
		notes, err = FilterNotesByCategory(category)
	}
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		return f.Errorf("no notes found to export")
	}

	var content string

	switch format {
	case FormatTXT: 
		content = generateTXTContent(notes, category)
	case FormatMarkdown:
		content = generateMarkdownContent(notes, category)
	default:
		return f.Errorf("unsupported export format: %s", format)
	}

	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return f.Errorf("failed to write export file: %w", err)
	}

	return nil 

}



// function to create plain text content 
func generateTXTContent(notes []Note, category string) string{
	var builder strings.Builder

	if category != "" {
		builder.WriteString(f.Sprintf("Notes for: %s\n", category))
	} else {
		builder.WriteString("All Notes\n")
	}

	builder.WriteString(f.Sprintf("Exported on: %s\n", time.Now().Format("2006-01-02 15:04:05")))

	builder.WriteString(strings.Repeat("=====", 10) + "\n\n")

for i, note := range notes {
	timeStr := note.Timestamp.Format("2006-01-02 15:04")

	builder.WriteString(f.Sprintf("%d. [%s] %s\n", i+1, timeStr, note.Category))
	builder.WriteString(f.Sprintf("  %s\n", note.Text))

	if len(note.Tags) > 0 {
		tagsStr := strings.Join(note.Tags, ",")
		builder.WriteString(f.Sprintf("   Tags: %s\n", tagsStr))
	}
	builder.WriteString("\n")
}
return builder.String()
}


func generateMarkdownContent(notes []Note, category string) string{
	var builder strings.Builder


	if category != ""{
		builder.WriteString(f.Sprintf("# Notes for: %s\n\n", category))
	} else {
		builder.WriteString("# All Notes\n\n")
	}

	builder.WriteString(f.Sprintf("*Exported on: %s*\n\n", time.Now().Format("2006-01-02 15:04:05")))

	for _, note := range notes{
		builder.WriteString(f.Sprintf("## %s\n", note.Category))
		builder.WriteString(f.Sprintf("**Date:** %s\n", note.Timestamp.Format("2006-01-02 15:04")))
		builder.WriteString(f.Sprintf("%s\n\n", note.Text))


		if len(note.Tags) > 0 {
			tagsStr := strings.Join(note.Tags, ", ")
			builder.WriteString(f.Sprintf("**Tags:** %s\n\n", tagsStr))
		}
builder.WriteString("---\n\n")
	}
	return builder.String()
}