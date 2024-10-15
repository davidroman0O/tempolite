package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/davidroman0O/go-tempolite"
)

type YtDl struct {
	Url string
}

func Workflow(ctx tempolite.WorkflowContext[string], task YtDl) error {

	defer fmt.Printf("Conversion completed: %s \n", task.Url)

	return ctx.ActivityFunc("ytdl", Download, task).Get()
}

func Download(ctx tempolite.ActivityContext[string], task YtDl) error {

	timeout := time.NewTimer(time.Minute)

	select {
	case <-timeout.C:
		return fmt.Errorf("ytdl timed out")
	default:

		// yt-dlp command with format options for best video and best audio
		cmd := exec.Command("yt-dlp", "-f", "bestvideo+bestaudio", task.Url, "-o", "%(title)s.%(ext)s")

		// Get pipes for stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Printf("Error obtaining stdout pipe: %v\n", err)
			return err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			fmt.Printf("Error obtaining stderr pipe: %v\n", err)
			return err
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			fmt.Printf("Error starting command: %v\n", err)
			return err
		}

		// Print stdout and stderr in real time
		go io.Copy(os.Stdout, stdout)
		go io.Copy(os.Stderr, stderr)

		// Wait for the command to complete
		if err := cmd.Wait(); err != nil {
			fmt.Printf("Error during download: %v\n", err)
			return err
		}
	}

	return nil
}

func main() {

	// Define command-line flags
	youtubeURL := flag.String("url", "", "url youtube url")
	flag.Parse()

	// Ensure both input and output are provided
	if *youtubeURL == "" {
		fmt.Println("Usage: go run convert.go -url <URL>")
		return
	}

	ctx := context.Background()
	tp, err := tempolite.New[string](
		ctx,
		tempolite.WithPath("./db/yt.db"),
	)

	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}

	if err := tp.RegisterActivityFunc(Download); err != nil {
		log.Fatalf("Failed to register activity: %v", err)
	}

	if err := tp.RegisterWorkflow(Workflow); err != nil {
		log.Fatalf("Failed to register workflow: %v", err)
	}

	if err := tp.Workflow("ytdl", Workflow, YtDl{
		Url: *youtubeURL,
	}).Get(); err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	if err = tp.Wait(); err != nil {
		log.Fatalf("Failed to wait for Tempolite instance: %v", err)
	}

	tp.Close()
}
