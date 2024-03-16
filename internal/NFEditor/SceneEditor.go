package NFEditor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func CreateSceneEditor(window fyne.Window) fyne.CanvasObject {
	MainSplit := container.NewHSplit(
		container.NewVBox(widget.NewLabel("Scene Selector")),
		container.NewVSplit(container.NewVBox(widget.NewLabel("Scene Preview")),
			container.NewHSplit(container.NewVBox(widget.NewLabel("Scene Properties")),
				container.NewVBox(widget.NewLabel("Scene Objects")),
			),
		),
	)
	return MainSplit
}
