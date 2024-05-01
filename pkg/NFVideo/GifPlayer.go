package NFVideo

import (
	"bytes"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFFS"
	"golang.org/x/sys/windows"
	"image"
	"image/draw"
	"image/gif"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

//TODO finish the implementation of the gif player

// ModifiedAnimatedGif is a modified version of the AnimatedGif widget from the fyne x package.
// This version will allow for frame handlers and precise timing.
// TODO Need to add in the ability to pause resume and reset the gif back to the beginning
type ModifiedAnimatedGif struct {
	widget.BaseWidget
	min fyne.Size

	src               *gif.GIF
	dst               *canvas.Image
	noDisposeIndex    int
	remaining         int
	stopping, running bool
	runLock           sync.RWMutex
	preciseTiming     bool
	frameEvents       map[int][]func()
	paused            bool
	currentFrame      int
}

func (g *ModifiedAnimatedGif) AddFrameHandler(frame int, handler func()) {
	if g.frameEvents == nil {
		g.frameEvents = make(map[int][]func())
	}
	if _, ok := g.frameEvents[frame]; !ok {
		g.frameEvents[frame] = []func(){handler}
	}
	g.frameEvents[frame] = append(g.frameEvents[frame], handler)
}

func (g *ModifiedAnimatedGif) CreateRenderer() fyne.WidgetRenderer {
	return &gifRenderer{gif: g}
}

func (g *ModifiedAnimatedGif) MinSize() fyne.Size {
	return g.min
}

func (g *ModifiedAnimatedGif) SetMinSize(size fyne.Size) {
	g.min = size
}

func (g *ModifiedAnimatedGif) draw(dst draw.Image, index int) {
	defer g.dst.Refresh()
	if index == 0 {
		// first frame
		draw.Draw(dst, g.dst.Image.Bounds(), g.src.Image[index], image.Point{}, draw.Src)
		g.dst.Image = dst
		g.noDisposeIndex = -1
		return
	}

	switch g.src.Disposal[index-1] {
	case gif.DisposalNone:
		// Do not dispose old frame, draw new frame over old
		draw.Draw(dst, g.dst.Image.Bounds(), g.src.Image[index], image.Point{}, draw.Over)
		// will be used in case of disposalPrevious
		g.noDisposeIndex = index - 1
	case gif.DisposalBackground:
		// clear with background then render new frame Over it
		// replacing entirely with new frame should achieve this?
		draw.Draw(dst, g.dst.Image.Bounds(), g.src.Image[index], image.Point{}, draw.Src)
	case gif.DisposalPrevious:
		// restore frame with previous image then render new over it
		if g.noDisposeIndex >= 0 {
			draw.Draw(dst, g.dst.Image.Bounds(), g.src.Image[g.noDisposeIndex], image.Point{}, draw.Src)
			draw.Draw(dst, g.dst.Image.Bounds(), g.src.Image[index], image.Point{}, draw.Over)
		} else {
			// there was no previous graphic, render background instead?
			draw.Draw(dst, g.dst.Image.Bounds(), g.src.Image[index], image.Point{}, draw.Src)
		}
	default:
		// Disposal = Unspecified/Reserved, simply draw new frame over previous
		draw.Draw(dst, g.dst.Image.Bounds(), g.src.Image[index], image.Point{}, draw.Over)
	}
}

func (g *ModifiedAnimatedGif) LoadGif(inGif *gif.GIF) error {
	g.src = inGif
	g.dst.Image = inGif.Image[0]
	g.dst.Refresh()
	return nil
}

// Start begins the animation. The speed of the transition is controlled by the loaded gif file.
func (g *ModifiedAnimatedGif) Start() {
	if g.isRunning() {
		return
	}
	g.runLock.Lock()
	g.running = true
	g.runLock.Unlock()

	buffer := image.NewNRGBA(g.dst.Image.Bounds())
	g.draw(buffer, 0)

	go func() {
		//if on windows set the timer resolution to 1ms
		switch runtime.GOOS {
		case "windows":
			err := windows.TimeBeginPeriod(1)
			if err != nil {
				return
			}
			defer func() {
				err := windows.TimeEndPeriod(1)
				if err != nil {
					return
				}
			}()
		}

		switch g.src.LoopCount {
		case -1: // don't loop
			g.remaining = 1
		case 0: // loop forever
			g.remaining = -1
		default:
			g.remaining = g.src.LoopCount + 1
		}
	loop:
		for g.remaining != 0 {
			lastFrameTime := time.Now()
			for c := g.currentFrame; c < len(g.src.Image); c++ {
				if g.isStopping() {
					break loop
				}

				frameStartTime := time.Now()

				g.draw(buffer, c)
				g.currentFrame = c
				go g.HandleFrame(c)
				frameProcessingTime := time.Since(frameStartTime)
				delay := (time.Duration(g.src.Delay[c]) * time.Millisecond * 10) - frameProcessingTime
				if time.Since(lastFrameTime) <= delay {
					if g.preciseTiming {
						//Busy wait until the delay is over
						for time.Since(lastFrameTime) <= delay {
						}
					} else {
						time.Sleep(delay - time.Since(lastFrameTime))
					}

				}
				lastFrameTime = time.Now()
			}
			if g.remaining > -1 { // don't underflow int
				g.remaining--
			}
		}
		g.runLock.Lock()
		g.running = false
		g.stopping = false
		g.runLock.Unlock()
	}()
}

func (g *ModifiedAnimatedGif) Pause() {
	g.runLock.Lock()
	g.paused = true
	g.runLock.Unlock()
}

func (g *ModifiedAnimatedGif) Resume() {
	g.runLock.Lock()
	g.paused = false
	g.runLock.Unlock()
	g.Start()
}

func (g *ModifiedAnimatedGif) Reset() {
	g.runLock.Lock()
	g.currentFrame = 0
	g.runLock.Unlock()
	g.Stop()
}

func (g *ModifiedAnimatedGif) isStopping() bool {
	g.runLock.RLock()
	defer g.runLock.RUnlock()
	return g.stopping
}

func (g *ModifiedAnimatedGif) isRunning() bool {
	g.runLock.RLock()
	defer g.runLock.RUnlock()
	return g.running
}

// Stop will request that the animation stops running, the last frame will remain visible
func (g *ModifiedAnimatedGif) Stop() {
	if !g.isRunning() {
		return
	}
	g.runLock.Lock()
	g.stopping = true
	g.runLock.Unlock()
}

func (g *ModifiedAnimatedGif) HandleFrame(c int) {
	//Check if functions exist for the frame
	if g.frameEvents != nil {
		if _, ok := g.frameEvents[c]; ok {
			for _, handler := range g.frameEvents[c] {
				handler()
				//TODO if needed add in an additional arguments array to pass to the handler
				// I cannot come up with a valid use case for this at the moment but it will be easy to add
			}
		}
	}
}

func NewModifiedAnimatedGif(inGif *gif.GIF, usePreciseTiming bool) (*ModifiedAnimatedGif, error) {
	ret := &ModifiedAnimatedGif{
		preciseTiming: usePreciseTiming,
	}
	ret.ExtendBaseWidget(ret)
	ret.dst = &canvas.Image{}
	ret.dst.FillMode = canvas.ImageFillContain
	return ret, ret.LoadGif(inGif)
}

type gifRenderer struct {
	gif *ModifiedAnimatedGif
}

func (g *gifRenderer) Destroy() {
	g.gif.Stop()
}

func (g *gifRenderer) Layout(size fyne.Size) {
	g.gif.dst.Resize(size)
}

func (g *gifRenderer) MinSize() fyne.Size {
	return g.gif.MinSize()
}

func (g *gifRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{g.gif.dst}
}

func (g *gifRenderer) Refresh() {
	g.gif.dst.Refresh()
}

type GifPlayer struct {
	player          *ModifiedAnimatedGif
	LoopCount       int
	AudioPath       string
	AudioFileConfig NFFS.Configuration
}

func (g *GifPlayer) Start() {
	//Check if the audio file exists
	_, err := NFFS.Stat(g.AudioPath, g.AudioFileConfig)
	if err == nil {
		audioBytes, err := NFFS.ReadFile(g.AudioPath, g.AudioFileConfig)
		if err != nil {
			return
		}
		go func() {
			err = AudioTrack.PlayAudioFromBytes(audioBytes, 1, 1, g.LoopCount)
			if err != nil {
				return
			}
		}()
	} else {
		log.Println("Audio file not found")
	}
	g.player.Start()
}

func (g *GifPlayer) Stop() {
	AudioTrack.PauseAudio() //This needs to be a reset/clear
	g.player.Stop()
}

func (g *GifPlayer) Pause() {
	AudioTrack.PauseAudio()
	g.player.Pause()
}

func (g *GifPlayer) Resume() {
	AudioTrack.ResumeAudio()
	g.player.Resume()
}

func (g *GifPlayer) Restart() {
	AudioTrack.PauseAudio() //This needs to be a reset/clear
	// We need to restart audio from the beginning here
	g.player.Reset()
	g.player.Start()
}

func (g *GifPlayer) Player() *ModifiedAnimatedGif {
	return g.player
}

// AddFrameHandler adds a function to be called when a specific frame is reached
func (g *GifPlayer) AddFrameHandler(frame int, handler func()) {
	g.player.AddFrameHandler(frame, handler)
}

// NewGifPlayer creates a new gif player from the given path.
func NewGifPlayer(path string, usePreciseTiming bool, fileConfig NFFS.Configuration) (*GifPlayer, error) {
	//Remove the extension from the path
	path = strings.TrimSuffix(path, filepath.Ext(path))

	gifPath := path + ".gif"
	audioPath := path + ".mp3"

	player := &GifPlayer{
		AudioPath:       audioPath,
		AudioFileConfig: fileConfig,
	}

	//Check if the file exists
	_, err := NFFS.Stat(gifPath, fileConfig)
	if err != nil {
		return nil, err
	}

	gifBytes, err := NFFS.ReadFile(gifPath, fileConfig)
	if err != nil {
		return nil, err
	}

	byteReader := bytes.NewReader(gifBytes)
	newGif, err := gif.DecodeAll(byteReader)
	if err != nil {
		return nil, err
	}
	player.LoopCount = newGif.LoopCount
	animatedGif, err := NewModifiedAnimatedGif(newGif, usePreciseTiming)
	if err != nil {
		return nil, err
	}

	player.player = animatedGif

	return player, nil
}
