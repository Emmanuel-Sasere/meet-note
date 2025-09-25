package main

import (
	f "fmt"
	l "log"
	"regexp"
	"sort"
	"strings"
	"time"
	"os"
)



//Function to generate summary
func generateSessionSummary(session *MeetingSession) error {
	l.Printf("📊 Generating summary for session: %s", session.Title)

	//Get all notes from this session
	notes, err := getNotesBySessionID(session.ID)
	if err != nil {
		return f.Errorf("Failed to get session notes: %w", err)
	}

	if len(notes) == 0 {
		l.Println("⚠️ No notes found for session - creating empty summary")
		session.Summary = "No transcription data available for this session."
		return nil
	}

	//STEP 1: Combine all transcripts into one text
	fullTranscript := combineTranscripts(notes)
	l.Printf("📑 Combined transcript: %d words", countWords(fullTranscript))

	//STEP 2: Extract key points using text analysis
	keyPoints := extractKeyPoints(fullTranscript)
	session.KeyPoints = keyPoints
	l.Printf("🎯 Found %d key points", len(keyPoints))

	//STEP 3: Identify action items and tasks
	actionItems := extractActionItems(fullTranscript)
	session.ActionItems = actionItems
	l.Printf("✅ Found %d action items", len(actionItems))

	//STEP 4: Detect participants (if mentioned by name)
	participants := extractParticipants(fullTranscript)
	session.Participants = participants
	l.Printf("👀 Detected %d participants", len(participants))


	//STEP 5: Generate overall summary
	summary := generateTextSummary(fullTranscript, keyPoints, actionItems)
	session.Summary = summary

	//STEP 6: Calculte final statistics
	session.TotalWords = countWords(fullTranscript)

	l.Printf("✅ Summary generated successfully")
	return nil
}


//COMBINE TRANSCRIPTS
func combineTranscripts(notes []Note) string{
	var transcriptParts []string


	//Sort notes by timestamp to maintain chronological order
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Timestamp.Before(notes[j].Timestamp)
	})

	//Combine all transcript text
	for _, note := range notes {
		if note.IsTranscript && note.Text != ""{
			//Add timestamp marker for context
			timeMarker := note.Timestamp.Format("15:04")
			transcriptParts = append(transcriptParts, f.Sprintf("[%s] %s",timeMarker, note.Text))
		}
	}

	return strings.Join(transcriptParts, "\n")
}

	//EXTRACT KEY POINTS
	func extractKeyPoints(transcript string) []string{
		l.Println(" Extracting key points...")

		var keyPoints []string

		//Split transcript into sentences
		sentences := splitIntoSentences(transcript)

		importantKeywords := []string{
			//Decision keyword
			"decided", "decision", "conclude", "agreed", "approve",
			//Action Keywords
			"will do","action","task", "responsible","deadline","by next week",
			//Import tpics
			"next steps", "follow up", "schedule", "meeting", "discuss. further",
		}

		//Score each sentence based on keyword matches
		type ScoredSentence struct{
			Text string
			Score int
		}

		var scoredSentences []ScoredSentence
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			if len(sentence) < 10{
				continue
			}

			score := 0
			lowerSentence := strings.ToLower(sentence)

			//Count keyword matches
			for _, keyword := range importantKeywords {
				if strings.Contains(lowerSentence,keyword){
					score += 2 //Each keyword match add 2 points
				}
			}

			//Bonus points for questions(often important in meetings)
			if strings.Contains(sentence, "?"){
				score += 1
			}


			///Bonus points for numbers (often indicate decisions or deadlines)
			if containsNumbers(sentence){
				score += 1
			}

			if score > 0 {
				scoredSentences = append(scoredSentences, ScoredSentence{
					Text: cleanSentence(sentence),
					Score: score,
				})
			}
		}

		//Sort by score (highest first)
		sort.Slice(scoredSentences, func(i, j int) bool {
			return scoredSentences[i].Score > scoredSentences[j].Score
		})

		//Take top 5-10 key points
		maxPoints := 8
		if len(scoredSentences) < maxPoints {
			maxPoints = len(scoredSentences)
		}

		for i := 0; i < maxPoints; i++ {
			keyPoints = append(keyPoints, scoredSentences[i].Text)
		}

		return keyPoints

	}

	//EXTRACT ACTION ITEMS

	func extractActionItems(transcript string) []string {
		l.Println("📔 Extracting action items...")

		var actionItems []string

		//Action item patterns to look for
		actionPatterns := []string {
			//Assignment patterns
			"will do", "will handle", "will take care", "will work on",
			"responsible for", "assigned to", "owns this", "taking on",
			//Taks patterns
			"need to", "should", "must", "have to", "task", "todo",
			//Deadline patterns
			"by next week", "by friday", "deadline", "due date", "complete by",
			//Follow-up patterns
			"follow up", "check on", "update on", "report back", "circle back",
		}

		sentences := splitIntoSentences(transcript)

		for _, sentence := range sentences{
			sentence = strings.TrimSpace(sentence)
			lowerSentence := strings.ToLower(sentence)

			//check if sentence contains action patterns
			for _, pattern := range actionPatterns {
				if strings.Contains(lowerSentence, pattern){
					//Clean up the sentence and add as action item
					cleanedSentence := cleanSentence(sentence)
					if len(cleanedSentence) > 10 && !isDuplicate(cleanedSentence,actionItems) {
						actionItems = append(actionItems, cleanedSentence)
					}
					break
				}
			}
		}



		//Limit to reasonable number of action items
		if len(actionItems) > 10 {
			actionItems = actionItems[:10]
		}

		return actionItems
	}

	//EXTRACT PARTICIPANTS

	func extractParticipants(transcript string) []string{
		l.Println("👀 Extracting participants...")

		var participants []string

		//Look for common name patterns


		//Patterns that might indicate names
		namePatterns := []string{
			"thanks ", "thank you ",
			"hi ", "hello ",
			"yes ", "yeah ",
			"@",
		}

		words := strings.Fields(transcript)

		for i, word := range words {
			lowerWord := strings.ToLower(word)

			//Check for name patterns
			for _, pattern := range namePatterns {
				if strings.HasPrefix(lowerWord, pattern) && i+1 < len(words){
					//Next word might be a name
				
				nextWord := strings.TrimSpace(words[i+1])
				nextWord = strings.Trim(nextWord, ".,!?:;")


				//Simple heuristic: names are usually capitalized and 2-15 characters
				if len(nextWord) >= 2 && len(nextWord) <= 15 && isCapitalized(nextWord){
					if !isDuplicate(nextWord, participants){
						participants = append(participants, nextWord)
					}
				}
			}
		}
	

		//Also look for @ mentions (like "@john" or "@sarah_smith")
	
		if strings.HasPrefix(word, "@") && len(word) > 1 {
			name := strings.TrimPrefix(word, "@")
		name = strings.Trim(name, ".,!?:;")
	if len(name) >= 2 && !isDuplicate(name, participants){
		participants = append(participants, name)
	}
	
}
}

	return participants
}


