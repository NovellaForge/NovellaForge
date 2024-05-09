package DefaultLayouts

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFLayout"
)

// VBoxLayoutHandler simply adds all children to a vertical box
func VBoxLayoutHandler(window fyne.Window, _ *NFData.NFInterfaceMap, l *NFLayout.Layout) (fyne.CanvasObject, error) {
	vbox := container.NewVBox()
	for _, child := range l.Children {
		if child.Type == "Button" {
			//Print the on tapped event
			var onTapped string
			err := child.Args.Get("OnTapped", &onTapped)
			if err != nil {
				return nil, NFError.NewErrMissingArgument(l.Type, "OnTapped")
			}
		}
		widget, err := child.Parse(window)
		if err != nil {
			return nil, err
		}
		vbox.Add(widget)
	}
	return vbox, nil
}

// HBoxLayoutHandler simply adds all children to a horizontal box
func HBoxLayoutHandler(window fyne.Window, _ *NFData.NFInterfaceMap, l *NFLayout.Layout) (fyne.CanvasObject, error) {
	hbox := container.NewHBox()
	for _, child := range l.Children {
		widget, err := child.Parse(window)
		if err != nil {
			return nil, err
		}
		hbox.Add(widget)
	}
	return hbox, nil
}

// GridLayoutHandler simply adds all children to a grid that has args["Columns"] columns
func GridLayoutHandler(window fyne.Window, args *NFData.NFInterfaceMap, l *NFLayout.Layout) (fyne.CanvasObject, error) {
	var columns int
	err := args.Get("Columns", &columns)
	if err != nil {
		return nil, NFError.NewErrMissingArgument(l.Type, "Columns")
	}
	grid := container.NewGridWithColumns(columns)
	for _, child := range l.Children {
		widget, err := child.Parse(window)
		if err != nil {
			return nil, err
		}
		grid.Add(widget)
	}
	return grid, nil
}

// TabLayoutHandler simply adds all children to a tab layout
func TabLayoutHandler(window fyne.Window, _ *NFData.NFInterfaceMap, l *NFLayout.Layout) (fyne.CanvasObject, error) {
	tabs := container.NewAppTabs()
	for _, child := range l.Children {
		widget, err := child.Parse(window)
		if err != nil {
			return nil, err
		}
		var tabTitle string
		err = child.Args.Get("TabTitle", &tabTitle)
		if err != nil {
			tabTitle = "Children Missing TabTitle Arg"
		}
		tabs.Append(container.NewTabItem(tabTitle, widget))
	}
	return tabs, nil
}

// BorderLayoutHandler simply adds all children to a border layout
// The children must have a "Position" property that is either "Top", "Bottom", "Left", "Right", or "Center"
// The children are placed into the border layout based on their position
func BorderLayoutHandler(window fyne.Window, _ *NFData.NFInterfaceMap, l *NFLayout.Layout) (fyne.CanvasObject, error) {
	var top fyne.CanvasObject = nil
	var bottom fyne.CanvasObject = nil
	var left fyne.CanvasObject = nil
	var right fyne.CanvasObject = nil
	var center []fyne.CanvasObject = nil
	for _, child := range l.Children {
		widget, err := child.Parse(window)
		if err != nil {
			return nil, err
		}
		var position string
		err = child.Args.Get("BorderPosition", &position)
		if err != nil {
			center = append(center, widget)
		}
		switch position {
		case "Top":
			top = widget
		case "Bottom":
			bottom = widget
		case "Left":
			left = widget
		case "Right":
			right = widget
		default:
			center = append(center, widget)
		}
	}
	return container.NewBorder(top, bottom, left, right, center...), nil
}
