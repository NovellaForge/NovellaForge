package Default

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/NovellaForge/NovellaForge/pkg/NFError"
	"github.com/NovellaForge/NovellaForge/pkg/NFLayout"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
	"log"
)

// Import is a function that is used to allow importing of the default layouts package without errors or warnings
func Import() { log.Println("Importing Default Layouts") }

func init() {
	NFLayout.Register("VBox", VBoxLayoutHandler)
	NFLayout.Register("HBox", HBoxLayoutHandler)
	NFLayout.Register("Grid", GridLayoutHandler)
	NFLayout.Register("Tabs", TabLayoutHandler)
	NFLayout.Register("Border", BorderLayoutHandler)
}

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

func HBoxLayoutHandler(window fyne.Window, args map[string]interface{}, children []NFWidget.Widget) (fyne.CanvasObject, error) {
	hbox := container.NewHBox()
	for _, child := range children {
		widget, err := child.Parse(window)
		if err != nil {
			return nil, err
		}
		hbox.Add(widget)
	}
	return hbox, nil
}

func GridLayoutHandler(window fyne.Window, args map[string]interface{}, children []NFWidget.Widget) (fyne.CanvasObject, error) {
	if _, ok := args["Columns"]; !ok {
		return nil, NFError.ErrMissingArgument
	}
	columns := args["Columns"].(int)
	grid := container.NewGridWithColumns(columns)
	for _, child := range children {
		widget, err := child.Parse(window)
		if err != nil {
			return nil, err
		}
		grid.Add(widget)
	}
	return grid, nil
}

func TabLayoutHandler(window fyne.Window, args map[string]interface{}, children []NFWidget.Widget) (fyne.CanvasObject, error) {
	tabs := container.NewAppTabs()
	for _, child := range children {
		widget, err := child.Parse(window)
		if err != nil {
			return nil, err
		}
		tabs.Append(container.NewTabItem(child.Properties["Title"].(string), widget))
	}
	return tabs, nil
}

func BorderLayoutHandler(window fyne.Window, args map[string]interface{}, children []NFWidget.Widget) (fyne.CanvasObject, error) {
	var top fyne.CanvasObject = nil
	var bottom fyne.CanvasObject = nil
	var left fyne.CanvasObject = nil
	var right fyne.CanvasObject = nil
	var center fyne.CanvasObject = nil
	for _, child := range children {
		widget, err := child.Parse(window)
		if err != nil {
			return nil, err
		}
		switch child.Properties["Position"] {
		case "Top":
			top = widget
		case "Bottom":
			bottom = widget
		case "Left":
			left = widget
		case "Right":
			right = widget
		case "Center":
			center = widget
		default:
			return nil, NFError.ErrInvalidArgument
		}
	}
	return container.NewBorder(top, bottom, left, right, center), nil
}
