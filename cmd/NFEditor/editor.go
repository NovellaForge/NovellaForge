package main

import (
	"bytes"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/data/assets"
)

var (
	Version     string // To be set by -ldflags on build
	BuildBranch string
	BuildAuthor string // the name of the repo is also valid
)

func main() {
	a := app.NewWithID("editor.novellaforge.dev")
	tosAccepted := a.Preferences().BoolWithFallback("TosAccepted", false)
	if !tosAccepted {
		// Create a confirmation dialog for TOS
		window := a.NewWindow("Terms of Service")

		tosFile, err := assets.LegalFS.Open("legal/TOS.md")
		if err != nil {
			panic(err)
		}
		defer tosFile.Close()
		//Get the data as bytes to put it into the static resource
		var tosBuf bytes.Buffer
		_, err = tosBuf.ReadFrom(tosFile)
		if err != nil {
			return
		}
		tosBytes := tosBuf.Bytes()
		tos := widget.NewRichTextFromMarkdown(string(tosBytes))
		tos.Wrapping = fyne.TextWrapWord
		tosContent := container.NewScroll(tos)
		confirm := widget.NewButton("Accept", func() {
			a.Preferences().SetBool("TosAccepted", true)
			launchApp(a)
			window.Close()
		})
		cancel := widget.NewButton("Decline", func() {
			window.Close()
			a.Quit()
		})

		buttons := container.NewHBox(confirm, cancel)
		content := container.NewBorder(nil, buttons, nil, nil, tosContent)

		window.SetContent(content)
		window.Resize(fyne.NewSize(400, 400)) // Set an appropriate size for the dialog
		window.Show()
	} else {
		launchApp(a) // If TOS is already accepted, directly launch the rest of the app
	}
	a.Run()
}

func createMainWindow(a fyne.App) fyne.Window {
	formattedTitle := "NovellaForge " + Version + " (" + BuildBranch + " - " + BuildAuthor + ")"
	mainWindow := a.NewWindow(formattedTitle)
	mainWindow.SetContent(container.NewVBox())
	mainWindow.SetMaster()
	winWidth := a.Preferences().FloatWithFallback("WindowWidth", 1280)
	winHeight := a.Preferences().FloatWithFallback("WindowHeight", 720)
	mainWindow.Resize(fyne.NewSize(float32(winWidth), float32(winHeight)))
	mainWindow.SetCloseIntercept(func() {
		size := mainWindow.Canvas().Size()
		//set the window size into preferences and then close it
		a.Preferences().SetFloat("WindowWidth", float64(size.Width))
		a.Preferences().SetFloat("WindowHeight", float64(size.Height))
		mainWindow.Close()
	})

	return mainWindow

}

func launchApp(a fyne.App) {
	w := createMainWindow(a)

	debugButton := widget.NewButton("Debug", func() {
		//Set the tos back to false
		a.Preferences().SetBool("TosAccepted", false)
	})

	w.SetContent(container.NewVBox(debugButton))
	w.Show()
}
