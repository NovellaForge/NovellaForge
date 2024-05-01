package NFVideo

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"log"
	"math"
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

func CreateGifFromExtractedFrames(absPath string, fps float64, loopCount int, outputPath string) error {
	//Calculate the delay before decoding so that we can adjust the fps accordingly
	delay := 100 / fps
	//Round the delay to the nearest integer
	delay = math.Round(delay)
	//Adjust the fps to match the rounded delay
	fps = 100 / delay

	log.Println("Creating gif from video")
	log.Println("Path: ", absPath)
	log.Println("FPS: ", fps)
	log.Println("Loop count: ", loopCount)
	log.Println("Output path: ", outputPath)

	cmd := exec.Command(FfmpegPath, "-i", absPath, "-vf", "fps="+fmt.Sprintf("%.2f", fps), "-f", "image2pipe", "-vcodec", "png", "-")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	log.Println("Starting command")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}
	log.Println("Command started")

	g := &gif.GIF{}

	reader := bufio.NewReader(stdout)

	for {
		img, err := png.Decode(reader)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return fmt.Errorf("failed to decode JPEG: %w", err)
		}
		log.Println("Decoded image")

		paletted := image.NewPaletted(img.Bounds(), palette.Plan9)
		draw.Draw(paletted, paletted.Rect, img, img.Bounds().Min, draw.Src)

		g.Image = append(g.Image, paletted)
		g.Delay = append(g.Delay, int(delay))
	}

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

func findClosestFPSAndDelay(targetFPS float64) (float64, float64) {
	minDiff := math.MaxFloat64
	bestFPS := targetFPS
	bestDelay := 100.0 / targetFPS

	for fps := 1.0; fps <= 60.0; fps++ {
		delay := 100.0 / fps
		delayRounded := math.Round(delay)
		if delay == delayRounded {
			diff := math.Abs(targetFPS - fps)
			if diff < minDiff {
				minDiff = diff
				bestFPS = fps
				bestDelay = delay
			}
		}
	}

	return bestFPS, bestDelay
}

func CreateGifFromVideo(absPath string, fps float64, scale scale) (*gif.GIF, error) {
	//Calculate the delay before decoding so that we can adjust the fps accordingly
	fps, delay := findClosestFPSAndDelay(fps)
	log.Println("Creating gif from video")
	log.Println("Path: ", absPath)
	log.Println("FPS: ", fps)
	log.Println("Delay: ", delay)
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
