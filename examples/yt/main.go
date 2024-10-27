package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidroman0O/tempolite"
)

type YtDl struct {
	Url       string
	OutputDir string
	Timeout   string
}

func Workflow(ctx tempolite.WorkflowContext, task YtDl) error {
	defer fmt.Printf("Download completed: %s\n", task.Url)
	return ctx.Activity("ytdl", Download, task).Get()
}

func Download(ctx tempolite.ActivityContext, task YtDl) error {
	var timeoutChan <-chan time.Time

	// Parse the timeout string if it's not empty
	if task.Timeout != "" {
		duration, err := time.ParseDuration(task.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout duration: %v", err)
		}
		if duration > 0 {
			timer := time.NewTimer(duration)
			defer timer.Stop()
			timeoutChan = timer.C
		}
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(task.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create a channel for command completion
	done := make(chan error, 1)
	go func() {
		// Get video information including channel name
		infoCmd := exec.Command("yt-dlp",
			"--print", "%(channel)s",
			"--print", "%(title)s",
			"--no-playlist",
			task.Url,
		)

		output, err := infoCmd.Output()
		if err != nil {
			done <- fmt.Errorf("error getting video info: %v", err)
			return
		}

		// Split output into channel and title (they come on separate lines)
		parts := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(parts) != 2 {
			done <- fmt.Errorf("unexpected output format from yt-dlp")
			return
		}

		channelName := strings.TrimSpace(parts[0])
		videoTitle := strings.TrimSpace(parts[1])

		// Clean up channel name and title
		channelName = sanitizeFilename(channelName)
		videoTitle = sanitizeFilename(videoTitle)

		// Create filename in format: "ChannelName - VideoTitle.mp4"
		filename := fmt.Sprintf("%s - %s.mp4", channelName, videoTitle)
		outputPath := filepath.Join(task.OutputDir, filename)

		fmt.Printf("Downloading to: %s\n", outputPath)

		// Main download command
		cmd := exec.Command("yt-dlp",
			"-f", "bestvideo+bestaudio",
			task.Url,
			"-o", outputPath,
			"--no-playlist",
			"--merge-output-format", "mp4",
			"--progress",
		)

		// Get pipes for stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			done <- fmt.Errorf("error obtaining stdout pipe: %v", err)
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			done <- fmt.Errorf("error obtaining stderr pipe: %v", err)
			return
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			done <- fmt.Errorf("error starting download: %v", err)
			return
		}

		// Print stdout and stderr in real time
		go io.Copy(os.Stdout, stdout)
		go io.Copy(os.Stderr, stderr)

		// Wait for the command to complete
		if err := cmd.Wait(); err != nil {
			done <- fmt.Errorf("error during download: %v", err)
			return
		}

		fmt.Printf("Download completed successfully: %s\n", outputPath)
		done <- nil
	}()

	// Wait for either timeout or completion
	select {
	case <-timeoutChan:
		return fmt.Errorf("download timed out after %v", task.Timeout)
	case err := <-done:
		return err
	}
}

// sanitizeFilename removes or replaces invalid filename characters
func sanitizeFilename(name string) string {
	// Replace invalid characters with underscore
	invalid := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	result := name

	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Replace multiple spaces with single space
	result = strings.Join(strings.Fields(result), " ")

	// Trim spaces and dots from ends
	result = strings.Trim(result, " .")

	// Fallback if name becomes empty after cleaning
	if result == "" {
		return "unnamed"
	}

	return result
}

func main() {
	// Define command-line flags
	youtubeURL := flag.String("url", "", "YouTube video URL")
	outputDir := flag.String("output", "downloads", "Output directory path")
	timeout := flag.String("timeout", "", "Download timeout (e.g., 5m, 1h). Empty for no timeout")
	flag.Parse()

	// Ensure URL is provided
	if *youtubeURL == "" {
		fmt.Println("Usage: go run main.go -url <URL> [-output <directory>] [-timeout <duration>]")
		fmt.Println("Examples:")
		fmt.Println("  go run main.go -url \"https://youtube.com/watch?v=...\"")
		fmt.Println("  go run main.go -url \"https://youtube.com/watch?v=...\" -output \"./videos\" -timeout 5m")
		return
	}

	// Validate timeout format if provided
	if *timeout != "" {
		if _, err := time.ParseDuration(*timeout); err != nil {
			log.Fatalf("Invalid timeout format: %v", err)
		}
	}

	// Clean and validate output directory path
	cleanOutputDir := filepath.Clean(*outputDir)

	// Create Tempolite instance
	ctx := context.Background()
	tp, err := tempolite.New(
		ctx,
		tempolite.NewRegistry().
			Workflow(Workflow).
			Activity(Download).
			Build(),
		tempolite.WithPath("./db/yt.db"),
	)

	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	// Start workflow
	if err := tp.Workflow(Workflow, nil, YtDl{
		Url:       *youtubeURL,
		OutputDir: cleanOutputDir,
		Timeout:   *timeout,
	}).Get(); err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	if err = tp.Wait(); err != nil {
		log.Fatalf("Failed to wait for Tempolite instance: %v", err)
	}
}
