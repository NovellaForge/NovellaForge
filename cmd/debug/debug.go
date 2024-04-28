package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFFS"
	"go.novellaforge.dev/novellaforge/pkg/NFLog"
	"go.novellaforge.dev/novellaforge/pkg/NFVideo"
	"log"
)

func main() {
	a := app.NewWithID("dev.novellaforge.testing")
	w := a.NewWindow("Hello")
	err := NFLog.SetUp(w)
	if err != nil {
		log.Println(err)
		return
	}

	/*err = NFVideo.ParseVideoIntoGif("data/assets/videos/jpg.mp4", 30)
	if err != nil {
		log.Println(err)
		return
	}*/

	/*err = NFVideo.ParseVideoIntoFrames("data/assets/videos/jpg.mp4", 30)
	if err != nil {
		log.Println(err)
	}*/

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Reset", func() {
				initContent(w)
			}),
		),
	)

	w.SetMainMenu(mainMenu)
	initContent(w)
	w.Resize(fyne.NewSize(400, 200))
	w.ShowAndRun()
}

func initContent(w fyne.Window) {
	config := NFFS.NewConfiguration(true)
	config.OnlyLocal = true
	log.Println("Loading video")

	/*video, err := NFVideo.NewGifPlayer("assets/videos/jpg.gif", config)
	if err != nil {
		log.Println(err)
		return
	}*/

	/*video, err := Experimental.NewFileVideo("data/assets/videos/jpg.mp4", 60, 30)
	if err != nil {
		log.Println(err)
		return
	}*/

	video, err := NFVideo.NewFramePlayer("assets/videos/jpg", config)
	if err != nil {
		log.Println(err)
		return
	}

	playButton := widget.NewButton("Play Video", func() {
		video.Start()
	})
	pauseButton := widget.NewButton("Pause Video", func() {
		video.Stop()
	})

	vbox := container.NewVBox(playButton, pauseButton)
	border := container.NewBorder(nil, nil, vbox, nil, video)
	w.SetContent(border)
}
