package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFVideo"
	"log"
)

func main() {
	a := app.NewWithID("dev.novellaforge.testing")
	w := a.NewWindow("Hello")
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
	video, err := NFVideo.NewVideoWidget("assets/videos/testVid.mp4", 60, 30)
	if err != nil {
		log.Println(err)
	}

	probe := video.GetProbe()
	log.Println(probe.Streams[0].Width, probe.Streams[0].Height)

	playButton := widget.NewButton("Play Video", func() {
		video.Play()
	})
	pauseButton := widget.NewButton("Pause Video", func() {
		video.Pause()
	})

	vbox := container.NewVBox(playButton, pauseButton)
	border := container.NewBorder(nil, nil, vbox, nil, video)
	w.SetContent(border)
}
