"use client";
import { useState, useEffect } from 'react';
import { 
  Mic, 
  MicOff, 
  Play, 
  Square, 
  Clock, 
  FileText, 
  Users, 
  BarChart3,
  Download,
  RefreshCw, 
   X,
  Check 
} from 'lucide-react';

// TYPES (simplified from Go backend)
interface Session {
  id: string;
  title: string;
  start_time: string;
  duration: number;
  status: 'active' | 'completed';
  total_words: number;
  key_points: string[];
  action_items: string[];
}

interface SystemStats {
  total_sessions: number;
  total_words: number;
  avg_words_per_session: number;
}

export default function Dashboard() {
  // STATE MANAGEMENT
  const [isRecording, setIsRecording] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [currentSession, setCurrentSession] = useState<Session | null>(null);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [mediaRecorder, setMediaRecorder] = useState<MediaRecorder | null>(null);
  const [currentSessionId, setCurrentSessionId] = useState<string | null>(null);
  const [showExportModal, setShowExportModal] = useState(false);
const [selectedFormat, setSelectedFormat] = useState<"txt" | "pdf" | "docx">("txt");
const [isProcessing, setIsProcessing] = useState(false);
const [uploadedFile, setUploadedFile] = useState<File | null>(null);




  const [stats, setStats] = useState<SystemStats>({
    total_sessions: 0,
    total_words: 0,
    avg_words_per_session: 0
  });
  const [meetingTitle, setMeetingTitle] = useState('');
  const [recordingTime, setRecordingTime] = useState(0);

  // API BASE URL - connects to Go backend
  const API_BASE = process.env.NODE_ENV === 'production' 
    ? '/api' 
    : 'http://localhost:8080';

  // LOAD DATA ON PAGE START
  useEffect(() => {
    loadSystemStatus();
    loadSessions();
    
    // Auto-refresh every 10 seconds
    const interval = setInterval(() => {
      if (isRecording) {
        loadSystemStatus();
      }
    }, 10000);

    return () => clearInterval(interval);
  }, [isRecording]);

  // RECORDING TIMER
 useEffect(() => {
    let interval: NodeJS.Timeout;
    if (isRecording) {
      interval = setInterval(() => setRecordingTime((prev) => prev + 1), 1000);
    }
    return () => clearInterval(interval);
  }, [isRecording]);



  // API FUNCTIONS
  async function loadSystemStatus() {
    try {
      const response = await fetch(`${API_BASE}/status`);
      const result = await response.json();
      
      if (result.success) {
        setIsConnected(true);
        setStats(result.data.system_stats);
        
        // Check if recording is active
        if (result.data.is_recording && result.data.current_session) {
          setCurrentSession(result.data.current_session);
          setIsRecording(true);
        }
      }
    } catch (error) {
      console.error('Failed to load status:', error);
      setIsConnected(false);
    }
  }

  async function loadSessions() {
    try {
      const response = await fetch(`${API_BASE}/sessions`);
      const result = await response.json();
      
      if (result.success) {
        setSessions(result.data || []);
      }
    } catch (error) {
      console.error('Failed to load sessions:', error);
    }
  }

// 🎙️ START RECORDING
const handleStartRecording = async () => {
  if (!meetingTitle.trim()) {
    alert("Please enter a meeting title");
    return;
  }

  try {
    const displayStream = await navigator.mediaDevices.getDisplayMedia({
      video: true,
      audio: true,
    });
    console.log("🖥️ Got display stream:", displayStream.getTracks());

    const micStream = await navigator.mediaDevices.getUserMedia({ audio: true });
    console.log("🎤 Got mic stream:", micStream.getTracks());

    const combinedStream = new MediaStream([
      ...displayStream.getAudioTracks(),
      ...micStream.getAudioTracks(),
    ]);
    console.log("🎧 Combined stream tracks:", combinedStream.getTracks());

    const mimeType = MediaRecorder.isTypeSupported("audio/webm;codecs=opus")
      ? "audio/webm;codecs=opus"
      : "audio/webm";

    const recorder = new MediaRecorder(combinedStream, { mimeType });
    const chunks: Blob[] = []; // ✅ Local variable, not React state

    recorder.ondataavailable = (event) => {
      if (event.data.size > 0) {
        chunks.push(event.data);
        console.log("🎵 Captured chunk size:", event.data.size);
      }
    };

    recorder.onstop = async () => {
      console.log("🎤 Recorder stopped. Gathering chunks...");
      if (chunks.length === 0) {
        console.warn("⚠️ No audio chunks captured. Nothing to upload.");
        return;
      }

      const blob = new Blob(chunks, { type: "audio/webm;codecs=opus" });
      console.log(`🎧 Recorded blob type: ${blob.type}, size: ${blob.size} bytes`);

      const formData = new FormData();
      formData.append("file", blob, "meeting_audio.webm");

      try {
        console.log("⬆️ Uploading to backend...");
        const res = await fetch(`${API_BASE}/transcribe`, {
          method: "POST",
          body: formData,
        });

        const data = await res.json();
        console.log("🧠 Transcription result:", data);

if (data.text) {
  try {
    const res = await fetch(`${API_BASE}/sessions`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        title: meetingTitle || "Untitled Meeting",
        text: data.text,
        status: "completed",
      }),
    });

    const saved = await res.json();

    if (saved.success && saved.data?.id) {
      console.log("✅ Saved session ID:", saved.data.id);
      setCurrentSessionId(saved.data.id);
    } else {
      console.error("⚠️ Failed to get session ID:", saved);
    }

    await loadSessions();
    console.log("💾 Session saved successfully.");
    setIsProcessing(false);
  } catch (saveError) {
    console.error("Error saving session:", saveError);
  }
}


      } catch (err) {
        console.error("Error uploading audio:", err);
      }
    };

    recorder.start();
    setMediaRecorder(recorder);
    setIsRecording(true);
    setRecordingTime(0);
    console.log("✅ Recording started with", mimeType);
  } catch (error) {
    console.error("Error starting recording:", error);
    alert("Could not start recording. Please allow permissions.");
  }
};

