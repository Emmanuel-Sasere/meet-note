"use client"

import { useState, useEffect, useRef } from 'react';
import { 
  Mic, 
  Square, 
  Upload,
  Download,
  FileText,
  Clock,
  AlertCircle,
  X,
  Check,
  Video
} from 'lucide-react';
import jsPDF from 'jspdf';

  // ✅ Add this helper at top of component
const checkIsMobile = () => {
  return typeof window !== 'undefined' && 
    /Android|iPhone|iPad|iPod/i.test(navigator.userAgent);
};

export default function MeetNote() {
  const [isRecording, setIsRecording] = useState(false);
  const [recordingTime, setRecordingTime] = useState(0);
  const [isProcessing, setIsProcessing] = useState(false);
  const [transcript, setTranscript] = useState('');
  const [summary, setSummary] = useState('');
  const [error, setError] = useState('');
  const [showWarning, setShowWarning] = useState(true);
  const [showExportModal, setShowExportModal] = useState(false);
  const [selectedFormat, setSelectedFormat] = useState<'txt' | 'pdf' | 'docx'>('txt');
  const [processingStep, setProcessingStep] = useState('');
  const [mobileWarning, setMobileWarning] = useState(false);
   const [mounted, setMounted] = useState(false);
   const [isMobileDevice, setIsMobileDevice] = useState(false);





  
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const API_URL = process.env.NEXT_PUBLIC_API_URL





    useEffect(() => {
    setMounted(true);
    setIsMobileDevice(checkIsMobile());
  }, []);

  // Prevent refresh when there's data
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (transcript || summary) {
        e.preventDefault();
        e.returnValue = 'You have unsaved transcripts. Are you sure you want to leave? Your data will be lost.';
        return e.returnValue;
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [transcript, summary]);

  // Recording timer
  useEffect(() => {
    let interval: NodeJS.Timeout;
    if (isRecording) {
      interval = setInterval(() => setRecordingTime((prev) => prev + 1), 1000);
    }
    return () => clearInterval(interval);
  }, [isRecording]);

  // Format time as MM:SS
  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  // Start recording
  const handleStartRecording = async () => {
    try {
       const isMobile = isMobileDevice;
     
      // declare streams in the outer scope so onstop can access them
      let displayStream: MediaStream | null = null;
      let micStream: MediaStream | null = null;
      let combinedStream: MediaStream | null = null;

       if (isMobile) {
      setMobileWarning(true);
        // Only use mic on mobile
        micStream = await navigator.mediaDevices.getUserMedia({ audio: true });
        combinedStream = micStream;
      } else {
        // Use screen + mic (desktop)
        displayStream = await navigator.mediaDevices.getDisplayMedia({
          video: true,
          audio: true,
        });
        micStream = await navigator.mediaDevices.getUserMedia({ audio: true });

        combinedStream = new MediaStream([
          ...displayStream.getAudioTracks(),
          ...micStream.getAudioTracks(),
        ]);
      }
if (!combinedStream) throw new Error("No media stream available");
      // const displayStream = await navigator.mediaDevices.getDisplayMedia({
      //   video: true,
      //   audio: true,
      // });

      // const micStream = await navigator.mediaDevices.getUserMedia({ audio: true });

      // const combinedStream = new MediaStream([
      //   ...displayStream.getAudioTracks(),
      //   ...micStream.getAudioTracks(),
      // ]);

      const mimeType = MediaRecorder.isTypeSupported("audio/webm;codecs=opus")
        ? "audio/webm;codecs=opus"
        : "audio/webm";

      const recorder = new MediaRecorder(combinedStream, { mimeType });
      chunksRef.current = [];

      recorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          chunksRef.current.push(event.data);
        }
      };

      recorder.onstop = async () => {
        const blob = new Blob(chunksRef.current, { type: "audio/webm" });
        await uploadAndTranscribe(blob, 'audio');
        
        displayStream?.getTracks().forEach(track => track.stop());
        micStream?.getTracks().forEach(track => track.stop());
      };

      recorder.start();
      mediaRecorderRef.current = recorder;
      setIsRecording(true);
      setRecordingTime(0);
      setError('');
    } catch (err) {
      console.error("Error starting recording:", err);
      setError("Could not start recording. Please allow permissions.");
    }
  };

  // Stop recording
  const handleStopRecording = () => {
    if (mediaRecorderRef.current && isRecording) {
      mediaRecorderRef.current.stop();
      setIsRecording(false);
    }
  };
  

  // Handle file upload (audio or video)
  const handleFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const fileType = file.type.startsWith('video/') ? 'video' : 'audio';
    await uploadAndTranscribe(file, fileType);
    event.target.value = '';
  };

  // Upload and transcribe audio/video
  const uploadAndTranscribe = async (fileBlob: Blob, fileType: 'audio' | 'video') => {
    setIsProcessing(true);
    setError('');
    setProcessingStep(fileType === 'video' ? 'Extracting audio from video...' : 'Uploading audio...');

    try {
      const formData = new FormData();
      formData.append("file", fileBlob, fileType === 'video' ? "video.mp4" : "audio.webm");
      formData.append("type", fileType);
     

      const response = await fetch(`${API_URL}/transcribe`, {
        method: "POST",
        body: formData,
      });

      const result = await response.json();

      if (result.success) {
        setTranscript(result.data.transcript);
        setSummary(result.data.summary);
        setProcessingStep('');
      } else {
        setError(result.error || "Processing failed");
        setProcessingStep('');
      }
    } catch (err) {
      console.error("Upload error:", err);
      setError("Failed to upload or process file. Make sure the server is running.");
      setProcessingStep('');
    } finally {
      setIsProcessing(false);
    }
  };

  // Download transcript and summary in selected format
  const handleDownload = () => {
    const timestamp = new Date().toISOString().slice(0, 10);
    const fileName = `notes-${timestamp}`;
    
    if (selectedFormat === 'txt') {
      downloadAsText(fileName);
    } else if (selectedFormat === 'pdf') {
      downloadAsPDF(fileName);
    } else if (selectedFormat === 'docx') {
      downloadAsDOCX(fileName);
    }
    
    setShowExportModal(false);
  };

  // Download as plain text
  const downloadAsText = (fileName: string) => {
    const content = `NOTES
${'='.repeat(80)}

${transcript}

${'='.repeat(80)}

SUMMARY
${'-'.repeat(80)}

${summary}

${'='.repeat(80)}
Generated on: ${new Date().toLocaleString()}`;
    
    const blob = new Blob([content], { type: 'text/plain' });
    downloadBlob(blob, `${fileName}.txt`);
  };


 // Download as PDF using jsPDF library
