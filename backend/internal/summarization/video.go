package summarization

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ExtractAudioFromVideo extracts audio from video file using ffmpeg
func ExtractAudioFromVideo(videoPath string) (string, error) {
	// Check if ffmpeg is installed
	if !isFFmpegInstalled() {
		return "", fmt.Errorf("ffmpeg not found. Please install ffmpeg:\n" +
			"  Mac: brew install ffmpeg\n" +
			"  Ubuntu/Debian: sudo apt-get install ffmpeg\n" +
			"  Windows: Download from https://ffmpeg.org/download.html")
	}

	// Create output audio file path
	audioPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + "_audio.mp3"

	fmt.Printf("🎬 Extracting audio from video...\n")
	fmt.Printf("   Input: %s\n", videoPath)
	fmt.Printf("   Output: %s\n", audioPath)

	// Build ffmpeg command
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vn",
		"-acodec", "libmp3lame",
		"-q:a", "2",
		"-y",
		audioPath,
	)

	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %v\nOutput: %s", err, string(output))
	}

	// Verify output file exists
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return "", fmt.Errorf("audio extraction failed: output file not created")
	}

	// Get file size
	info, _ := os.Stat(audioPath)
	fmt.Printf("✅ Audio extracted successfully (size: %d bytes)\n", info.Size())

	return audioPath, nil
}

// isFFmpegInstalled checks if ffmpeg is available in the system
func isFFmpegInstalled() bool {
	cmd := exec.Command("ffmpeg", "-version")
	err := cmd.Run()
	return err == nil
}

// GetFFmpegInstallInstructions returns platform-specific installation instructions
func GetFFmpegInstallInstructions() string {
	switch runtime.GOOS {
	case "darwin": // macOS
		return "Install ffmpeg on Mac:\n  brew install ffmpeg"
	case "linux":
		return "Install ffmpeg on Linux:\n  sudo apt-get install ffmpeg  # Ubuntu/Debian\n  sudo yum install ffmpeg      # CentOS/RHEL"
	case "windows":
		return "Install ffmpeg on Windows:\n  1. Download from https://ffmpeg.org/download.html\n  2. Extract and add to PATH"
	default:
		return "Please install ffmpeg for your operating system"
	}
}