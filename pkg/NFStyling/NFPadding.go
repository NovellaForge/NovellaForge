package NFStyling

import "fyne.io/fyne/v2"

type NFPadding struct {
	Top    float32
	Bottom float32
	Left   float32
	Right  float32
}

func NewPadding(top float32, bottom float32, left float32, right float32) NFPadding {
	return NFPadding{Top: top, Bottom: bottom, Left: left, Right: right}
}

func (p *NFPadding) SetAll(top, bottom, left, right float32, parent fyne.CanvasObject) {
	p.Top = top
	p.Bottom = bottom
	p.Left = left
	p.Right = right
	if parent != nil {
		parent.Refresh()
	}
}

func (p *NFPadding) SetVertical(vertical float32, parent fyne.CanvasObject) {
	p.Top = vertical
	p.Bottom = vertical
	if parent != nil {
		parent.Refresh()
	}
}

func (p *NFPadding) Vertical() float32 {
	return p.Top + p.Bottom
}

func (p *NFPadding) SetHorizontal(horizontal float32, parent fyne.CanvasObject) {
	p.Left = horizontal
	p.Right = horizontal
	if parent != nil {
		parent.Refresh()
	}
}

func (p *NFPadding) Horizontal() float32 {
	return p.Left + p.Right
}

// Origin returns the top left combination of the padding, but it can accept float other to offset the origin
//
// One arg will offset the origin by the arg in the x direction,
//
// # Two other will offset the origin by the first arg in the x direction and the second arg in the y direction
//
// # Three other will offset the origin by the first arg in the x direction, the second arg in the y direction, and the third added to both
//
// Example: padding.Origin() will return a fyne.Position with the x value of padding.Left and the y value of padding.Top
//
// Example: padding.Origin(10) will return a fyne.Position with the x value of padding.Left+10 and the y value of padding.Top
//
// Example: padding.Origin(10, 20) will return a fyne.Position with the x value of padding.Left+10 and the y value of padding.Top+20
//
// Example: padding.Origin(10, 20, 30) will return a fyne.Position with the x value of padding.Left+10+30 and the y value of padding.Top+20+30
func (p *NFPadding) Origin(args ...float32) fyne.Position {
	//Switch on the number of other
	switch len(args) {
	case 0:
		return fyne.NewPos(p.Left, p.Top)
	case 1:
		return fyne.NewPos(p.Left+(args[0]), p.Top)
	case 2:
		return fyne.NewPos(p.Left+args[0], p.Top+args[1])
	default:
		return fyne.NewPos(p.Left+args[0]+args[2], p.Top+args[1]+args[2])
	}
}

func (p *NFPadding) SetOrigin(x, y float32, parent fyne.CanvasObject) {
	p.Left = x
	p.Top = y
	if parent != nil {
		parent.Refresh()
	}
}

func (p *NFPadding) SetTop(top float32, parent fyne.CanvasObject) {
	p.Top = top
	if parent != nil {
		parent.Refresh()
	}
}

func (p *NFPadding) SetBottom(bottom float32, parent fyne.CanvasObject) {
	p.Bottom = bottom
	if parent != nil {
		parent.Refresh()
	}
}

func (p *NFPadding) SetLeft(left float32, parent fyne.CanvasObject) {
	p.Left = left
	if parent != nil {
		parent.Refresh()
	}
}

func (p *NFPadding) SetRight(right float32, parent fyne.CanvasObject) {
	p.Right = right
	if parent != nil {
		parent.Refresh()
	}
}
