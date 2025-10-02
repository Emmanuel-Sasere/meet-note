'use client'
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
  RefreshCw 
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
    : 'http://localhost:8080/api';

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
      interval = setInterval(() => {
        setRecordingTime(prev => prev + 1);
      }, 1000);
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

  async function startRecording() {
    if (!meetingTitle.trim()) {
      alert('Please enter a meeting title');
      return;
    }

    try {
      const response = await fetch(`${API_BASE}/transcription/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: meetingTitle,
          language: 'en'
        })
      });

      const result = await response.json();
      
      if (result.success) {
        setCurrentSession(result.data);
        setIsRecording(true);
        setRecordingTime(0);
        setMeetingTitle('');
      } else {
        alert(`Failed to start: ${result.message}`);
      }
    } catch (error) {
      console.error('Start recording error:', error);
      alert('Failed to start recording. Check server connection.');
    }
  }

  async function stopRecording() {
    try {
      const response = await fetch(`${API_BASE}/transcription/stop`, {
        method: 'POST'
      });

      const result = await response.json();
      
      if (result.success) {
        setIsRecording(false);
        setCurrentSession(null);
        setRecordingTime(0);
        
        // Refresh sessions
        setTimeout(loadSessions, 2000);
      } else {
        alert(`Failed to stop: ${result.message}`);
      }
    } catch (error) {
      console.error('Stop recording error:', error);
      alert('Failed to stop recording');
    }
  }

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
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      onKeyPress={(e) => e.key === 'Enter' && startRecording()}
                    />
                  </div>
                  
                  <button
                    onClick={startRecording}
                    disabled={!isConnected}
                    className="w-full bg-blue-500 hover:bg-blue-600 disabled:bg-gray-300 text-white font-semibold py-4 px-6 rounded-lg transition-colors flex items-center justify-center gap-2"
                  >
                    <Mic size={20} />
                    Start Recording
                  </button>
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
                    onClick={stopRecording}
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

                      <button className="text-sm text-blue-600 hover:text-blue-700 flex items-center gap-1">
                        <Download size={14} />
                        Export
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
    </div>
  );
}

