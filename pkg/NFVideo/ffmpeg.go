package NFVideo

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"math"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var FfmpegPath string

type Video struct {
	Probe          FFProbeOutput
	BufferedImages []image.Image
	CurrentFrame,
	TotalFrames,
	BufferCount,
	BufferWhenRemaining,
	BufferedTo,
	RealFrames,
	RealSeconds int
	Paused,
	Buffering bool
	TargetFPS        float64
	framesToSkip     int
	skippedLastFrame bool
	LastFrame        time.Time
	mu               sync.RWMutex
}

func (v *Video) ParseFramesToBuffer() error {
	// Calculate the start time in seconds
	startFrame := v.BufferedTo
	endFrame := startFrame + v.BufferCount
	fileName := v.Probe.Format.Filename

	// FFmpeg command to capture BufferCount frames starting from BufferedTo
	cmd := exec.Command(FfmpegPath, "-i", fileName,
		"-vf", fmt.Sprintf("select='gte(n\\,%d)*lte(n\\,%d)'", startFrame, endFrame), "-vsync", "0", "-f", "image2pipe", "-c:v", "png", "-")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	internalBuffer := make([]image.Image, 0)

	//Get the rounded up 10% of the buffer count
	roundedBuffer := math.Ceil(float64(v.BufferCount) * 0.1)

	for i := 0; ; i++ {
		img, err := png.Decode(stdout)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break // Buffer has been fully parsed or the video has ended
			}
			return fmt.Errorf("failed to decode frame: %w", err)
		}

		internalBuffer = append(internalBuffer, img)
		//Mod the i by the rounded buffer
		if math.Mod(float64(i), roundedBuffer) == 0 {
			//Append the buffer to the buffered images
			v.mu.Lock()
			v.BufferedImages = append(v.BufferedImages, internalBuffer...)
			v.BufferedTo += len(internalBuffer)
			v.mu.Unlock()
			//Clear the internal buffer
			internalBuffer = make([]image.Image, 0)
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg command failed: %w", err)
	}

	// Lock the buffer and append the new frames
	if len(internalBuffer) > 0 {
		v.mu.Lock()
		v.BufferedImages = append(v.BufferedImages, internalBuffer...)
		v.BufferedTo += len(internalBuffer)
		v.mu.Unlock()
	}
	v.Buffering = false

	return nil
}

func (v *Video) NextFrame() (image.Image, bool) {
	if v.CurrentFrame == v.TotalFrames {
		v.Paused = true
		return nil, false
	}
	if len(v.BufferedImages) < v.BufferWhenRemaining {
		if !v.Buffering {
			v.Buffering = true
			go func() {
				err := v.ParseFramesToBuffer()
				if err != nil {
					log.Println("Error parsing frames to buffer: ", err)
				}
			}()
		}
	}
	if len(v.BufferedImages) == 0 {
		frames := v.RealFrames
		//Wait for 5 frame to elapse to see if new frames are added to the buffer
		for i := 0; i < 5; i++ {
			if len(v.BufferedImages) > 0 {
				break
			}
			//Sleep for the duration of the frame
			time.Sleep(1 * time.Second / time.Duration(frames))
		}
		return nil, false
	}

	v.mu.RLock()
	frameToRender := 0
	if v.framesToSkip > 0 && !v.skippedLastFrame {
		frameToRender++
		v.framesToSkip--
		v.skippedLastFrame = true
	} else if v.skippedLastFrame {
		v.skippedLastFrame = false
	}
	//Check if frame to render is less than the length of the buffered images
	if frameToRender >= len(v.BufferedImages) {
		v.mu.RUnlock()
		return nil, true
	}
	img := v.BufferedImages[frameToRender]
	v.BufferedImages = v.BufferedImages[frameToRender+1:]
	v.LastFrame = time.Now()
	v.CurrentFrame++
	v.mu.RUnlock()
	return img, frameToRender != 0
}

// SkipFrames skips a number of frames
func (v *Video) SkipFrames(frames int) {
	v.mu.Lock()
	v.framesToSkip += frames
	v.mu.Unlock()
}

