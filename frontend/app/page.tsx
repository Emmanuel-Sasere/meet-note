
import { 
  Mic, 
  Square, 
  FileText, 
  BarChart3,
  Download,
  RefreshCw 
} from 'lucide-react';

export default function Home() {
  // RENDER DASHBOARD

   return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="container mx-auto px-4 py-8">
        
        {/* HEADER */}
        <div className="bg-white rounded-xl shadow-lg p-6 mb-8">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-bold text-gray-800 flex items-center gap-2">
                🎙️ Meet Note v3
              </h1>
              
              <p className="text-gray-600">Live Meeting Transcription</p>
            </div>
            
            <div className="flex items-center gap-4">
              {/* Connection Status */}
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-green-500 animate-pulse-slow"></div>
                <span className="text-sm text-gray-600">Connected</span>
              </div>
              
              {/* Recording Status */}
              <div className="flex items-center gap-2 bg-red-50 text-red-600 px-3 py-2 rounded-lg border border-red-200">
                <div className="w-2 h-2 bg-red-500 rounded-full animate-recording-pulse"></div>
                <span className="font-medium">Recording</span>
                <span className="font-mono">00:45</span>
              </div>
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
              
              {/* START RECORDING PANEL (static example) */}
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Meeting Title
                  </label>
                  <input
                    type="text"
                    placeholder="e.g., Team Standup, Q4 Planning"
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
                
                <button
                  className="w-full bg-blue-500 hover:bg-blue-600 text-white font-semibold py-4 px-6 rounded-lg transition-colors flex items-center justify-center gap-2"
                >
                  <Mic size={20} />
                  Start Recording
                </button>
              </div>
            </div>

            {/* RECENT SESSIONS */}
            <div className="bg-white rounded-xl shadow-lg p-6">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-xl font-semibold flex items-center gap-2">
                  📚 Recent Sessions
                </h2>
                <button className="text-gray-500 hover:text-gray-700 p-2 rounded-lg hover:bg-gray-100 transition-colors">
                  <RefreshCw size={16} />
                </button>
              </div>

              <div className="space-y-3">
                {/* Example Session 1 */}
                <div className="border border-gray-200 rounded-lg p-4 hover:border-blue-300 transition-colors">
                  <div className="flex justify-between items-start mb-2">
                    <h3 className="font-medium text-gray-800">Team Standup</h3>
                    <span className="px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-700">
                      completed
                    </span>
                  </div>
                  
                  <div className="flex gap-4 text-sm text-gray-600 mb-3">
                    <span>2025-09-19 10:00</span>
                    <span>1,230 words</span>
                  </div>

                  <div className="mb-3">
                    <p className="text-sm font-medium text-gray-700 mb-1">Key Points:</p>
                    <ul className="text-sm text-gray-600 space-y-1">
                      <li>• Project updates</li>
                      <li>• Deadline discussion</li>
                    </ul>
                  </div>

                  <button className="text-sm text-blue-600 hover:text-blue-700 flex items-center gap-1">
                    <Download size={14} />
                    Export
                  </button>
                </div>

                {/* Example Session 2 */}
                <div className="border border-gray-200 rounded-lg p-4 hover:border-blue-300 transition-colors">
                  <div className="flex justify-between items-start mb-2">
                    <h3 className="font-medium text-gray-800">Q4 Planning</h3>
                    <span className="px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-700">
                      active
                    </span>
                  </div>
                  
                  <div className="flex gap-4 text-sm text-gray-600 mb-3">
                    <span>2025-09-18 15:30</span>
                    <span>980 words</span>
                  </div>

                  <div className="mb-3">
                    <p className="text-sm font-medium text-gray-700 mb-1">Key Points:</p>
                    <ul className="text-sm text-gray-600 space-y-1">
                      <li>• Budget approval</li>
                      <li>• Hiring plans</li>
                    </ul>
                  </div>

                  <button className="text-sm text-blue-600 hover:text-blue-700 flex items-center gap-1">
                    <Download size={14} />
                    Export
                  </button>
                </div>
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
                  <div className="text-2xl font-bold text-blue-600">12</div>
                  <div className="text-sm text-blue-700">Total Sessions</div>
                </div>
                
                <div className="bg-green-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-green-600">14,560</div>
                  <div className="text-sm text-green-700">Words Transcribed</div>
                </div>
                
                <div className="bg-purple-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-purple-600">1,213</div>
                  <div className="text-sm text-purple-700">Avg Words/Session</div>
                </div>
              </div>
            </div>

            {/* QUICK ACTIONS */}
            <div className="bg-white rounded-xl shadow-lg p-6">
              <h2 className="text-xl font-semibold mb-6">⚡ Quick Actions</h2>
              
              <div className="space-y-3">
                <button className="w-full text-left p-3 rounded-lg border border-gray-200 hover:border-blue-300 hover:bg-blue-50 transition-colors flex items-center gap-3">
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
