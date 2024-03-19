package NFLayout

import (
	"errors"
	"fyne.io/fyne/v2"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
)

// This is the type of function that will be used to handle layouts
type layoutHandler func(window fyne.Window, args map[string]interface{}, children []NFWidget.Widget) (fyne.CanvasObject, error)

type Layout struct {
	Type       string                 `json:"Type"`
	Children   []NFWidget.Widget      `json:"Widgets"`
	Properties map[string]interface{} `json:"Properties"`
}

var layouts = map[string]layoutHandler{}

func (layout *Layout) Parse(window fyne.Window) (fyne.CanvasObject, error) {
	//Take the layout and check if its name exists in the default layouts, if it does, use the assigned handlers, if not, check if it has a custom handlers, if not, return nil
	if handler, ok := layouts[layout.Type]; ok {
		return handler(window, layout.Properties, layout.Children)
	} else {
		return nil, errors.New("layout unable to be parsed")
	}
}

// Register registers a custom layout handler
// If the name is already registered, it will not be registered again
// View the DefaultLayouts package for an example of how to use this function
func Register(name string, handler layoutHandler) {
	//Check if the name is already registered, if it is, return
	if _, ok := layouts[name]; ok {
		return
	}
	layouts[name] = handler
}
