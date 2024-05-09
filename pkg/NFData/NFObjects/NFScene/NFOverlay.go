package NFScene

import (
	"errors"
	"fyne.io/fyne/v2"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFLayout"
)

var Overlays = make(map[string]*NFOverlay)

type NFOverlay struct {
	name    string
	layout  *NFLayout.Layout
	visible bool
}

func (o *NFOverlay) Name() string {
	return o.name
}

func (o *NFOverlay) SetName(name string) {
	o.name = name
}

func (o *NFOverlay) Layout() *NFLayout.Layout {
	return o.layout
}

func (o *NFOverlay) SetLayout(layout *NFLayout.Layout) {
	o.layout = layout
}

func (o *NFOverlay) Visible() bool {
	return o.visible
}

func (o *NFOverlay) SetVisible(visible bool, window fyne.Window, updateScenes ...*SceneStack) {
	o.visible = visible
	for _, scene := range updateScenes {
		scene.RefreshOverlays(window)
	}
}

func (o *NFOverlay) Hide(window fyne.Window, updateScenes ...*SceneStack) {
	o.visible = false
	for _, scene := range updateScenes {
		scene.RefreshOverlays(window)
	}
}

func (o *NFOverlay) Show(window fyne.Window, updateScenes ...*SceneStack) {
	o.visible = true
	for _, scene := range updateScenes {
		scene.RefreshOverlays(window)
	}
}

func NewNFOverlay(name string, layout *NFLayout.Layout) *NFOverlay {
	return &NFOverlay{
		name:    name,
		layout:  layout,
		visible: false,
	}
}

func AddOverlay(overlay ...*NFOverlay) error {
	for _, o := range overlay {
		//check if the overlay name is already mapped
		if _, ok := Overlays[o.name]; !ok {
			//Map the overlay
			Overlays[o.name] = o
		} else {
			return errors.New("overlay with that name already exists")
		}
	}
	return nil
}
