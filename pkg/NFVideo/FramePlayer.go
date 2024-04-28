package NFVideo

import (
	"bytes"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xWidget "fyne.io/x/fyne/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFFS"
	"image"
	"image/color/palette"
	"image/gif"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

//Todo fix this up to better use the gif player

type FrameVideoRenderer struct {
	fv *FramePlayer
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

type FramePlayer struct {
	widget.BaseWidget
	*NFVideo
	player   *xWidget.AnimatedGif
	frames   []string
	fsConfig NFFS.Configuration
}

func (f *FramePlayer) CreateRenderer() fyne.WidgetRenderer {
	return &FrameVideoRenderer{fv: f}
}

func NewFramePlayer(path string, config NFFS.Configuration) (*FramePlayer, error) {
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

	newGif, err := xWidget.NewAnimatedGifFromResource(theme.BrokenImageIcon())

	f := &FramePlayer{
		player:   newGif,
		frames:   make([]string, 0),
		NFVideo:  nfVideo,
		fsConfig: config,
	}
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
	err = f.CreateGifFromFrames()
	if err != nil {
		return nil, err
	}
	f.ExtendBaseWidget(f)

	return f, nil
}

func (f *FramePlayer) CreateGifFromFrames() error {
	var gifFrames []*image.Paletted
	for _, frame := range f.frames {
		file, err := NFFS.Open(frame, f.fsConfig)
		if err != nil {
			log.Println("Could not open file")
			return err
		}
		img, _, err := image.Decode(file)
		if err != nil {
			log.Println("Could not decode image")
			return err
		}
		palettedImage := image.NewPaletted(img.Bounds(), palette.Plan9)
		gifFrames = append(gifFrames, palettedImage)
	}

	//Calculate the target frame duration
	targetFrameDuration := (time.Second / time.Duration(f.TargetFPS)).Milliseconds()

	//Create the gif
	newGif := &gif.GIF{}
	for _, frame := range gifFrames {
		newGif.Image = append(newGif.Image, frame)
		newGif.Delay = append(newGif.Delay, int(targetFrameDuration))
	}

	var buf bytes.Buffer
	err := gif.EncodeAll(&buf, newGif)
	if err != nil {
		log.Println("Could not encode gif")
		return err
	}
	data := buf.Bytes()
	resource := fyne.NewStaticResource("video.gif", data)
	err = f.player.LoadResource(resource)
	if err != nil {
		return err
	}
	f.Refresh()
	return nil
}

func (f *FramePlayer) Start() {
	f.player.Start()
}

func (f *FramePlayer) Stop() {
	f.player.Stop()
}
