package NFVideo

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type Video struct {
	Probe          FFProbeOutput
	BufferedImages []*image.RGBA
	CurrentFrame,
	TotalFrames,
	BufferCount,
	BufferedTo,
	RealFrames,
	RealSeconds int
	Paused bool
}

func (v *Video) ParseFramesToBuffer() error {
	// Calculate the start time in seconds
	startTime := float64(v.BufferedTo) / float64(v.RealFrames)

	// FFmpeg command to capture BufferCount frames starting from BufferedTo
	cmd := exec.Command(ffmpegPath, "-ss", fmt.Sprintf("%.2f", startTime), "-i", v.Probe.Format.Filename,
		"-vframes", fmt.Sprintf("%d", v.BufferCount), "-f", "image2pipe", "-c:v", "png", "-")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	//Reinitialize the buffer
	v.BufferedImages = nil
	v.BufferedImages = make([]*image.RGBA, 0)
	for i := 0; i < v.BufferCount; i++ {
		img, err := png.Decode(stdout)
		if err != nil {
			if err == io.EOF {
				break // Less frames than expected
			}
			return fmt.Errorf("failed to decode frame: %w", err)
		}

		v.BufferedImages = append(v.BufferedImages, img.(*image.RGBA))
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg command failed: %w", err)
	}

	v.BufferedTo += v.BufferCount
	return nil
}

func (v *Video) NextFrame() *image.RGBA {
	if v.CurrentFrame == v.TotalFrames {
		return nil
	}

	if v.CurrentFrame == v.BufferedTo {
		err := v.ParseFramesToBuffer()
		if err != nil {
			return nil
		}
	}

	img := v.BufferedImages[v.CurrentFrame%v.BufferCount]
	v.CurrentFrame++
	return img
}

func NewVideo(probe FFProbeOutput, bufferCount int) (*Video, error) {
	video := &Video{
		Probe:       probe,
		BufferCount: bufferCount,
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
