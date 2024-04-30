package NFVideo

import (
	"bytes"
	"fyne.io/fyne/v2"
	xWidget "fyne.io/x/fyne/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFFS"
	"golang.org/x/sys/windows"
	"image/gif"
	"log"
	"path/filepath"
	"runtime"
	"strings"
)

type GifPlayer struct {
	player          *xWidget.AnimatedGif
	LoopCount       int
	AudioPath       string
	AudioFileConfig NFFS.Configuration
}

func (g *GifPlayer) Start() {

	switch runtime.GOOS {
	case "windows":
		//Set the timer resolution to 1ms
		err := windows.TimeBeginPeriod(1)
		if err != nil {
			return
		}

	}

	//Check if the audio file exists
	_, err := NFFS.Stat(g.AudioPath, g.AudioFileConfig)
	if err == nil {
		audioBytes, err := NFFS.ReadFile(g.AudioPath, g.AudioFileConfig)
		if err != nil {
			return
		}
		go func() {
			err = AudioTrack.PlayAudioFromBytes(audioBytes, 1, 1, false)
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
	AudioTrack.PauseAudio()
	g.player.Stop()
	switch runtime.GOOS {
	case "windows":
		err := windows.TimeEndPeriod(1)
		if err != nil {
			return
		}
	}
}

func (g *GifPlayer) Player() *xWidget.AnimatedGif {
	return g.player
}

// NewGifPlayer creates a new gif player from the given path.
func NewGifPlayer(path string, fileConfig NFFS.Configuration) (*GifPlayer, error) {
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
	resource := fyne.NewStaticResource(filepath.Base(gifPath), gifBytes)
	animatedGif, err := xWidget.NewAnimatedGifFromResource(resource)
	if err != nil {
		return nil, err
	}

	player.player = animatedGif

	return player, nil
}