//GENERATE TEXT SUMMARY
func generateTextSummary(transcript string, keyPoints []string, actionItems []string) string {
	l.Println("📝 Generating text summary...")

	var summary strings.Builder


	//HEADER
	summary.WriteString("Meeting Summary\n")
	summary.WriteString("==============================\n\n")



	//OVERVIEW
	wordCount := countWords(transcript)
	summary.WriteString(f.Sprintf("This meeting contained %d words of discussion. ", wordCount))

	if len(keyPoints) > 0 {
		summary.WriteString(f.Sprintf("Key topics included: %s. ", strings.Join(keyPoints[:min(3, len(keyPoints))], ", ")))
	}

	if len(actionItems) > 0 {
		summary.WriteString(f.Sprintf("%d action item were identified.", len(actionItems)))
	}

	summary.WriteString("\n\n")

	//KEY POINTS SECTION
	if len(keyPoints) > 0 {
		summary.WriteString("Key Discussion Points:\n")
		for i, point := range keyPoints {
			summary.WriteString(f.Sprintf("%d. %s\n", i+1, point))
		}
		summary.WriteString("\n")
	}

	//Action Items section
	if len(actionItems) > 0 {
		summary.WriteString("Action Items:\n")
		for _, item := range actionItems {
			summary.WriteString(f.Sprintf(". %s\n", item))

		}
		summary.WriteString("\n")
	}

	//FOOTER
	summary.WriteString(f.Sprintf("Summary generated on %s\n", time.Now().Format("January 2, 2006 at 3:04 PM")))

	return summary.String()
}

//HELPER FUNCTION FOR TEXT PROCESSING

func splitIntoSentences(text string) []string {
	//remove timestamp markers like [15:30]
	re := regexp.MustCompile(`\[\d{2}\]\s*`)
	text = re.ReplaceAllString(text, "")

	//split on sentence ending punctuation
	sentences := regexp.MustCompile(`[.!?]+\s`).Split(text,-1)

	var cleanSentences []string
	for _, sentence := range sentences{
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 5 {
			cleanSentences = append(cleanSentences, sentence)
		}
	}

	return cleanSentences
}

//check if text contains numbers
func containsNumbers(text string) bool {
	re := regexp.MustCompile(`\d`)
	return re.MatchString(text)
}


//Check if word is capitalized (likely aproper noun/name)
func isCapitalized(word string) bool {
	if len(word) == 0 {
		return false
	}
	first := rune(word[0])
	return first >= 'A' && first <= 'Z'
}

//Clean up sentence (remove extra whitespace , timestramp, etc)

func cleanSentence(sentence string) string{
	//Remove timestamp markers
	re := regexp.MustCompile(`\[\d{2}:\d{2}\]\s*`)
	sentence = re.ReplaceAllString(sentence, "")

	//Remove extra whitespace
	sentence = regexp.MustCompile(`\s+`).ReplaceAllString(sentence, " ")

	//TRim and capitalize first letter
sentence = strings.TrimSpace(sentence)
if len(sentence) > 0 {
	sentence = strings.ToUpper(string(sentence[0])) + sentence[1:]
}

return sentence
}


