package NFWidget

import "fyne.io/fyne/v2"

type TextStyling struct {
	fyne.TextStyle
	Wrapping fyne.TextWrap
}

func NewTextStyling() *TextStyling {
	return &TextStyling{Wrapping: fyne.TextWrapWord, TextStyle: fyne.TextStyle{Bold: false, Italic: false, Monospace: false, Symbol: false, TabWidth: 4}}
}

func (t *TextStyling) SetBold(bold bool, parent fyne.CanvasObject) {
	t.Bold = bold
	if parent != nil {
		parent.Refresh()
	}

}

func (t *TextStyling) SetItalic(italic bool, parent fyne.CanvasObject) {
	t.Italic = italic
	if parent != nil {
		parent.Refresh()
	}

}

func (t *TextStyling) SetMonospace(monospace bool, parent fyne.CanvasObject) {
	t.Monospace = monospace
	if parent != nil {
		parent.Refresh()
	}

}

func (t *TextStyling) SetSymbol(symbol bool, parent fyne.CanvasObject) {
	t.Symbol = symbol
	if parent != nil {
		parent.Refresh()
	}

}

func (t *TextStyling) SetTabWidth(tabWidth int, parent fyne.CanvasObject) {
	t.TabWidth = tabWidth
	if parent != nil {
		parent.Refresh()
	}
}

type Sizing struct {
	MinWidth  float32
	MinHeight float32
	MaxWidth  float32
	MaxHeight float32
	FitWidth  bool
	FitHeight bool
}

func NewSizing(minWidth float32, minHeight float32, maxWidth float32, maxHeight float32, fitWidth bool, fitHeight bool) *Sizing {
	return &Sizing{MinWidth: minWidth, MinHeight: minHeight, MaxWidth: maxWidth, MaxHeight: maxHeight, FitWidth: fitWidth, FitHeight: fitHeight}
}

func (s *Sizing) SetMinWidth(minWidth float32, parent fyne.CanvasObject) {
	s.MinWidth = minWidth
	if parent != nil {
		parent.Refresh()
	}
}

func (s *Sizing) SetMinHeight(minHeight float32, parent fyne.CanvasObject) {
	s.MinHeight = minHeight
	if parent != nil {
		parent.Refresh()
	}
}

func (s *Sizing) SetMaxWidth(maxWidth float32, parent fyne.CanvasObject) {
	s.MaxWidth = maxWidth
	if parent != nil {
		parent.Refresh()
	}
}

func (s *Sizing) SetMaxHeight(maxHeight float32, parent fyne.CanvasObject) {
	s.MaxHeight = maxHeight
	if parent != nil {
		parent.Refresh()
	}
}

func (s *Sizing) SetFitWidth(fitWidth bool, parent fyne.CanvasObject) {
	s.FitWidth = fitWidth
	if parent != nil {
		parent.Refresh()
	}
}

func (s *Sizing) SetFitHeight(fitHeight bool, parent fyne.CanvasObject) {
	s.FitHeight = fitHeight
	if parent != nil {
		parent.Refresh()
	}
}

type Padding struct {
	Top    float32
	Bottom float32
	Left   float32
	Right  float32
}

func NewPadding(top float32, bottom float32, left float32, right float32) *Padding {
	return &Padding{Top: top, Bottom: bottom, Left: left, Right: right}
}

func (p *Padding) SetAll(top, bottom, left, right float32, parent fyne.CanvasObject) {
	p.Top = top
	p.Bottom = bottom
	p.Left = left
	p.Right = right
	if parent != nil {
		parent.Refresh()
	}
}

func (p *Padding) SetVertical(vertical float32, parent fyne.CanvasObject) {
	p.Top = vertical
	p.Bottom = vertical
	if parent != nil {
		parent.Refresh()
	}
}

func (p *Padding) Vertical() float32 {
	return p.Top + p.Bottom
}

func (p *Padding) SetHorizontal(horizontal float32, parent fyne.CanvasObject) {
	p.Left = horizontal
	p.Right = horizontal
	if parent != nil {
		parent.Refresh()
	}
}

func (p *Padding) Horizontal() float32 {
	return p.Left + p.Right
}

func (p *Padding) SetTop(top float32, parent fyne.CanvasObject) {
	p.Top = top
	if parent != nil {
		parent.Refresh()
	}
}

func (p *Padding) SetBottom(bottom float32, parent fyne.CanvasObject) {
	p.Bottom = bottom
	if parent != nil {
		parent.Refresh()
	}
}

func (p *Padding) SetLeft(left float32, parent fyne.CanvasObject) {
	p.Left = left
	if parent != nil {
		parent.Refresh()
	}
}

func (p *Padding) SetRight(right float32, parent fyne.CanvasObject) {
	p.Right = right
	if parent != nil {
		parent.Refresh()
	}
}
