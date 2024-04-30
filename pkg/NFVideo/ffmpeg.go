package NFVideo

import (
	"fmt"
	"image/gif"
	"log"
	"os"
	"os/exec"
)

var FfmpegPath string

type scale string

const (
	Scale240  scale = "240"
	Scale360  scale = "360"
	Scale720  scale = "720"
	Scale1080 scale = "1080"
)

func (s scale) String() string {
	//Check if the scale is a valid scale
	switch s {
	case Scale240, Scale360, Scale720, Scale1080:
		break
	default:
		return Scale720.String()
	}
	return string(s)
}

func CreateGifFromVideo(absPath string, fps float64, scale scale) (*gif.GIF, error) {
	log.Println("Creating gif from video")
	log.Println("Path: ", absPath)
	log.Println("FPS: ", fps)
	log.Println("Scale: ", scale)
	cmd := exec.Command(FfmpegPath, "-i", absPath, "-vf", "fps="+fmt.Sprintf("%.2f", fps)+",scale="+scale.String()+":-1:flags=lanczos", "-f", "gif", "-c:v", "gif", "-y", "pipe:1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	g, err := gif.DecodeAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to decode gif: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("ffmpeg command failed: %w", err)
	}

	return g, nil
}

func ExtractAudioToMP3(videoPath string, outputPath string) error {
	// Create the ffmpeg command to extract audio and save as MP3
	cmd := exec.Command(FfmpegPath, "-i", videoPath, "-vn", "-ab", "192k", "-ar", "48000", "-y", outputPath)

	// Run the command and capture any error
	err := cmd.Run()
	if err != nil {
		log.Println("Could not extract audio: ", err)
		return err
	}

	return nil
}

func SaveGifWithLoopCount(g *gif.GIF, loopCount int, outputPath string) error {
	g.LoopCount = loopCount

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, g)
	if err != nil {
		return fmt.Errorf("failed to encode gif: %w", err)
	}

	return nil
}
