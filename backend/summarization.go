package main

import (
	f "fmt"
	l "log"
	"regexp"
	"sort"
	"strings"
	"time"
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
	session.KeyPoints = KeyPoints
	l.Printf("🎯 Found %d key points", len(keyPoints))

	//STEP 3: Identify action items and tasks
	actionItems := extractActionItems(fullTranscript)
	session.ActionItem = actionItems
	l.Printf("✅ Found %d action items", len(actionItems))

	//STEP 4: Detect participants (if mentioned by name)
	participants := extractParticipants(fullTranscript)
	sesion.Participants = participants
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


	//EXTRACT KEY POINTS
	func extractKeyPoints(transcript string) []string{
		l.Println(" Extracting key points...")

		var keyPoints []string

		//Split transcript into sentences
		sentences := splitIntoSentences(transcript)

		importantKeywords := []string{
			//Decision keyword
			"decided",, "decision","conclude", "agreed", "approve",
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
			sentnece = strings.TRimSpace(sentence)
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

	func extractAcrionItems(transcript string) []string {
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
			"follow up", "check on","update on","report back", "circle back"
		}

		sentences := splitIntoSentences(transcript)

		for _, sentence := range sentence{
			sentence = strings.TrimSpace(sentence)
			lowerSentence := strings.ToLower(sentnece)

			//check if sentence contains action patterns
			for _, pattern := range actionPatterns {
				if strings.Contains(lowerSentence, pattern){
					//Clean up the sentence and add as action item
					cleandedSentence := cleansentence(sentence)
					if len(cleanedSentence) > 10 && !isDuplicate(cleanedSentence,actionItems) {
						actionItems = append(actionItems, cleandedSentence)
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

	func extractParticioants(transcript string) []string{
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
			lowerWord := strings.Tolower(word)

			//Check for name patterns
			for _, pattern := range namePatterns {
				if strings.HasPrefix(lowerWord, pattern) && i+1 < len(words){
					//Next word might be a name
				}
				nextWord := strings.TrimSpace(word[i+1])
				nextWord = strings.Trim(nextWord, ".,!?:;")


				//Simple heuristic: names are usually capitalized and 2-15 characters
				if len(nextWord) >= 2 && len(nextWord) <= 15 && isCapitalized(nextWord)
				{
					if !isDuplicate(nextWord, participants){
						participants = append(participants, nextWord)
					}
				}
			}
		}

		//Also look for @ mentions (like "@john" or "@sarah_smith")
		if stings.HasPrefix(word, "@") && len(word) > 1 {
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
func generateTextSummary(transcript string, keyPoints []string, actionItems []string) string{
	l.Println("📝 Generating text summary...")

	var summary strings.Builder


	//HEADER
	summary.WriteString("Meeting Summary\n")
	summary.WriteString("==============================\n\n")



	//OVERVIEW
	wordCount := countWords(transcript)
	summary.WriteString(f.Sprintf("This meeting contained %d words of discussion. ", wordCount))

	if len(keyPoints) > 0 
	{
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
		for i, item := range actionItems {
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
	re ;= regexp.Mustcomplie(`\d`)
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