func NewVideo(probe FFProbeOutput, bufferCount int, bufferWhenRemaining int) (*Video, error) {
	video := &Video{
		Probe:               probe,
		BufferCount:         bufferCount,
		BufferWhenRemaining: bufferWhenRemaining,
	}

	rFrames := probe.Streams[0].RFrameRate
	//Split on the / character
	rFrameSplit := strings.Split(rFrames, "/")
	//Get the first element of the split
	rFrameNum := rFrameSplit[0]
	//Convert the string to an integer
	realFrames, err := strconv.Atoi(rFrameNum)
	if err != nil {
		return nil, err
	}
	video.RealFrames = realFrames
	//Get the second element of the split
	rFrameSeconds := rFrameSplit[1]
	//Convert the string to an integer
	realSeconds, err := strconv.Atoi(rFrameSeconds)
	if err != nil {
		return nil, err
	}
	video.RealSeconds = realSeconds

	//Calculate the target fps as a float from the real frames and real seconds
	targetFPS := float64(realFrames) / float64(realSeconds)
	video.TargetFPS = targetFPS

	//Get the number of frames
	nbFrames, err := strconv.Atoi(probe.Streams[0].NbFrames)
	if err != nil {
		return nil, err
	}
	video.TotalFrames = nbFrames

	err = video.ParseFramesToBuffer()
	if err != nil {
		return nil, err
	}
	return video, nil
}

func _(file string, splitLength string) error {
	err := CheckBinaries()
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(file)
	if err != nil {
		return err
	}

	fileName := filepath.Base(file)
	//Remove the extension from the file name
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	extension := filepath.Ext(absPath)

	timeCode, err := getTimeCode(splitLength)
	outPutPath := filepath.Join(filepath.Dir(absPath), "NFSplit", fileName)
	outPutPath += "_NFSplit_%03d" + extension

	cmd := exec.Command(FfmpegPath, "-i", file, "-c", "copy", "-map", "0", "-segment_time", timeCode, "-reset_timestamps", "1", "-f", "segment", outPutPath)
	err = cmd.Run()
	if err != nil {
		log.Println("Could not split video: ", err)
		return err
	}
	return nil
}

func getTimeCode(code string) (string, error) {
	// Check if the string contains at most one of each h, m, and s
	match, _ := regexp.MatchString(`^([0-9]*h)?([0-9]*m)?([0-9]*s)?$`, strings.ToLower(code))
	if !match {
		return "", errors.New("malformed string")
	}

	hours := 0
	minutes := 0
	seconds := 0

	// Split on h or H
	code = strings.ToLower(code)
	if strings.Contains(code, "h") {
		hoursSplit := strings.Split(code, "h")
		hours, _ = strconv.Atoi(hoursSplit[0])
		code = hoursSplit[1]
	}

	// Split on m
	if strings.Contains(code, "m") {
		minutesSplit := strings.Split(code, "m")
		minutes, _ = strconv.Atoi(minutesSplit[0])
		code = minutesSplit[1]
	}

	// Split on s
	if strings.Contains(code, "s") {
		secondsSplit := strings.Split(code, "s")
		seconds, _ = strconv.Atoi(secondsSplit[0])
	}

	// Carry over values that are too high
	minutes += seconds / 60
	seconds = seconds % 60
	hours += minutes / 60
	minutes = minutes % 60

	// Format the timecode
	timecode := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	return timecode, nil
}

func ExtractFrames(absPath string, folder string, fps float64) error {
	cmd := exec.Command(FfmpegPath, "-i", absPath, "-vf", "fps="+fmt.Sprintf("%.2f", fps), folder+"/%d.jpg")
	err := cmd.Run()
	if err != nil {
		log.Println("Could not extract frames: ", err)
		return err
	}
	return nil
}

func CreateGifFromVideo(absPath string, outPath string, fps float64) error {
	cmd := exec.Command(FfmpegPath, "-i", absPath, "-vf", "fps="+fmt.Sprintf("%.2f", fps)+",scale=720:-1:flags=lanczos", "-c:v", "gif", outPath)
	err := cmd.Run()
	if err != nil {
		log.Println("Could not create gif from video: ", err)
		return err
	}
	return nil
}
