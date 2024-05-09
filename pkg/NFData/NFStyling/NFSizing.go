package NFStyling

import "fyne.io/fyne/v2"

type NFSizing struct {
	MinWidth  float32
	MinHeight float32
	MaxWidth  float32
	MaxHeight float32
	FitWidth  bool
	FitHeight bool
}

func NewSizing(minWidth float32, minHeight float32, maxWidth float32, maxHeight float32, fitWidth bool, fitHeight bool) NFSizing {
	return NFSizing{MinWidth: minWidth, MinHeight: minHeight, MaxWidth: maxWidth, MaxHeight: maxHeight, FitWidth: fitWidth, FitHeight: fitHeight}
}

func (s *NFSizing) SetMinWidth(minWidth float32, parent fyne.CanvasObject) {
	s.MinWidth = minWidth
	if parent != nil {
		parent.Refresh()
	}
}

func (s *NFSizing) SetMinHeight(minHeight float32, parent fyne.CanvasObject) {
	s.MinHeight = minHeight
	if parent != nil {
		parent.Refresh()
	}
}

func (s *NFSizing) SetMaxWidth(maxWidth float32, parent fyne.CanvasObject) {
	s.MaxWidth = maxWidth
	if parent != nil {
		parent.Refresh()
	}
}

func (s *NFSizing) SetMaxHeight(maxHeight float32, parent fyne.CanvasObject) {
	s.MaxHeight = maxHeight
	if parent != nil {
		parent.Refresh()
	}
}

func (s *NFSizing) SetFitWidth(fitWidth bool, parent fyne.CanvasObject) {
	s.FitWidth = fitWidth
	if parent != nil {
		parent.Refresh()
	}
}

func (s *NFSizing) SetFitHeight(fitHeight bool, parent fyne.CanvasObject) {
	s.FitHeight = fitHeight
	if parent != nil {
		parent.Refresh()
	}
}
