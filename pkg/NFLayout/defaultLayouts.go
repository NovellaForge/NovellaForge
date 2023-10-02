package NFLayout

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
)

func VBoxLayoutHandler(window fyne.Window, args map[string]interface{}, children []NFWidget.Widget) (fyne.CanvasObject, error) {
	vbox := container.NewVBox()
	for _, child := range children {
		widget, err := child.Parse(window)
		if err != nil {
			return nil, err
		}
		vbox.Add(widget)
	}
	return vbox, nil
}