//Check if item is already in list (case-insensitive)
func isDuplicate(item string, list []string) bool {
	lowerItem := strings.ToLower(item)
	for _, existing := range list {
		if strings.ToLower(existing) == lowerItem {
			return true
		}
	}
	return false
}

//Get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a 
	}
	return b
}


//DATABASE FUNCTIONS FOR SESSIONS

//Get all notes belonging to a specific session
func getNotesBySessionID(sessionID string) ([]Note, error) {
	//Load all notes
	allNotes, err := GetAllNotes()
	if err != nil {
		return nil, err
	}

	//Filter by session ID
	var sessionNotes []Note
	for _, note := range allNotes {
		if note.SessionID == sessionID {
			sessionNotes = append(sessionNotes, note)
		}
	}

	return sessionNotes, nil
}

//Save session to database
func saveSessionToDB(session *MeetingSession) error {
	//Load exiting database
	db, err := LoadNotesDB()
	if err != nil {
		return err
	}

	//Check if session already exits (update) or is new (insert)
	found := false
	for i, existing := range db.Sessions {
		if existing.ID == session.ID {
			//Update existing session
			db.Sessions[i] = *session
			found = true
			break
		}
	} 

	//If not found, add new session
	if !found{
		db.Sessions = append(db.Sessions, *session)
	}


	//Save back to file
	return SaveNotesDB(db)
}


//Save transcript segment to database
func saveSegmentToDB(segment TranscriptSegment) error {
	//Load existing database
	db, err := LoadNotesDB()
	if err != nil {
		return err
	}


	//Add segment to database
	db.Segments = append(db.Segments, segment)

	//Add segment to database
	db.Segments = append(db.Segments, segment)

	//Save back to file
	return SaveNotesDB(db)
}





//ADVANCED SUMMARIZATION (Future enhancement)

//this function could integrate with AI service for better summaries
func generateAIsSummary(transcript string) string {


	wordCount := countWords(transcript)
	if wordCount < 50 {
		return "Short discussion with limited content."
	}else if wordCount < 200 {
		return "Brief meeting covering several topics with some actionable items."
	}else {
		return "Extended discussion with multiple topics, decisions, and follow-up items identified"
	}
}


//EXPORT SESSION SUMMARY
func exportSessionSummary(sessionID string, format ExportFormat, filename string) error {
	//Find the session
	db, err := LoadNotesDB()
	if err != nil {
		return err
	}

	var session *MeetingSession
	for _, s := range db.Sessions {
		if s.ID == sessionID {
			session = &s
			break
		}
	}

	if session == nil {
		return f.Errorf("session not found: %s", sessionID)
	}

	//enerate content based on format
	var content string
	switch format {
	case FormatSummary:
		content = session.Summary
	case FormatTranscript:
		//get full transcript with timestamps
		notes, err := getNotesBySessionID(sessionID)
		if err != nil {
			return err
		}

		content = combineTranscripts(notes)
	case FormatMarkdown:
		content = formatSessionAsMarkdown(session)
	default:
		return f.Errorf("unsupported format for session export: %s", format)
	}

	//Write to file
	return writeToFile(filename, content)
}


//Format session as markdown
func formatSessionAsMarkdown(session *MeetingSession) string {
	var md strings.Builder

md.WriteString(f.Sprintf("# %s\n\n", session.Title))
md.WriteString(f.Sprintf("**Date:** %s\n", session.StartTime.Format("January 2, 2006")))
md.WriteString(f.Sprintf("**Duration:** %v\n", session.Duration))
md.WriteString(f.Sprintf("**Status:** %s\n\n", session.Status))


if len(session.Participants) > 0 {
	md.WriteString("## Participants\n")
	for _, participant := range session.Participants {
		md.WriteString(f.Sprintf("- %s\n", participant))
	}
	md.WriteString("\n")
}

md.WriteString("## Summary\n")
md.WriteString(session.Summary)
md.WriteString("\n\n")


if len(session.KeyPoints) > 0 {
	md.WriteString("## Key Points\n")
	for _, point := range session.KeyPoints {
		md.WriteString(f.Sprintf("- %s\n", point))
	}
	md.WriteString("\n")
}

if len(session.ActionItems) > 0 {
	md.WriteString("## Action Item\n")
	for _, item := range session.ActionItems {
		md.WriteString(f.Sprintf("- [ ] %s\n", item))
	}
	md.WriteString("\n")
}

md.WriteString(f.Sprintf("## Statistics\n"))
md.WriteString(f.Sprintf("- Total Words: %d\n", session.TotalWords))
md.WriteString(f.Sprintf("- Segments: %d\n",  session.SegmentCount))

return md.String()
}

func writeToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}