const downloadAsPDF = (fileName: string) => {
  const doc = new jsPDF();
  const pageWidth = doc.internal.pageSize.getWidth();
  const pageHeight = doc.internal.pageSize.getHeight();
  const margin = 20;
  const maxWidth = pageWidth - (margin * 2);
  let yPosition = margin;

  // Helper function to add text with wrapping
  const addWrappedText = (text: string, fontSize: number, isBold: boolean = false) => {
    doc.setFontSize(fontSize);
    if (isBold) {
      doc.setFont('helvetica', 'bold');
    } else {
      doc.setFont('helvetica', 'normal');
    }

    const lines = doc.splitTextToSize(text, maxWidth);
    
    lines.forEach((line: string) => {
      // Check if we need a new page
      if (yPosition > pageHeight - margin) {
        doc.addPage();
        yPosition = margin;
      }
      
      doc.text(line, margin, yPosition);
      yPosition += fontSize * 0.5; // Line spacing
    });
  };

  // Title
  doc.setFillColor(37, 99, 235); // Blue
  doc.rect(margin, yPosition, maxWidth, 15, 'F');
  doc.setTextColor(255, 255, 255); // White text
  doc.setFontSize(20);
  doc.setFont('helvetica', 'bold');
  doc.text('📋 Notes', margin + 5, yPosition + 10);
  yPosition += 25;

  // Reset text color
  doc.setTextColor(0, 0, 0);

  // Transcript Section
  doc.setFontSize(16);
  doc.setFont('helvetica', 'bold');
  doc.setTextColor(75, 85, 99); // Gray
  doc.text('Transcript', margin, yPosition);
  yPosition += 10;

  // Draw line under header
  doc.setDrawColor(229, 231, 235);
  doc.line(margin, yPosition, pageWidth - margin, yPosition);
  yPosition += 8;

  // Transcript content
  doc.setTextColor(0, 0, 0);
  addWrappedText(transcript, 10);
  yPosition += 15;

  // Summary Section
  doc.setFontSize(16);
  doc.setFont('helvetica', 'bold');
  doc.setTextColor(75, 85, 99);
  doc.text('Summary', margin, yPosition);
  yPosition += 10;

  // Draw line under header
  doc.line(margin, yPosition, pageWidth - margin, yPosition);
  yPosition += 8;

  // Summary content (with background box)
  const summaryStartY = yPosition;
  doc.setFillColor(239, 246, 255); // Light blue background
  
  // Calculate summary height
  doc.setFontSize(10);
  const summaryLines = doc.splitTextToSize(summary, maxWidth - 10);
  const summaryHeight = summaryLines.length * 5 + 10;
  
  doc.rect(margin, summaryStartY - 5, maxWidth, summaryHeight, 'F');
  
  // Add summary text
  doc.setTextColor(0, 0, 0);
  addWrappedText(summary, 10);
  yPosition += 15;

  // Footer
  if (yPosition > pageHeight - 30) {
    doc.addPage();
    yPosition = margin;
  }
  
  doc.setFontSize(8);
  doc.setTextColor(107, 114, 128); // Gray
  doc.setFont('helvetica', 'normal');
  doc.text(
    `Generated on: ${new Date().toLocaleString()} | Created with Noted`,
    pageWidth / 2,
    pageHeight - 15,
    { align: 'center' }
  );

  // Save the PDF
  doc.save(`${fileName}.pdf`);
};

  // Download as DOCX
  const downloadAsDOCX = (fileName: string) => {
    const htmlContent = `
<html xmlns:o='urn:schemas-microsoft-com:office:office' 
      xmlns:w='urn:schemas-microsoft-com:office:word' 
      xmlns='http://www.w3.org/TR/REC-html40'>
<head>
  <meta charset='utf-8'>
  <title>Notes</title>
  <style>
    body { font-family: 'Calibri', sans-serif; line-height: 1.6; padding: 40px; }
    h1 { color: #2563eb; border-bottom: 3px solid #2563eb; padding-bottom: 10px; }
    h2 { color: #4b5563; margin-top: 30px; border-bottom: 2px solid #e5e7eb; padding-bottom: 5px; }
    .notes { background-color: #f9fafb; padding: 20px; border: 1px solid #e5e7eb; margin: 20px 0; white-space: pre-wrap; }
    .summary { background-color: #eff6ff; padding: 20px; border-left: 4px solid #2563eb; margin: 20px 0; }
    .footer { margin-top: 40px; padding-top: 20px; border-top: 1px solid #e5e7eb; color: #6b7280; font-size: 0.9em; }
  </style>
</head>
<body>
  <h1>📋 Notes</h1>
  
  <h2>Transcript</h2>
  <div class="notes">${transcript.replace(/\n/g, '<br>')}</div>
  
  <h2>Summary</h2>
  <div class="summary">${summary.replace(/\n/g, '<br>')}</div>
  
  <div class="footer">
    Generated on: ${new Date().toLocaleString()}<br>
    Created with Noted
  </div>
</body>
</html>`;
    
    const blob = new Blob([htmlContent], { 
      type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' 
    });
    downloadBlob(blob, `${fileName}.docx`);
  };

  // Helper function to trigger download
  const downloadBlob = (blob: Blob, fileName: string) => {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = fileName;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  // Clear all data
  const handleClear = () => {
    if (confirm('Are you sure? This will delete your transcript and summary.')) {
      setTranscript('');
      setSummary('');
      setError('');
    }
  };

 
    return (
  
      <div className="min-h-screen bg-gradient-to-br from-indigo-50 via-purple-50 to-pink-50">
      <div className="container mx-auto px-4 py-8 max-w-6xl">
        
        {/* Warning Banner */}
        {showWarning && (
          <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4 mb-6 rounded-r-lg shadow-sm">
            <div className="flex items-start">
              <AlertCircle className="w-5 h-5 text-yellow-600 mt-0.5 mr-3 flex-shrink-0" />
              <div className="flex-1">
                <h3 className="text-sm font-semibold text-yellow-800 mb-1">
                  Data Not Saved Permanently
                </h3>
                <p className="text-sm text-yellow-700">
                  Your transcripts are stored in your browser only. If you refresh or close this page, all data will be lost. 
                  Make sure to download your notes before leaving!
                </p>
              </div>
              <button
                onClick={() => setShowWarning(false)}
                className="ml-3 text-yellow-600 hover:text-yellow-800"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
          </div>
        )}

        {/* Header */}
        <div className="bg-white rounded-2xl shadow-xl p-8 mb-8">
          <div className="text-center">
            <h1 className="text-4xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent mb-2">
              🎙️ Noted
            </h1>
            <p className="text-gray-600">
              Record or upload audio/video for instant transcription and summary
            </p>
          </div>
        </div>

        {/* Main Control Panel */}
        <div className="bg-white rounded-2xl shadow-xl p-8 mb-8">
          <h2 className="text-2xl font-semibold text-gray-800 mb-6 flex items-center gap-2">
            🎛️ Record or Upload
          </h2>

          {!isRecording ? (
            <div className="space-y-4">
              {/* Start Recording Button */}
              <button
                onClick={handleStartRecording}
                disabled={isProcessing}
                className="w-full bg-gradient-to-r from-red-500 to-pink-500 hover:from-red-600 hover:to-pink-600 disabled:from-gray-300 disabled:to-gray-400 text-white font-semibold py-4 px-6 rounded-xl transition-all transform hover:scale-105 disabled:scale-100 shadow-lg flex items-center justify-center gap-3"
              >
                <Mic className="w-6 h-6" />
                Start Recording Audio
              </button>
              {/* Mobile Warning */}
{mounted && mobileWarning && (
  <div className="mt-4 bg-blue-50 border border-blue-200 rounded-lg p-4">
    <p className="text-blue-700 text-sm">📱 Mobile detected: Recording audio only (screen recording not supported)</p>
  </div>
)}

              {/* Divider */}
              <div className="relative py-4">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-gray-300"></div>
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="px-4 bg-white text-gray-500 font-medium">OR UPLOAD FILES</span>
                </div>
              </div>

              {/* Upload Options */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {/* Upload Audio */}
                <label className="border-2 border-dashed border-indigo-300 hover:border-indigo-500 bg-indigo-50 hover:bg-indigo-100 rounded-xl py-6 px-6 transition-all cursor-pointer flex flex-col items-center gap-3">
                  <Upload className="w-8 h-8 text-indigo-600" />
                  <span className="text-gray-700 font-medium">Upload Audio</span>
                  <span className="text-sm text-gray-500">MP3, WAV, WEBM, etc.</span>
                  <input
                    type="file"
                    accept="audio/*"
                    onChange={handleFileUpload}
                    disabled={isProcessing}
                    className="hidden"
                  />
                </label>

                {/* Upload Video */}
                <label className="border-2 border-dashed border-purple-300 hover:border-purple-500 bg-purple-50 hover:bg-purple-100 rounded-xl py-6 px-6 transition-all cursor-pointer flex flex-col items-center gap-3">
                  <Video className="w-8 h-8 text-purple-600" />
                  <span className="text-gray-700 font-medium">Upload Video</span>
                  <span className="text-sm text-gray-500">MP4, MOV, AVI, etc.</span>
                  <input
                    type="file"
                    accept="video/*"
                    onChange={handleFileUpload}
                    disabled={isProcessing}
                    className="hidden"
                  />
                </label>
              </div>
            </div>
          ) : (
            // Recording in Progress
            <div className="space-y-6">
              <div className="bg-red-50 border-2 border-red-200 rounded-xl p-6 text-center">
                <div className="flex items-center justify-center gap-3 mb-4">
                  <div className="w-4 h-4 bg-red-500 rounded-full animate-pulse"></div>
                  <span className="text-2xl font-bold text-red-600">RECORDING</span>
                </div>
                <div className="flex items-center justify-center gap-2 text-gray-700">
                  <Clock className="w-5 h-5" />
                  <span className="text-3xl font-mono font-bold">{formatTime(recordingTime)}</span>
                </div>
              </div>

              <button
                onClick={handleStopRecording}
                className="w-full bg-gray-800 hover:bg-gray-900 text-white font-semibold py-4 px-6 rounded-xl transition-all shadow-lg flex items-center justify-center gap-3"
              >
                <Square className="w-6 h-6" />
                Stop Recording
              </button>
            </div>
          )}

          {/* Error Message */}
          {error && (
            <div className="mt-4 bg-red-50 border border-red-200 rounded-lg p-4">
              <p className="text-red-700 text-sm">{error}</p>
            </div>
          )}
        </div>

        {/* Processing Indicator */}
        {isProcessing && (
          <div className="bg-white rounded-2xl shadow-xl p-8 mb-8">
            <div className="flex flex-col items-center justify-center gap-4">
              <div className="w-16 h-16 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin"></div>
              <p className="text-lg font-medium text-gray-700">
                {processingStep || 'Processing your file...'}
              </p>
              <p className="text-sm text-gray-500">This may take a minute depending on file length</p>
            </div>
          </div>
        )}

        {/* Results Section */}
        {(transcript || summary) && !isProcessing && (
          <div className="space-y-6">
            {/* Action Buttons */}
            <div className="flex gap-4">
              <button
                onClick={() => setShowExportModal(true)}
                className="flex-1 bg-gradient-to-r from-green-500 to-emerald-500 hover:from-green-600 hover:to-emerald-600 text-white font-semibold py-3 px-6 rounded-xl transition-all shadow-lg flex items-center justify-center gap-2"
              >
                <Download className="w-5 h-5" />
                Download Notes
              </button>
              <button
                onClick={handleClear}
                className="flex-1 bg-gray-200 hover:bg-gray-300 text-gray-700 font-semibold py-3 px-6 rounded-xl transition-all flex items-center justify-center gap-2"
              >
                <X className="w-5 h-5" />
                Clear All
              </button>
            </div>

            {/* Transcript Card */}
            {transcript && (
              <div className="bg-white rounded-2xl shadow-xl p-8 border border-gray-200">
                <h3 className="text-2xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                  <FileText className="w-6 h-6" />
                  Notes
                </h3>
                <div className="bg-gray-50 rounded-xl p-6 max-h-96 overflow-y-auto">
                  <p className="text-gray-800 whitespace-pre-wrap leading-relaxed">
                    {transcript}
                  </p>
                </div>
              </div>
            )}

            {/* Summary Card */}
            {summary && (
              <div className="bg-gradient-to-br from-purple-50 to-pink-50 rounded-2xl shadow-xl p-8 border border-purple-100">
                <h3 className="text-2xl font-bold text-purple-900 mb-4 flex items-center gap-2">
                  📋 Summary
                </h3>
                <div className="prose max-w-none">
                  <p className="text-gray-800 whitespace-pre-wrap leading-relaxed">{summary}</p>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Empty State */}
        {!transcript && !summary && !isProcessing && (
          <div className="bg-white rounded-2xl shadow-xl p-12 text-center">
            <div className="text-6xl mb-4">🎤</div>
            <h3 className="text-xl font-semibold text-gray-700 mb-2">
              Ready to transcribe
            </h3>
            <p className="text-gray-500">
              Start recording or upload an audio/video file to get started
            </p>
          </div>
        )}
      </div>

      {/* Export Modal */}
      {showExportModal && (
        <div className="fixed inset-0 flex items-center justify-center bg-black/60 backdrop-blur-sm z-50 p-4">
          <div className="bg-white rounded-2xl shadow-2xl w-full max-w-md transform transition-all">
            <div className="px-6 py-5 border-b border-gray-100">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">
                    <Download className="w-5 h-5 text-green-600" />
                  </div>
                  <div>
                    <h2 className="text-xl font-semibold text-gray-900">Export Notes</h2>
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

            <div className="px-6 py-6">
              <div className="space-y-3">
                <label className={`flex items-center gap-4 p-4 rounded-xl border-2 cursor-pointer transition-all ${selectedFormat === "txt" ? "border-green-500 bg-green-50 shadow-sm" : "border-gray-200 hover:border-green-300 hover:bg-gray-50"}`}>
                  <input type="radio" name="format" value="txt" checked={selectedFormat === "txt"} onChange={() => setSelectedFormat("txt")} className="w-5 h-5 text-green-600 focus:ring-2 focus:ring-green-500" />
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <FileText className="w-5 h-5 text-gray-600" />
                      <span className="font-medium text-gray-900">Plain Text (.txt)</span>
                    </div>
                    <p className="text-sm text-gray-500 mt-1">Simple format, works everywhere</p>
                  </div>
                  {selectedFormat === "txt" && <div className="w-6 h-6 bg-green-600 rounded-full flex items-center justify-center"><Check className="w-4 h-4 text-white" /></div>}
                </label>

                <label className={`flex items-center gap-4 p-4 rounded-xl border-2 cursor-pointer transition-all ${selectedFormat === "pdf" ? "border-green-500 bg-green-50 shadow-sm" : "border-gray-200 hover:border-green-300 hover:bg-gray-50"}`}>
                  <input type="radio" name="format" value="pdf" checked={selectedFormat === "pdf"} onChange={() => setSelectedFormat("pdf")} className="w-5 h-5 text-green-600 focus:ring-2 focus:ring-green-500" />
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <FileText className="w-5 h-5 text-red-500" />
                      <span className="font-medium text-gray-900">PDF Document (.pdf)</span>
                    </div>
                    <p className="text-sm text-gray-500 mt-1">Professional format for sharing</p>
                  </div>
                  {selectedFormat === "pdf" && <div className="w-6 h-6 bg-green-600 rounded-full flex items-center justify-center"><Check className="w-4 h-4 text-white" /></div>}
                </label>

                <label className={`flex items-center gap-4 p-4 rounded-xl border-2 cursor-pointer transition-all ${selectedFormat === "docx" ? "border-green-500 bg-green-50 shadow-sm" : "border-gray-200 hover:border-green-300 hover:bg-gray-50"}`}>
                  <input type="radio" name="format" value="docx" checked={selectedFormat === "docx"} onChange={() => setSelectedFormat("docx")} className="w-5 h-5 text-green-600 focus:ring-2 focus:ring-green-500" />
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <FileText className="w-5 h-5 text-blue-600" />
                      <span className="font-medium text-gray-900">Word Document (.docx)</span>
                    </div>
                    <p className="text-sm text-gray-500 mt-1">Editable in Microsoft Word</p>
                  </div>
                  {selectedFormat === "docx" && <div className="w-6 h-6 bg-green-600 rounded-full flex items-center justify-center"><Check className="w-4 h-4 text-white" /></div>}
                </label>
              </div>
            </div>

            <div className="px-6 py-4 bg-gray-50 rounded-b-2xl flex items-center justify-end gap-3">
              <button onClick={() => setShowExportModal(false)} className="px-5 py-2.5 text-gray-700 font-medium rounded-lg hover:bg-gray-200 transition-colors">Cancel</button>
              <button onClick={handleDownload} className="px-5 py-2.5 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg shadow-sm hover:shadow transition-all flex items-center gap-2">
                <Download className="w-4 h-4" />
                Download
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  
  );
  
  
}