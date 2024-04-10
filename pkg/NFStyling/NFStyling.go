package NFStyling

import "fyne.io/fyne/v2"

type NFStyling struct {
	fyne.TextStyle
	Wrapping fyne.TextWrap
}

func NewTextStyling() NFStyling {
	return NFStyling{Wrapping: fyne.TextWrapWord, TextStyle: fyne.TextStyle{Bold: false, Italic: false, Monospace: false, Symbol: false, TabWidth: 4}}
}

func (t *NFStyling) SetBold(bold bool, parent fyne.CanvasObject) {
	t.Bold = bold
	if parent != nil {
		parent.Refresh()
	}

}

func (t *NFStyling) SetItalic(italic bool, parent fyne.CanvasObject) {
	t.Italic = italic
	if parent != nil {
		parent.Refresh()
	}

}

func (t *NFStyling) SetMonospace(monospace bool, parent fyne.CanvasObject) {
	t.Monospace = monospace
	if parent != nil {
		parent.Refresh()
	}

}

func (t *NFStyling) SetSymbol(symbol bool, parent fyne.CanvasObject) {
	t.Symbol = symbol
	if parent != nil {
		parent.Refresh()
	}

}

func (t *NFStyling) SetTabWidth(tabWidth int, parent fyne.CanvasObject) {
	t.TabWidth = tabWidth
	if parent != nil {
		parent.Refresh()
	}
}
