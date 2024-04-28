package Experimental

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

type FileVideoRenderer struct {
	vp *FileVideo
}

func (v FileVideoRenderer) Destroy() {}

func (v FileVideoRenderer) Layout(size fyne.Size) {
	v.vp.player.Resize(size)
}

func (v FileVideoRenderer) MinSize() fyne.Size {
	return v.vp.player.MinSize()
}

func (v FileVideoRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{v.vp.player}
}

func (v FileVideoRenderer) Refresh() {
	v.vp.Refresh()
}

type FileVideo struct {
	widget.BaseWidget
	video  *NFVideo.Video
	player *canvas.Image
}

func (v *FileVideo) CreateRenderer() fyne.WidgetRenderer {
	return &FileVideoRenderer{vp: v}
}

func NewFileVideo(file string, frameBuffer, bufferWhenRemaining int) (*FileVideo, error) {
	//Check if the file exists
	_, err := os.Stat(file)
	if err != nil {
		return &FileVideo{}, err
	}

	absFile, err := filepath.Abs(file)
	if err != nil {
		return &FileVideo{}, err
	}

	//Check if the binaries are present
	err = NFVideo.CheckBinaries()
	if err != nil {
		return &FileVideo{}, err
	}

	probe, err := NFVideo.ProbeVideo(absFile)
	if err != nil {
		return &FileVideo{}, err
	}

	video, err := NFVideo.NewVideo(probe, frameBuffer, min(bufferWhenRemaining, frameBuffer))
	if err != nil {
		return &FileVideo{}, err
	}

	firstFrame, _ := video.NextFrame()
	videoPlayer := &FileVideo{
		video:  video,
		player: canvas.NewImageFromImage(firstFrame),
	}
	videoPlayer.player.FillMode = canvas.ImageFillContain
	videoPlayer.player.Refresh()
	videoPlayer.ExtendBaseWidget(videoPlayer)
	return videoPlayer, nil
}

func (v *FileVideo) GetPlayer() *canvas.Image {
	return v.player
}

func (v *FileVideo) Play() {
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

func (v *FileVideo) CurrentFrame() int {
	return v.video.CurrentFrame
}

// GetProbe returns the probe output of the video
func (v *FileVideo) GetProbe() NFVideo.FFProbeOutput {
	return v.video.Probe
}

func (v *FileVideo) Pause() {
	v.video.Paused = true
}

func (v *FileVideo) GetTargetFPS() float64 {
	return v.video.TargetFPS
}
