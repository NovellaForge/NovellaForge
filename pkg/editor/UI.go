package editor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CustomTree extends the widget.Tree for more control over the nodes
type CustomTree struct {
	widget.Tree
}

// NewCustomTree creates a new instance of the custom tree
func NewCustomTree() *CustomTree {
	t := &CustomTree{}
	t.ExtendBaseWidget(t)
	return t
}

// CreateNode is responsible for creating each node in the tree.
func (t *CustomTree) CreateNode(branch bool) fyne.CanvasObject {
	if branch {
		return widget.NewLabel("Branch template")
	}
	box := container.NewHBox()
	box.Add(widget.NewLabel("Leaf template"))
	box.Add(widget.NewButton("+", func() {
		// TODO: Logic to move the scene/scene group up in its slice
	}))
	box.Add(widget.NewButton("-", func() {
		// TODO: Logic to move the scene/scene group down in its slice
	}))
	return box
}
