package main

import (
	"fyne.io/fyne/v2/app"
	"go.novellaforge.dev/novellaforge/internal/NFEditor/EditorWidgets"
)

func main() {
	a := app.New()
	w := a.NewWindow("Widget Tester")
	testingEntry := EditorWidgets.NewTypedEntry(EditorWidgets.StringType)
	w.SetContent(testingEntry)
	w.ShowAndRun()

}
