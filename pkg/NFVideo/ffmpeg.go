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
	"strconv"
	"strings"
	"sync"
	"time"
)

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
	cmd := exec.Command(ffmpegPath, "-i", fileName,
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
		log.Println("Buffer is empty")
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
	//Sleep for the duration of the frame
	frames := float64(v.RealFrames)
	seconds := time.Duration(v.RealSeconds)
	//Calculate the target frame duration and adjust it based on the time since the last frame
	targetDuration := seconds * time.Second / time.Duration(frames)
	timeSinceLastFrame := time.Since(v.LastFrame)
	if timeSinceLastFrame < targetDuration {
		time.Sleep(targetDuration - timeSinceLastFrame)
	}
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
