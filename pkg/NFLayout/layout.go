package NFLayout

import (
	"errors"
	"fyne.io/fyne/v2"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
)

// This is the type of function that will be used to handle layouts
type layoutHandler func(window fyne.Window, args map[string]interface{}, children []NFWidget.Widget) (fyne.CanvasObject, error)

type Layout struct {
	Type       string                 `json:"Type"`
	Children   []NFWidget.Widget      `json:"Widgets"`
	Properties map[string]interface{} `json:"Properties"`
}

var defaultLayouts = map[string]layoutHandler{
	"VBox":   VBoxLayoutHandler,
	"HBox":   HBoxLayoutHandler,
	"Grid":   GridLayoutHandler,
	"Tab":    TabLayoutHandler,
	"Border": BorderLayoutHandler,
}

var customLayouts = map[string]layoutHandler{}

func (layout *Layout) Parse(window fyne.Window) (fyne.CanvasObject, error) {
	//Take the layout and check if its name exists in the default layouts, if it does, use the assigned handlers, if not, check if it has a custom handlers, if not, return nil
	if handler, ok := defaultLayouts[layout.Type]; ok {
		return handler(window, layout.Properties, layout.Children)
	} else if handler, ok = customLayouts[layout.Type]; ok {
		return handler(window, layout.Properties, layout.Children)
	} else {
		return nil, errors.New("layout unable to be parsed")
	}
}

// Register registers a custom layout handler
func Register(name string, handler layoutHandler) {
	//Check if the name is already registered, if it is, return
	if _, ok := customLayouts[name]; ok {
		return
	}
	customLayouts[name] = handler
}
