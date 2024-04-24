package NFVideo

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/assets"
	"io/fs"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var ffmpegPath string
var probePath string

func UnpackBinaries(location string) error {
	//Check for the binaries folder inside the location
	binaryFolder := filepath.Join(location, "binaries")

	_, err := os.Stat(binaryFolder)
	if os.IsNotExist(err) {
		err := os.MkdirAll(binaryFolder, os.ModePerm)
		if err != nil {
			return err
		}
	}

	//Check if the binaries folder is empty if it is not return an error
	//If it is empty, unpack the binaries from the embedded binaries folder

	binaryDir, err := os.ReadDir(binaryFolder)
	if err != nil {
		return err
	}

	if len(binaryDir) == 0 {
		switch runtime.GOOS {
		case "windows":
			err := fs.WalkDir(assets.BinaryFS, "binaries/win", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				data, err := assets.BinaryFS.ReadFile(path)
				if err != nil {
					return err
				}
				err = os.WriteFile(filepath.Join(binaryFolder, filepath.Base(path)), data, os.ModePerm)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
		default:
			return errors.New("unsupported OS")
		}
	}
	return nil
}

// CheckBinaries checks if the binaries are present in the location
func CheckBinaries() error {
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	defaultLocation := filepath.Join(userConfig, "NovellaForge", "ffmpeg")
	binaryLocation := fyne.CurrentApp().Preferences().StringWithFallback("ffmpegLocation", defaultLocation)

	//Check if the binaries are present
	err = UnpackBinaries(binaryLocation)
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		ffmpegPath = filepath.Join(binaryLocation, "binaries", "ffmpeg.exe")
		probePath = filepath.Join(binaryLocation, "binaries", "ffprobe.exe")
	default:
		log.Println("Unsupported OS")
		return errors.New("unsupported OS")
	}

	//Check if the ffmpeg and ffprobe binaries are present
	_, err = os.Stat(ffmpegPath)
	if err != nil {
		return errors.New("ffmpeg binary not found in the specified location")
	}
	_, err = os.Stat(probePath)
	if err != nil {
		return errors.New("ffprobe binary not found in the specified location")
	}

	//Check ffmpeg version
	cmd := exec.Command(ffmpegPath, "-version")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New("ffmpeg binary is invalid")
	}

	//Check ffprobe version
	cmd = exec.Command(probePath, "-version")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New("ffprobe binary is invalid")
	}
	return nil
}

type VideoRenderer struct {
	vw *VideoWidget
}

func (v VideoRenderer) Destroy() {}

func (v VideoRenderer) Layout(size fyne.Size) {
	//Resize the player
	v.vw.player.Resize(size)
}

func (v VideoRenderer) MinSize() fyne.Size {
	return v.vw.player.MinSize()
}

func (v VideoRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{v.vw.player}
}

func (v VideoRenderer) Refresh() {
	v.vw.player.Refresh()
}

type VideoWidget struct {
	widget.BaseWidget
	video  *Video
	player *canvas.Image
}

func (v *VideoWidget) CreateRenderer() fyne.WidgetRenderer {
	return &VideoRenderer{v}
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
	err = CheckBinaries()
	if err != nil {
		return &VideoWidget{}, err
	}

	probeArgs := []string{"-show_format", "-show_streams", "-print_format", "json", "-v", "quiet"}
	probe, err := ProbeVideo(absFile, probeArgs...)
	if err != nil {
		return &VideoWidget{}, err
	}

	video, err := NewVideo(probe, frameBuffer, min(bufferWhenRemaining, frameBuffer))
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

func (v *VideoWidget) Play() {
	totalFrames := v.video.TotalFrames
	currentFrame := v.video.CurrentFrame
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
						v.video.SkipFrames(int(skippedFrames))
					}
				}
				frameCount = 0
				TimeStart = time.Now()
			}
		}
	}()

	go func() {
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
		}
	}()
}

func (v *VideoWidget) CurrentFrame() int {
	return v.video.CurrentFrame
}

// GetProbe returns the probe output of the video
func (v *VideoWidget) GetProbe() FFProbeOutput {
	return v.video.Probe
}

func (v *VideoWidget) Pause() {
	v.video.Paused = true
}
