package NFLayout

import (
	"NovellaForge/NFWidget"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func VBoxLayoutHandler(args map[string]interface{}, children []*NFWidget.Widget) (fyne.CanvasObject, error) {
	window := args["window"].(fyne.Window)
	vbox := container.NewVBox()
	for _, child := range children {
		widget, err := NFWidget.WidgetParser(window, child)
		if err != nil {
			return nil, err
		}
		vbox.Add(widget)
	}
	return vbox, nil
}
