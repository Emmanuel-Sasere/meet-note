🎯 Project Goal

Build an open-source assistant that can automatically take, organize, and summarize notes from any online meeting platform (Google Meet, Zoom, Jitsi, …) — for free.
Each version builds upon the last, teaching Go step by step.

🛠️ Version Milestones
✅ Version 1: Basic CLI Notes (Done)

Add, list, and delete notes from command line.

Store notes in local JSON file.

Learned: structs, slices, functions, JSON handling, basic CLI.



🔜 Version 2: Organized Notes & Export

Add timestamps to notes automatically.

Group notes by meeting title or category.

Add search & filter commands.

Export notes to .txt or .md.

Improve CLI commands (meetnote help, meetnote version).

Focus: structs with nested data, file I/O, searching/filtering in Go.



🔜 Version 3: Live Transcription (Offline-first)

Capture live audio transcription (local speech-to-text).

Save transcriptions directly into notes file.

Summarize meeting text into short bullet points.

Add support for multiple meeting sessions.

Focus: concurrency in Go, external libraries, text processing.



🔜 Version 4: Browser & App Integration

Chrome extension or desktop client to auto-capture Google Meet/Zoom/Jitsi audio.

Sync notes between CLI & browser/app.

Option to edit notes inside a small UI.

Focus: Go web servers, APIs, frontend-backend connection.



🔜 Version 5: Polished Personal Meeting Assistant

Full assistant: joins online meetings, records notes, creates summaries, organizes by date/meeting.

Option to export/share notes in multiple formats.

Add collaborative features (share notes with teammates).

Focus: scaling project structure, modular Go code, real-world open-source release.


cl

📚 Learning Path in Go

v1 → Data structures, JSON, file system basics.

v2 → Struct nesting, CLI commands, searching/filtering.

v3 → Concurrency, APIs, external packages, algorithms.

v4 → Web servers, APIs, integration with frontend/browser.

v5 → Large-scale project design, open-source workflow.