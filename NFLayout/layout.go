package NFLayout

import (
	"NovellaForge/NFWidget"
	"errors"
	"fyne.io/fyne/v2"
)

// This is the type of function that will be used to handle layouts
type layoutHandler func(map[string]interface{}, []*NFWidget.Widget) (fyne.CanvasObject, error)

type Layout struct {
	Type       string                 `json:"Type"`
	Children   []*NFWidget.Widget     `json:"Widgets"`
	Properties map[string]interface{} `json:"Properties"`
}

var defaultLayouts = map[string]layoutHandler{
	//TODO Create the basic fyne layouts
	"VBoxLayout": VBoxLayoutHandler,
}

var customLayouts = map[string]layoutHandler{}

func LayoutParser(window fyne.Window, layout Layout) (fyne.CanvasObject, error) {
	//Add the window to the layout properties
	layout.Properties["window"] = window
	//Take the layout and check if its name exists in the default layouts, if it does, use the assigned handlers, if not, check if it has a custom handlers, if not, return nil
	if handler, ok := defaultLayouts[layout.Type]; ok {
		return handler(layout.Properties, layout.Children)
	} else if handler, ok = customLayouts[layout.Type]; ok {
		return handler(layout.Properties, layout.Children)
	} else {
		return nil, errors.New("layout unable to be parsed")
	}
}