// ⏹️ STOP RECORDING
const handleStopRecording = () => {
  if (mediaRecorder && isRecording) {
    console.log("⏹️ Stopping recorder...");
    setIsProcessing(true);
    mediaRecorder.stop();
    setIsRecording(false);
  }
};


// 💾 EXPORT TEXT
const handleExportSession = async (sessionId?: string) => {
  const id = sessionId || currentSessionId;
  if (!id) {
    alert("No session selected");
    return;
  }

  try {
    const res = await fetch(`${API_BASE}/session/${id}`);
    const data = await res.json();

    if (!data.success || !data.data?.text) {
      alert("Could not find session transcript");
      return;
    }

    const { title, text } = data.data;
    const fileName = `${title || "session"}.${selectedFormat}`;

    let blob;

    if (selectedFormat === "txt") {
      blob = new Blob([text], { type: "text/plain" });
    } else if (selectedFormat === "pdf") {
      // simple text-to-pdf generator
      const pdfContent = `
        ${title || "Session Transcript"}

        ${text}
      `;
      blob = new Blob([pdfContent], { type: "application/pdf" });
    } else if (selectedFormat === "docx") {
      const docxContent = `
        <html xmlns:o='urn:schemas-microsoft-com:office:office' 
              xmlns:w='urn:schemas-microsoft-com:office:word' 
              xmlns='http://www.w3.org/TR/REC-html40'>
          <head><meta charset='utf-8'><title>${title}</title></head>
          <body><h1>${title}</h1><p>${text}</p></body>
        </html>`;
      blob = new Blob([docxContent], { type: "application/vnd.openxmlformats-officedocument.wordprocessingml.document" });
    }

    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = fileName;
    a.click();
    URL.revokeObjectURL(url);

    setShowExportModal(false);
    console.log("📄 Exported:", fileName);
  } catch (err) {
    console.error("Export failed:", err);
  }
};





  // UTILITY FUNCTIONS
  function formatTime(seconds: number): string {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  }

  function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString() + ' ' + 
           new Date(dateString).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }


