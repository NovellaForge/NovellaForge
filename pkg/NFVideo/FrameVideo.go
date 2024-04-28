package NFVideo

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/x/fyne/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFFS"
	"golang.org/x/sys/windows"
	"image"
	"io/fs"
	"log"
	"math"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FrameVideoRenderer struct {
	fv *FrameVideo
}

func (v *FrameVideoRenderer) Destroy() {}

func (v *FrameVideoRenderer) Layout(size fyne.Size) {
	if v.fv.player == nil {
		return
	}
	//Resize the player
	v.fv.player.Resize(size)
}

func (v *FrameVideoRenderer) MinSize() fyne.Size {
	if v.fv.player == nil {
		return fyne.NewSize(0, 0)
	}
	return v.fv.player.MinSize()
}

func (v *FrameVideoRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{v.fv.player}
}

func (v *FrameVideoRenderer) Refresh() {
	if v.fv.player == nil {
		return
	}
	v.fv.player.Refresh()
}

type NFVideo struct {
	TotalFrames int     `json:"totalFrames"`
	TargetFPS   float64 `json:"targetFPS"`
	RealFrames  int     `json:"realFrames"`
	RealSeconds int     `json:"realSeconds"`
	Path        string  `json:"path"`
}

func (v *NFVideo) Parse(file []byte) error {
	//Presume the file is json and unmarshal it
	return json.Unmarshal(file, v)
}

type FrameVideo struct {
	widget.BaseWidget
	*NFVideo
	player                   *canvas.Image
	currentFrame, bufferedTo int
	paused                   bool
	frameBuffer              []image.Image
	frames                   []string
	fsConfig                 NFFS.Configuration
	buffering                bool
	mu                       sync.RWMutex
}

func (f *FrameVideo) CreateRenderer() fyne.WidgetRenderer {
	return &FrameVideoRenderer{fv: f}
}

func NewFrameVideo(path string, config NFFS.Configuration) (*FrameVideo, error) {
	//Remove the extension from the path and check if a folder with a .NFVideo file exists
	//If it does, read the file and return the NFVideo struct
	nfVideo := &NFVideo{}

	//Trim the extension and check if the folder exists
	noExtPath := strings.TrimSuffix(path, filepath.Ext(path))
	fileName := filepath.Base(noExtPath)
	_, err := NFFS.Stat(noExtPath, config)
	if err != nil {
		log.Println("Video folder does not exist: " + noExtPath)
		return nil, err
	}
	//Check for a .nfvideo file in the folder
	nfVideoPath := filepath.Join(noExtPath, fileName+".NFVideo")
	_, err = NFFS.Stat(nfVideoPath, config)
	if err != nil {
		log.Println("NFVideo file does not exist: " + nfVideoPath)
		return nil, err
	}

	//Read the file and populate the NFVideo struct
	nfVideoFile, err := NFFS.ReadFile(nfVideoPath, config)
	if err != nil {
		return nil, err
	}

	//Parse the file
	err = nfVideo.Parse(nfVideoFile)

	f := &FrameVideo{
		player:       canvas.NewImageFromResource(theme.BrokenImageIcon()),
		frames:       make([]string, 0),
		currentFrame: 0,
		NFVideo:      nfVideo,
		paused:       true,
		fsConfig:     config,
	}
	f.player.FillMode = canvas.ImageFillContain

	framePath := filepath.Join(noExtPath, fileName+"_frames")

	err = NFFS.Walk(framePath, config, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		//Check if the file can be opened and decoded as an image
		file, err := NFFS.Open(path, config)
		if err != nil {
			log.Println("Could not open file: ", err)
			return nil
		}
		_, _, err = image.Decode(file)
		if err != nil {
			log.Println("Could not decode file as image: ", err)
			return nil
		}

		f.frames = append(f.frames, path)
		return nil
	})
	if err != nil {
		log.Println("Could not walk the frames folder: ", err)
		return nil, err
	}

	//Sort the frame slice based on the name of the file converted to an integer
	//This is to ensure that the frames are in order
	sort.Slice(f.frames, func(i, j int) bool {
		frame1, err := strconv.Atoi(strings.TrimSuffix(filepath.Base(f.frames[i]), filepath.Ext(f.frames[i])))
		if err != nil {
			log.Println("Could not convert frame to integer: ", err)
			return false
		}
		frame2, err := strconv.Atoi(strings.TrimSuffix(filepath.Base(f.frames[j]), filepath.Ext(f.frames[j])))
		if err != nil {
			log.Println("Could not convert frame to integer: ", err)
			return false
		}
		return frame1 < frame2
	})
	f.BufferFrames(2)
	f.NextFrame()
	f.ExtendBaseWidget(f)

	return f, nil
}

