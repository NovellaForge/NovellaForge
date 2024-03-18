package DefaultLayouts

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/NovellaForge/NovellaForge/pkg/NFError"
	"github.com/NovellaForge/NovellaForge/pkg/NFLayout"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
)

// Import is a function used to allow importing of the default layouts package without errors or warnings
func Import() {}

// init() registers the default layouts to be used within the game
// The init function is called when the package is imported, but in order
// to avoid unused import warnings, you can call the Import() function, which does nothing
func init() {
	Import()
	log.Println("Registering Default Layouts")
	NFLayout.Register("VBox", VBoxLayoutHandler)
	NFLayout.Register("HBox", HBoxLayoutHandler)
	NFLayout.Register("Grid", GridLayoutHandler)
	NFLayout.Register("Tabs", TabLayoutHandler)
	NFLayout.Register("Border", BorderLayoutHandler)
}

// VBoxLayoutHandler simply adds all children to a vertical box
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

// HBoxLayoutHandler simply adds all children to a horizontal box
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

// GridLayoutHandler simply adds all children to a grid that has args["Columns"] columns
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

// TabLayoutHandler simply adds all children to a tab layout
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

// BorderLayoutHandler simply adds all children to a border layout
// The children must have a "Position" property that is either "Top", "Bottom", "Left", "Right", or "Center"
// The children are placed into the border layout based on their position
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
