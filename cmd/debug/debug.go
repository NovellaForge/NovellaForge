package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFVideo"
	"log"
	"time"
)

func main() {
	a := app.NewWithID("dev.novellaforge.testing")
	w := a.NewWindow("Hello")
	tempWindow := a.NewWindow("Temp")

	video, err := NFVideo.ParseVideo("assets/videos/banan.mp4")
	if err != nil {
		log.Println(err)
	}

	imageObj := canvas.NewImageFromResource(theme.BrokenImageIcon())
	tempWindow.SetContent(imageObj)
	var readImageButton *widget.Button
	readImageButton = widget.NewButton("Play Video", func() {
		if video == nil {
			log.Println("Video is nil")
			return
		}
		video.Paused = false
		readImageButton.Disable()
		go func() {
			for i := 0; i < video.TotalFrames-video.CurrentFrame; i++ {
				if video.Paused {
					break
				}
				img := video.NextFrame()
				if img == nil {
					log.Println("Image is nil")
					return
				}
				imageObj = canvas.NewImageFromImage(img)
				tempWindow.SetContent(imageObj)
				//Sleep for the duration of the frame
				frames := video.RealFrames
				log.Println("Sleeping for ", 1*time.Second/time.Duration(frames))
				time.Sleep(1 * time.Second / time.Duration(frames))
			}
			readImageButton.Enable()
		}()
	})
	pauseButton := widget.NewButton("Pause Video", func() {
		video.Paused = true
	})

	vbox := container.NewVBox(readImageButton, pauseButton)

	w.SetContent(vbox)
	w.Resize(fyne.NewSize(400, 200))
	w.Show()
	tempWindow.Resize(fyne.NewSize(400, 200))
	tempWindow.Show()
	a.Run()

}