// 🎧 HANDLE AUDIO UPLOAD
const handleAudioUpload = async () => {
  if (!uploadedFile) {
    alert("Please select an audio file first!");
    return;
  }

  setIsProcessing(true);

  try {
    const formData = new FormData();
    formData.append("file", uploadedFile);

    console.log("⬆️ Uploading audio file:", uploadedFile.name);
    const res = await fetch(`${API_BASE}/transcribe`, {
      method: "POST",
      body: formData,
    });

    const data = await res.json();
    console.log("🧠 Transcription result:", data);

    if (data.text) {
      // Automatically summarize & store as a new session
      const saveRes = await fetch(`${API_BASE}/sessions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          title: uploadedFile.name.replace(/\.[^/.]+$/, "") || "Uploaded Audio",
          text: data.text,
          status: "completed",
        }),
      });

      const saved = await saveRes.json();
      if (saved.success) {
        console.log("✅ Uploaded session saved:", saved.data);
        await loadSessions();
        alert("✅ Transcription completed and saved!");
      } else {
        console.error("Failed to save uploaded session:", saved);
      }
    }
  } catch (error) {
    console.error("Upload error:", error);
    alert("Failed to upload or transcribe audio.");
  } finally {
    setIsProcessing(false);
    setUploadedFile(null);
  }
};


  // RENDER DASHBOARD
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="container mx-auto px-4 py-8">
        
        {/* HEADER */}
        <div className="bg-white rounded-xl shadow-lg p-6 mb-8">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-bold text-gray-800 flex items-center gap-2">
                🎙️ MeetNote v3
              </h1>
              <p className="text-gray-600">Live Meeting Transcription</p>
            </div>
            
            <div className="flex items-center gap-4">
              {/* Connection Status */}
              <div className="flex items-center gap-2">
                <div className={`w-3 h-3 rounded-full ${isConnected ? 'bg-green-500 animate-pulse-slow' : 'bg-red-500'}`}></div>
                <span className="text-sm text-gray-600">
                  {isConnected ? 'Connected' : 'Disconnected'}
                </span>
              </div>
              
              {/* Recording Status */}
              {isRecording && (
                <div className="flex items-center gap-2 bg-red-50 text-red-600 px-3 py-2 rounded-lg border border-red-200">
                  <div className="w-2 h-2 bg-red-500 rounded-full animate-recording-pulse"></div>
                  <span className="font-medium">Recording</span>
                  <span className="font-mono">{formatTime(recordingTime)}</span>
                </div>
              )}
            </div>
          </div>
        </div>
        {/* =================== */}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          
          {/* MAIN CONTROL PANEL */}
          <div className="lg:col-span-2">
            <div className="bg-white rounded-xl shadow-lg p-6 mb-8">
              <h2 className="text-xl font-semibold mb-6 flex items-center gap-2">
                🎛️ Transcription Control
              </h2>
              
              {!isRecording ? (
                // START RECORDING PANEL
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Meeting Title
                    </label>
                    <input
                      type="text"
                      value={meetingTitle}
                      onChange={(e) => setMeetingTitle(e.target.value)}
                      placeholder="e.g., Team Standup, Q4 Planning"
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-black"
                      onKeyPress={(e) => e.key === 'Enter' && handleStartRecording()}
                    />
                  </div>
                  
                  <button
                    onClick={handleStartRecording}
                    disabled={!isConnected}
                    className="w-full bg-blue-500 hover:bg-blue-600 disabled:bg-gray-300 text-white font-semibold py-4 px-6 rounded-lg transition-colors flex items-center justify-center gap-2"
                  >
                    <Mic size={20} />
                    Start Recording
                  </button>
                  <div className="mt-6 border-t pt-4">
  <p className="text-sm text-gray-500 text-center mb-3">
    Or upload an existing audio file
  </p>
  
  <div className="flex flex-col sm:flex-row gap-3 items-center">
    <input
      type="file"
      accept="audio/*"
      onChange={(e) => setUploadedFile(e.target.files?.[0] || null)}
      className="border border-gray-300 rounded-lg px-4 py-2 w-full text-sm"
    />
    <button
      onClick={handleAudioUpload}
      disabled={!uploadedFile || isProcessing}
      className="bg-indigo-500 hover:bg-indigo-600 disabled:bg-gray-300 text-white px-4 py-2 rounded-lg w-full sm:w-auto"
    >
      Upload & Transcribe
    </button>
  </div>
</div>

                </div>
              ) : (
                // STOP RECORDING PANEL
                <div className="space-y-6">
                  <div className="bg-blue-50 p-4 rounded-lg border border-blue-200">
                    <h3 className="font-semibold text-blue-800 mb-2">
                      {currentSession?.title}
                    </h3>
                    <div className="flex gap-4 text-sm text-blue-600">
                      <span>{currentSession?.total_words || 0} words</span>
                      <span>Active</span>
                    </div>
                  </div>
                  
                  <button
                    onClick={handleStopRecording}
                    className="w-full bg-red-500 hover:bg-red-600 text-white font-semibold py-4 px-6 rounded-lg transition-colors flex items-center justify-center gap-2"
                  >
                    <Square size={20} />
                    Stop Recording
                  </button>
                </div>
              )}
            </div>


            {/* RECENT SESSIONS */}
            <div className="bg-white rounded-xl shadow-lg p-6">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-xl font-semibold flex items-center gap-2">
                  📚 Recent Sessions
                </h2>
                <button
                  onClick={loadSessions}
                  className="text-gray-500 hover:text-gray-700 p-2 rounded-lg hover:bg-gray-100 transition-colors"
                >
                  <RefreshCw size={16} />
                </button>
              </div>

              <div className="space-y-3">
                {sessions.length === 0 ? (
                  <div className="text-center py-8 text-gray-500">
                    <FileText size={48} className="mx-auto mb-2 text-gray-300" />
                    <p>No sessions yet</p>
                    <p className="text-sm">Start your first recording above</p>
                  </div>
                ) : (
                  sessions.slice(0, 5).map((session) => (
                    <div key={session.id} className="border border-gray-200 rounded-lg p-4 hover:border-blue-300 transition-colors">
                      <div className="flex justify-between items-start mb-2">
                        <h3 className="font-medium text-gray-800">{session.title}</h3>
                        <span className={`px-2 py-1 rounded text-xs font-medium ${
                          session.status === 'active' 
                            ? 'bg-green-100 text-green-700' 
                            : 'bg-gray-100 text-gray-700'
                        }`}>
                          {session.status}
                        </span>
                      </div>
                      
                      <div className="flex gap-4 text-sm text-gray-600 mb-3">
                        <span>{formatDate(session.start_time)}</span>
                        <span>{session.total_words} words</span>
                      </div>

                      {session.key_points && session.key_points.length > 0 && (
                        <div className="mb-3">
                          <p className="text-sm font-medium text-gray-700 mb-1">Key Points:</p>
                          <ul className="text-sm text-gray-600 space-y-1">
                            {session.key_points.slice(0, 2).map((point, i) => (
                              <li key={i} className="truncate">• {point}</li>
                            ))}
                          </ul>
                        </div>
                      )}
                        {showExportModal && (
  <div className="fixed inset-0 flex items-center justify-center bg-black/60 backdrop-blur-sm z-50 p-4">
  <div className="bg-white rounded-2xl shadow-2xl w-full max-w-md transform transition-all">
    {/* Header */}
    <div className="px-6 py-5 border-b border-gray-100">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
            <Download className="w-5 h-5 text-blue-600" />
          </div>
          <div>
            <h2 className="text-xl font-semibold text-gray-900">Export Session</h2>
            <p className="text-sm text-gray-500">Choose your preferred format</p>
          </div>
        </div>
        <button
          onClick={() => setShowExportModal(false)}
          className="text-gray-400 hover:text-gray-600 transition-colors"
        >
          <X className="w-5 h-5" />
        </button>
      </div>
    </div>

    {/* Body */}
    <div className="px-6 py-6">
      <div className="space-y-3">
        {/* PDF Option */}
        <label
          className={`flex items-center gap-4 p-4 rounded-xl border-2 cursor-pointer transition-all ${
            selectedFormat === "pdf"
              ? "border-blue-500 bg-blue-50 shadow-sm"
              : "border-gray-200 hover:border-blue-300 hover:bg-gray-50"
          }`}
        >
          <input
            type="radio"
            name="format"
            value="pdf"
            checked={selectedFormat === "pdf"}
            onChange={() => setSelectedFormat("pdf")}
            className="w-5 h-5 text-blue-600 focus:ring-2 focus:ring-blue-500"
          />
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <FileText className="w-5 h-5 text-red-500" />
              <span className="font-medium text-gray-900">PDF Document</span>
            </div>
            <p className="text-sm text-gray-500 mt-1">
              Professional format, perfect for sharing
            </p>
          </div>
          {selectedFormat === "pdf" && (
            <div className="w-6 h-6 bg-blue-600 rounded-full flex items-center justify-center">
              <Check className="w-4 h-4 text-white" />
            </div>
          )}
        </label>

        {/* DOCX Option */}
        <label
          className={`flex items-center gap-4 p-4 rounded-xl border-2 cursor-pointer transition-all ${
            selectedFormat === "docx"
              ? "border-blue-500 bg-blue-50 shadow-sm"
              : "border-gray-200 hover:border-blue-300 hover:bg-gray-50"
          }`}
        >
          <input
            type="radio"
            name="format"
            value="docx"
            checked={selectedFormat === "docx"}
            onChange={() => setSelectedFormat("docx")}
            className="w-5 h-5 text-blue-600 focus:ring-2 focus:ring-blue-500"
          />
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <FileText className="w-5 h-5 text-blue-600" />
              <span className="font-medium text-gray-900">Word Document</span>
            </div>
            <p className="text-sm text-gray-500 mt-1">
              Editable format for Microsoft Word
            </p>
          </div>
          {selectedFormat === "docx" && (
            <div className="w-6 h-6 bg-blue-600 rounded-full flex items-center justify-center">
              <Check className="w-4 h-4 text-white" />
            </div>
          )}
        </label>

        {/* TXT Option */}
        <label
          className={`flex items-center gap-4 p-4 rounded-xl border-2 cursor-pointer transition-all ${
            selectedFormat === "txt"
              ? "border-blue-500 bg-blue-50 shadow-sm"
              : "border-gray-200 hover:border-blue-300 hover:bg-gray-50"
          }`}
        >
          <input
            type="radio"
            name="format"
            value="txt"
            checked={selectedFormat === "txt"}
            onChange={() => setSelectedFormat("txt")}
            className="w-5 h-5 text-blue-600 focus:ring-2 focus:ring-blue-500"
          />
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <FileText className="w-5 h-5 text-gray-600" />
              <span className="font-medium text-gray-900">Plain Text</span>
            </div>
            <p className="text-sm text-gray-500 mt-1">
              Simple text format, universal compatibility
            </p>
          </div>
          {selectedFormat === "txt" && (
            <div className="w-6 h-6 bg-blue-600 rounded-full flex items-center justify-center">
              <Check className="w-4 h-4 text-white" />
            </div>
          )}
        </label>
      </div>
    </div>

    {/* Footer */}
    <div className="px-6 py-4 bg-gray-50 rounded-b-2xl flex items-center justify-end gap-3">
      <button
        onClick={() => setShowExportModal(false)}
        className="px-5 py-2.5 text-gray-700 font-medium rounded-lg hover:bg-gray-200 transition-colors"
      >
        Cancel
      </button>
      <button
        onClick={() => handleExportSession()}
        className="px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg shadow-sm hover:shadow transition-all flex items-center gap-2"
      >
        <Download className="w-4 h-4" />
        Export
      </button>
    </div>
  </div>
</div>
)}

                      <button   onClick={() => setShowExportModal(true)}className="text-sm text-blue-600 hover:text-blue-700 flex items-center gap-1 cursor-pointer">
                        <Download size={14} />
                        Export Session
                      </button>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>

          {/* SIDEBAR - STATS */}
          <div className="space-y-8">
            
            {/* SYSTEM STATS */}
            <div className="bg-white rounded-xl shadow-lg p-6">
              <h2 className="text-xl font-semibold mb-6 flex items-center gap-2">
                📊 Statistics
              </h2>
              
              <div className="space-y-4">
                <div className="bg-blue-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-blue-600">{stats.total_sessions}</div>
                  <div className="text-sm text-blue-700">Total Sessions</div>
                </div>
                
                <div className="bg-green-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-green-600">
                    {stats.total_words.toLocaleString()}
                  </div>
                  <div className="text-sm text-green-700">Words Transcribed</div>
                </div>
                
                <div className="bg-purple-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-purple-600">
                    {Math.round(stats.avg_words_per_session || 0)}
                  </div>
                  <div className="text-sm text-purple-700">Avg Words/Session</div>
                </div>
              </div>
            </div>

            {/* QUICK ACTIONS */}
            <div className="bg-white rounded-xl shadow-lg p-6">
              <h2 className="text-xl font-semibold mb-6">⚡ Quick Actions</h2>
              
              <div className="space-y-3">
                <button 
                  onClick={loadSessions}
                  className="w-full text-left p-3 rounded-lg border border-gray-200 hover:border-blue-300 hover:bg-blue-50 transition-colors flex items-center gap-3"
                >
                  <RefreshCw size={16} className="text-gray-600" />
                  <span>Refresh Data</span>
                </button>
                
                <button className="w-full text-left p-3 rounded-lg border border-gray-200 hover:border-blue-300 hover:bg-blue-50 transition-colors flex items-center gap-3">
                  <BarChart3 size={16} className="text-gray-600" />
                  <span>View Analytics</span>
                </button>
                
                <button className="w-full text-left p-3 rounded-lg border border-gray-200 hover:border-blue-300 hover:bg-blue-50 transition-colors flex items-center gap-3">
                  <FileText size={16} className="text-gray-600" />
                  <span>Export All Sessions</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
      {isProcessing && (
  <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
    <div className="bg-white p-6 rounded-lg shadow-lg text-center">
      <div className="animate-spin rounded-full h-12 w-12 border-4 border-blue-500 border-t-transparent mx-auto mb-4"></div>
      <p className="text-lg font-medium text-gray-800">Processing transcription...</p>
      <p className="text-sm text-gray-500 mt-2">Please wait while we upload and transcribe your recording.</p>
    </div>
  </div>
)}

    </div>
  );
}

