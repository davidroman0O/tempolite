package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Video struct {
	ID    string
	Title string
}

const (
	maxConcurrentDownloads = 3
	maxFileNameLength      = 200
)

func checkYtDlp() error {
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return fmt.Errorf("yt-dlp not found. Please install it first: brew install yt-dlp")
	}
	return nil
}

func getChannelInfo(channelURL string) (string, string, error) {
	cmdID := exec.Command("yt-dlp", "--skip-download", "--print", "channel_id", channelURL)
	channelID, err := cmdID.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get channel ID: %v", err)
	}

	cmdName := exec.Command("yt-dlp", "--skip-download", "--print", "channel", channelURL)
	channelName, err := cmdName.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get channel name: %v", err)
	}

	return strings.TrimSpace(string(channelID)), sanitizeFileName(strings.TrimSpace(string(channelName))), nil
}

func getVideoList(channelURL string) ([]Video, error) {
	// Use --flat-playlist to speed up video list retrieval
	cmd := exec.Command("yt-dlp", "--skip-download", "--print", "%(id)s,%(title)s", "--flat-playlist", channelURL)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get video list: %v", err)
	}

	var videos []Video
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ",", 2)
		if len(parts) == 2 {
			videos = append(videos, Video{
				ID:    parts[0],
				Title: parts[1],
			})
		}
	}
	return videos, nil
}

func sanitizeFileName(name string) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)

	if len(safe) > maxFileNameLength {
		safe = safe[:maxFileNameLength]
	}
	return safe
}

func checkSubtitlesAvailable(videoID string) bool {
	cmd := exec.Command("yt-dlp", "--skip-download", "--list-subs", fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	outputStr := string(output)
	return strings.Contains(outputStr, "has no subtitles") == false
}

func downloadSubtitles(video Video, outputDir string, wg *sync.WaitGroup, semaphore chan struct{}, stats *sync.Map) {
	defer wg.Done()
	defer func() { <-semaphore }()

	// Check if subtitles are available before attempting download
	if !checkSubtitlesAvailable(video.ID) {
		log.Printf("No subtitles available for: %s", video.Title)
		stats.Store(video.ID, "no_subtitles")
		return
	}

	log.Printf("Downloading subtitles for: %s", video.Title)

	safeTitle := sanitizeFileName(video.Title)
	outputTemplate := filepath.Join(outputDir, fmt.Sprintf("%.150s-%s.%%(ext)s", safeTitle, video.ID))

	cmd := exec.Command("yt-dlp",
		"--skip-download",
		"--write-sub",
		"--sub-lang", "en",
		"--convert-subs", "srt",
		"--output", outputTemplate,
		fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.ID),
	)

	if err := cmd.Run(); err != nil {
		log.Printf("Error downloading subtitles for %s: %v", video.Title, err)
		stats.Store(video.ID, "error")
		return
	}

	stats.Store(video.ID, "success")
	log.Printf("Successfully downloaded subtitles for: %s", video.Title)
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run main.go <channel_url>")
	}

	channelURL := os.Args[1]

	if err := checkYtDlp(); err != nil {
		log.Fatal(err)
	}

	channelID, channelName, err := getChannelInfo(channelURL)
	if err != nil {
		log.Fatal(err)
	}

	outputDir := fmt.Sprintf("subs_%s", channelName[:20]) // Limit directory name length
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	channelInfoPath := filepath.Join(outputDir, "channel_info.txt")
	channelInfo := fmt.Sprintf("Channel ID: %s\nChannel Name: %s\nURL: %s\n", channelID, channelName, channelURL)
	if err := os.WriteFile(channelInfoPath, []byte(channelInfo), 0644); err != nil {
		log.Printf("Warning: Failed to save channel info: %v", err)
	}

	log.Println("Fetching video list...")
	videos, err := getVideoList(channelURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found %d videos", len(videos))

	var stats sync.Map
	semaphore := make(chan struct{}, maxConcurrentDownloads)
	var wg sync.WaitGroup

	for _, video := range videos {
		wg.Add(1)
		semaphore <- struct{}{}
		go downloadSubtitles(video, outputDir, &wg, semaphore, &stats)
		time.Sleep(500 * time.Millisecond) // Reduced delay
	}

	wg.Wait()

	// Print summary
	var successful, noSubs, failed int
	stats.Range(func(key, value interface{}) bool {
		switch value.(string) {
		case "success":
			successful++
		case "no_subtitles":
			noSubs++
		case "error":
			failed++
		}
		return true
	})

	log.Printf("\nDownload Summary:")
	log.Printf("Total videos processed: %d", len(videos))
	log.Printf("Successfully downloaded: %d", successful)
	log.Printf("Videos without subtitles: %d", noSubs)
	log.Printf("Failed downloads: %d", failed)
	log.Printf("Subtitles are saved in: %s", outputDir)
}