func (f *FrameVideo) BufferFrames(seconds float64) {
	//Multiply the seconds by the target FPS to get the number of frames to buffer
	f.mu.RLock()
	framesToBuffer := math.Ceil(seconds * f.TargetFPS)
	//Calculate the start and end frames to buffer
	bufferStart := f.bufferedTo
	bufferEnd := int(math.Min(float64(f.TotalFrames), float64(bufferStart)+framesToBuffer))
	copyFrames := make([]string, len(f.frames))
	copy(copyFrames, f.frames)
	copyConfig := f.fsConfig
	f.mu.RUnlock()
	log.Println("Buffering from: ", bufferStart, " to: ", bufferEnd)
	innerBuffer := make([]image.Image, 0)
	bufferedCount := 0
	brokenImage := canvas.NewImageFromResource(theme.BrokenImageIcon())

	//Start from the bufferedTo frame and buffer the frames
	for i := bufferStart; i <= bufferEnd; i++ {
		img := brokenImage.Image
		file, err := NFFS.Open(copyFrames[i], copyConfig)
		if err != nil {
			log.Println("Could not open file")
		} else {
			img, _, err = image.Decode(file)
			if err != nil {
				log.Println("Could not decode image")
			}
		}
		bufferedCount++
		innerBuffer = append(innerBuffer, img)
		if bufferedCount == 10 {
			f.mu.Lock()
			f.frameBuffer = append(f.frameBuffer, innerBuffer...)
			f.bufferedTo += len(innerBuffer)
			log.Println("Buffered to: ", f.bufferedTo)
			f.mu.Unlock()
			innerBuffer = make([]image.Image, 0)
			bufferedCount = 0
		}
	}
	//Append the remaining frames
	f.mu.Lock()
	if len(innerBuffer) > 0 {
		f.frameBuffer = append(f.frameBuffer, innerBuffer...)
		f.bufferedTo += len(innerBuffer)
	}
	f.buffering = false
	f.mu.Unlock()
}

func (f *FrameVideo) NextFrame() image.Image {
	f.mu.Lock()
	defer f.mu.Unlock()
	brokenImage := canvas.NewImageFromResource(theme.BrokenImageIcon()).Image

	if f.currentFrame == f.TotalFrames {
		f.Pause()
		return brokenImage
	}
	if float64(len(f.frameBuffer)) < f.TargetFPS {
		log.Println("Needs Buffering")
		if !f.buffering {
			f.buffering = true
			go func() {
				f.BufferFrames(2)
			}()
		}
		return brokenImage
	}
	//Return early if the buffer is empty
	if len(f.frameBuffer) == 0 {
		return brokenImage
	}
	img := f.frameBuffer[0]
	f.frameBuffer = f.frameBuffer[1:]
	f.currentFrame++
	return img
}

func (f *FrameVideo) Play() {
	brokenImage := canvas.NewImageFromResource(theme.BrokenImageIcon())
	brokenImage.FillMode = canvas.ImageFillContain
	f.paused = false
	//Make sure the frames slice is not empty
	if len(f.frames) == 0 {
		log.Println("No frames found")
		f.player = brokenImage
		f.player.Refresh()
		return
	}

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
				frameCount = 0
				TimeStart = time.Now()
			}
		}
	}()

	targetFrameDuration := time.Second / time.Duration(f.TargetFPS)

	switch runtime.GOOS {
	case "windows":
		log.Println("Setting timer resolution")
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

	go func() {
		for !f.paused {
			//Check if the current frame is less than the total frames
			if f.currentFrame < f.TotalFrames {
				startFrame := time.Now()
				f.player.Image = f.NextFrame()
				f.Refresh()
				endFrame := time.Now()
				frameRenderedChan <- struct{}{}
				//Calculate the time taken to render the frame
				frameDuration := endFrame.Sub(startFrame)
				//Check if the frame duration is less than the target frame duration
				if frameDuration < targetFrameDuration {
					time.Sleep(targetFrameDuration - frameDuration)
				}
			}

		}
	}()
}

func (f *FrameVideo) Pause() {
	f.paused = true
}

func (f *FrameVideo) CurrentFrame() int {
	return f.currentFrame
}
