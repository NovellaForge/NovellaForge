package Expirmental

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFVideo"
	"golang.org/x/sys/windows"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

//TODO: remove game dependency on ffmpeg, by generating png frames and parsing them at a set framerate
// For game build this will first create folders for each video, before splitting them into all their frames.
// It then needs to move the videos out of the video folder into a temp directory before checking embed and build/archiving the build
// It will then move the videos back before deleting the frame folders.

type VideoWidget struct {
	widget.BaseWidget
	video  *NFVideo.Video
	player *canvas.Image
}

func (v *VideoWidget) CreateRenderer() fyne.WidgetRenderer {
	return &NFVideo.VideoRenderer{}
}

func NewVideoWidget(file string, frameBuffer, bufferWhenRemaining int) (*VideoWidget, error) {
	//Check if the file exists
	_, err := os.Stat(file)
	if err != nil {
		return &VideoWidget{}, err
	}

	absFile, err := filepath.Abs(file)
	if err != nil {
		return &VideoWidget{}, err
	}

	//Check if the binaries are present
	err = NFVideo.CheckBinaries()
	if err != nil {
		return &VideoWidget{}, err
	}

	probe, err := NFVideo.ProbeVideo(absFile)
	if err != nil {
		return &VideoWidget{}, err
	}

	video, err := NFVideo.NewVideo(probe, frameBuffer, min(bufferWhenRemaining, frameBuffer))
	if err != nil {
		return &VideoWidget{}, err
	}

	firstFrame, _ := video.NextFrame()
	videoPlayer := &VideoWidget{
		video:  video,
		player: canvas.NewImageFromImage(firstFrame),
	}
	videoPlayer.player.FillMode = canvas.ImageFillContain
	videoPlayer.player.Refresh()
	videoPlayer.ExtendBaseWidget(videoPlayer)
	return videoPlayer, nil
}

func (v *VideoWidget) GetPlayer() *canvas.Image {
	return v.player
}

func (v *VideoWidget) Play() {
	totalFrames := v.video.TotalFrames
	currentFrame := v.video.CurrentFrame
	frames := float64(v.video.RealFrames)
	seconds := time.Duration(v.video.RealSeconds)
	v.video.LastFrame = time.Now()
	v.video.Paused = false
	frameRenderedChan := make(chan struct{}, 100)
	go func() {
		FPSTimer := time.NewTicker(5 * time.Second)
		TimeStart := time.Now()
		frameCount := 0
		for {
			select {
			case <-frameRenderedChan:
				frameCount++
			case <-FPSTimer.C:
				elapsed := time.Since(TimeStart)
				fps := float64(frameCount) / elapsed.Seconds()
				log.Println("FPS: ", fps)
				if fps < v.video.TargetFPS {
					//Subtract the fps from the target fps
					skippedFrames := math.Floor(v.video.TargetFPS - fps)
					if skippedFrames > 0 {
						log.Println("Video is lagging, skipping frames: ", skippedFrames)
						v.video.SkipFrames(int(skippedFrames))
					}
				}
				frameCount = 0
				TimeStart = time.Now()
			}
		}
	}()
	targetDuration := seconds * time.Second / time.Duration(frames)
	log.Println("Target FPS: ", v.video.TargetFPS)
	go func() {
		switch runtime.GOOS {
		case "windows":
			err := windows.TimeBeginPeriod(1)
			if err != nil {
				panic(err)
			}
			defer func() {
				err = windows.TimeEndPeriod(1)
				if err != nil {
					panic(err)
				}
			}()
		}
		for i := 0; i < totalFrames-currentFrame; i++ {
			if v.video.Paused {
				break
			}
			img, skipped := v.video.NextFrame()
			if skipped {
				frameRenderedChan <- struct{}{}
			}
			if img != nil {
				v.player.Image = img
				v.Refresh()
			}
			frameRenderedChan <- struct{}{}
			<-time.After(targetDuration)
		}
	}()
}

func (v *VideoWidget) CurrentFrame() int {
	return v.video.CurrentFrame
}

// GetProbe returns the probe output of the video
func (v *VideoWidget) GetProbe() NFVideo.FFProbeOutput {
	return v.video.Probe
}

func (v *VideoWidget) Pause() {
	v.video.Paused = true
}

func (v *VideoWidget) GetTargetFPS() float64 {
	return v.video.TargetFPS
}
