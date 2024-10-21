package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/davidroman0O/go-tempolite"
)

type Webm2Mp4 struct {
	InputFile  string
	OutputFile string
}

func Workflow(ctx tempolite.WorkflowContext[string], task Webm2Mp4) error {

	var isUrl bool
	if err := ctx.SideEffect("check url or not", func(ctx tempolite.SideEffectContext[string]) bool {
		return isURL(task.InputFile)
	}).Get(&isUrl); err != nil {
		return err
	}

	var intputFile string

	if isUrl {
		if err := ctx.Activity("downloadFile", DownloadFile, task).Get(&intputFile); err != nil {
			return err
		}
	} else {
		if err := ctx.Activity("checkDiskInputFile", CheckDiskInputFile, task).Get(&intputFile); err != nil {
			return err
		}
	}

	task.InputFile = intputFile

	if err := ctx.Activity("checkDiskOutputFile", CheckDiskOutputFile, task).Get(); err != nil {
		return err
	}

	defer fmt.Printf("Conversion completed: %s -> %s\n", task.InputFile, task.OutputFile)

	return ctx.Activity("transcoding", Transcoding, task).Get()
}

func DownloadFile(ctx tempolite.ActivityContext[string], task Webm2Mp4) (string, error) {
	url := task.InputFile
	// Create a temporary file for the download
	tempFile, err := os.CreateTemp("", "downloaded-*.webm")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: received status code %d", resp.StatusCode)
	}

	// Copy the response body to the file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write downloaded file: %v", err)
	}

	return tempFile.Name(), nil
}

func CheckDiskInputFile(ctx tempolite.ActivityContext[string], task Webm2Mp4) (string, error) {
	inputFile := task.InputFile
	// Check if the local file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Printf("Input file does not exist: %s\n", inputFile)
		return "", nil
	}
	return inputFile, nil
}

func CheckDiskOutputFile(ctx tempolite.ActivityContext[string], task Webm2Mp4) error {
	outputFile := task.OutputFile
	// Check if the output file already exists
	if _, err := os.Stat(outputFile); err == nil {
		// Output file exists, delete it
		fmt.Printf("Output file already exists, deleting: %s\n", outputFile)
		err = os.Remove(outputFile)
		if err != nil {
			fmt.Printf("Failed to delete existing output file: %v\n", err)
			return err
		}
	}
	return nil
}

func Transcoding(ctx tempolite.ActivityContext[string], task Webm2Mp4) error {

	timeout := time.NewTimer(time.Minute)

	select {
	case <-timeout.C:
		return fmt.Errorf("transcoding timed out")
	default:
		cmd := exec.Command("ffmpeg", "-i", task.InputFile, "-c:v", "libx264", "-crf", "23", "-preset", "medium", "-c:a", "aac", "-b:a", "192k", task.OutputFile)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error during conversion: %v\n", err)
			return err
		}
	}

	return nil
}

func isURL(input string) bool {
	return strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")
}

func main() {

	// Define command-line flags
	input := flag.String("input", "", "Path to the input file or URL (required)")
	output := flag.String("output", "", "Path to the output MP4 file (required)")
	flag.Parse()

	// Ensure both input and output are provided
	if *input == "" || *output == "" {
		fmt.Println("Usage: go run convert.go -input <input.webm or URL> -output <output.mp4>")
		return
	}

	ctx := context.Background()
	tp, err := tempolite.New[string](
		ctx,
		tempolite.NewRegistry[string]().
			Workflow(Workflow).
			Activity(Transcoding).
			Activity(CheckDiskOutputFile).
			Activity(CheckDiskInputFile).
			Activity(DownloadFile).
			Build(),
		tempolite.WithPath("./db/webm2mp4.db"),
	)

	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}

	if err := tp.Workflow("webm2mp4", Workflow, nil, Webm2Mp4{
		InputFile:  *input,
		OutputFile: *output,
	}).Get(); err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	if err = tp.Wait(); err != nil {
		log.Fatalf("Failed to wait for Tempolite instance: %v", err)
	}

	tp.Close()
}
