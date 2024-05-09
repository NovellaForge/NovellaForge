package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFFS"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFLog"
	NFVideo2 "go.novellaforge.dev/novellaforge/pkg/NFData/NFVideo"
	"log"
	"time"
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
		err = NFVideo2.FormatVideo("data/assets/videos/test4.mp4", 30, -1, "720")
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
	stopWatchBinding := binding.NewString()
	err := stopWatchBinding.Set("0")
	if err != nil {
		return
	}
	stopWatch := widget.NewLabelWithData(stopWatchBinding)
	log.Println("Loading video")
	video, err := NFVideo2.NewGifPlayer("assets/videos/test4.gif", false, config)
	if err != nil {
		log.Println(err)
		return
	}
	startTime := time.Now()
	resetWatch := func() {
		startTime = time.Now()
	}
	go func() {
		for i := 0; i < 1000; i++ {
			select {
			case <-time.After(100 * time.Millisecond):
				elapsed := time.Since(startTime)
				err := stopWatchBinding.Set(elapsed.String())
				if err != nil {
					return
				}
			}
		}
		log.Println("Watch done")
	}()
	video.Player().SetMinSize(fyne.NewSize(500, 500))
	playButton := widget.NewButton("Play Video", func() {
		video.Start()
		resetWatch()
	})
	stopButton := widget.NewButton("Stop Video", func() {
		video.Stop()
	})

	vbox := container.NewVBox(playButton, stopButton)
	border := container.NewBorder(nil, nil, vbox, nil, container.NewVBox(stopWatch, video.Player()))
	w.SetContent(border)
}
