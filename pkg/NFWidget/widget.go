package NFWidget

import (
	"NovellaForge/pkg/NFError"
	"fyne.io/fyne/v2"
)

type widgetHandler func(window fyne.Window, args map[string]interface{}, children []*Widget) (fyne.CanvasObject, error)

var customWidgets = map[string]widgetHandler{}

type Widget struct {
	ID         string                 `json:"ID"`
	Type       string                 `json:"Type"`
	Children   []*Widget              `json:"Children"`
	Properties map[string]interface{} `json:"Properties"`
}

func WidgetParser(window fyne.Window, widget *Widget) (fyne.CanvasObject, error) {
	newWidget, err := defaultWidgetParser(window, widget)
	if newWidget == nil {
		newWidget, err = customWidgetParser(window, widget)
	}
	return newWidget, err
}

func defaultWidgetParser(window fyne.Window, widget *Widget) (fyne.CanvasObject, error) {
	switch widget.Type {
	case "VBoxContainer":
		return VBoxContainerHandler(widget.Properties, widget.Children, window)
	case "HBoxContainer":
		return HBoxContainerHandler(widget.Properties, widget.Children, window)
	case "Form":
		return FormHandler(widget.Properties, widget.Children, window)
	case "Label":
		return LabelHandler(widget.Properties, widget.Children, window)
	case "Button":
		return ButtonHandler(widget.Properties, widget.Children, window)
	case "ToolBar":
		return ToolBarHandler(widget.Properties, widget.Children, window)
	case "Image":
		return ImageHandler(widget.Properties, widget.Children, window)
	case "Entry":
		return EntryHandler(widget.Properties, widget.Children, window)
	case "NumberEntry":
		return NumEntryHandler(widget.Properties, widget.Children, window)
	case "PasswordEntry":
		return PasswordEntryHandler(widget.Properties, widget.Children, window)
	case "Slider":
		return SliderHandler(widget.Properties, widget.Children, window)
	default:
		return nil, NFError.ErrNotImplemented
	}
}

func customWidgetParser(window fyne.Window, widget *Widget) (fyne.CanvasObject, error) {
	if handler, ok := customWidgets[widget.Type]; ok {
		return handler(window, widget.Properties, widget.Children)
	} else {
		return nil, NFError.ErrNotImplemented
	}
}

// Add adds a custom widget to the customWidgets map
func Add(name string, handler widgetHandler) {
	customWidgets[name] = handler
}
