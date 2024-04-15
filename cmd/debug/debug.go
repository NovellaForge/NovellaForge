package main

import (
	"fyne.io/fyne/v2/app"
	"go.novellaforge.dev/novellaforge/internal/NFEditor/EditorWidgets"
)

func main() {
	a := app.New()
	w := a.NewWindow("Hello")

	keyEntry := EditorWidgets.NewKeyEntry("Test")
	w.SetContent(keyEntry)
	w.ShowAndRun()
}
