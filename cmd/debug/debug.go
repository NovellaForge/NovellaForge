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

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Reset", func() {
				initContent(w)
			}),
		),
	)

	if _, e := NFFS.Stat("assets/videos/test4.gif", NFFS.NewConfiguration(true)); e != nil {
		log.Println("Creating video")
		err = NFVideo.FormatVideo("data/assets/videos/test4.mp4", 30, -1, "720")
		if err != nil {
			log.Println("Failed to format video: ", err)
			return
		}
	}

	w.SetMainMenu(mainMenu)
	initContent(w)
	w.Resize(fyne.NewSize(400, 200))
	w.ShowAndRun()
}

func initContent(w fyne.Window) {
	config := NFFS.NewConfiguration(true)
	config.OnlyLocal = true
	log.Println("Loading video")
	video, err := NFVideo.NewGifPlayer("assets/videos/test4.gif", false, config)
	if err != nil {
		log.Println(err)
		return
	}
	playButton := widget.NewButton("Play Video", func() {
		video.Start()
	})
	stopButton := widget.NewButton("Stop Video", func() {
		video.Stop()
	})

	vbox := container.NewVBox(playButton, stopButton)
	border := container.NewBorder(nil, nil, vbox, nil, video.Player())
	w.SetContent(border)
}
