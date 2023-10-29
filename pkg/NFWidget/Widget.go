package NFWidget

import (
	"fyne.io/fyne/v2"
	"github.com/NovellaForge/NovellaForge/pkg/NFError"
)

// Widget is the struct that holds all the information about a widget
type Widget struct {
	//ID is the unique ID of the widget for later reference in editing
	ID string `json:"ID"`
	//Type is the type of widget that is used to parse the widget this should be Globally Unique, so when making
	//custom ones prefix it with your package name like "MyPackage.MyWidget"
	Type string `json:"Type"`
	//Children is a list of widgets that are children of this widget
	Children []Widget `json:"Children"`
	//Properties is a map of properties that are used to parse the widget
	Properties map[string]interface{} `json:"Properties"`
	//JsonSafe is a json safe version of the widget being parsed to allow quick setting of properties; this is not required
	JsonSafe interface{} `json:"JsonSafe"`
}

type widgetHandler func(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error)

// Widgets is a map of all the widgets that are registered and can be used by the engine
var Widgets = map[string]widgetHandler{}

// Register adds a custom widget to the customWidgets map
func Register(name string, handler widgetHandler) {
	//Check if the name is already registered
	if _, ok := Widgets[name]; ok {

	}
	Widgets[name] = handler
}

func (w Widget) Parse(window fyne.Window) (fyne.CanvasObject, error) {
	if handler, ok := Widgets[w.Type]; ok {
		return handler(window, w.Properties, w.Children, w.JsonSafe)
	} else {
		return nil, NFError.ErrNotImplemented
	}
}
