package CalsWidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
	"image/color"
	"strings"
	"time"
)

// Dialog is a widget that displays a message to the user while also having interactive elements
type Dialog struct {
	//Interfaces for the dialog
	widget.BaseWidget
	widget.DisableableWidget

	//Pointers to the objects that make up the dialog
	label, nameLabel                      *widget.Label
	border, debug                         *canvas.Rectangle
	scroll                                *container.Scroll
	nameBorder                            *canvas.Rectangle
	sizing, nameSizing                    *NFWidget.Sizing
	padding, namePadding, externalPadding *NFWidget.Padding
	contentStyle, nameStyle               *NFWidget.TextStyling
	tapAnim, hoverAnim, stateAnim         *fyne.Animation

	//Settings for the dialog
	AllText                       []string
	curText                       string
	State                         int
	MaxState                      int
	Name                          string
	HasName                       bool
	StrokeColor, NameStrokeColor  color.Color
	Fill, NameFill                color.Color
	Stroke, NameStroke            float32
	NamePosition                  fyne.Position
	NameAffectsLayout             bool
	StateOnTap, TextOnStateChange bool
	ConcatText, AnimateText       bool
	skipAnim, animating, CanSkip  bool
	TextDelay                     float32
	animatedStates                []int

	//Functions for the dialog
	OnStateChange     func(int)
	OnTapped          func()
	OnSecondaryTapped func()
	OnDoubleTapped    func()
	OnHover           func()
	OnEndHover        func()
	WhileHover        func()
}

type dialogRenderer struct {
	objects []fyne.CanvasObject
	dialog  *Dialog
}

func NewDialog(text []string) *Dialog {
	dialog := &Dialog{
		AllText:           text,
		State:             0,
		MaxState:          0,
		StateOnTap:        true,
		TextOnStateChange: true,
		ConcatText:        false,
		AnimateText:       true,
		CanSkip:           true,
		TextDelay:         25,
		Name:              "",
		contentStyle:      NFWidget.NewTextStyling(),
		sizing:            NFWidget.NewSizing(100, 100, 200, 200, true, false),
		padding:           NFWidget.NewPadding(5, 5, 5, 5),
		externalPadding:   NFWidget.NewPadding(5, 5, 5, 5),
		StrokeColor:       color.RGBA{A: 255},
		Fill:              color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Stroke:            1,
		HasName:           false,
		NameAffectsLayout: true,
		nameStyle:         NFWidget.NewTextStyling(),
		namePadding:       NFWidget.NewPadding(1, 1, 1, 1),
		nameSizing:        NFWidget.NewSizing(100, 50, 200, 50, false, false),
		NameStrokeColor:   color.RGBA{A: 255},
		NameFill:          color.RGBA{B: 255, A: 255},
		NameStroke:        1,
		NamePosition:      fyne.NewPos(200, 50),
	}

	dialog.ExtendBaseWidget(dialog)
	return dialog
}

func NewDialogWithName(name string, text []string) *Dialog {
	dialog := NewDialog(text)
	dialog.SetName(name)
	dialog.SetHasName(true)
	return dialog
}

func (d *Dialog) CreateRenderer() fyne.WidgetRenderer {
	debugBorder := canvas.NewRectangle(color.RGBA{R: 255, A: 255})
	debugBorder.StrokeColor = color.RGBA{R: 255, A: 255}
	debugBorder.StrokeWidth = 1
	d.debug = debugBorder

	border := canvas.NewRectangle(color.Transparent)
	border.FillColor = d.Fill
	border.StrokeColor = d.StrokeColor
	border.StrokeWidth = d.Stroke
	label := widget.NewLabel("")
	label.TextStyle = d.contentStyle.TextStyle
	label.Wrapping = d.contentStyle.Wrapping
	scroll := container.NewScroll(label)
	nameLabel := widget.NewLabel(d.Name)
	nameLabel.TextStyle = d.nameStyle.TextStyle
	nameLabel.Wrapping = d.nameStyle.Wrapping
	nameBorder := canvas.NewRectangle(color.Transparent)
	nameBorder.FillColor = d.NameFill
	nameBorder.StrokeColor = d.NameStrokeColor
	nameBorder.StrokeWidth = d.NameStroke
	d.border = border
	d.label = label
	d.scroll = scroll
	d.nameLabel = nameLabel
	d.nameBorder = nameBorder
	//objects := []fyne.CanvasObject{debugBorder, border, scroll, nameLabel, nameBorder}
	objects := []fyne.CanvasObject{border, scroll, nameLabel, nameBorder}
	r := &dialogRenderer{
		objects: objects,
		dialog:  d,
	}
	return r
}

func (r *dialogRenderer) MinSize() fyne.Size {
	d := r.dialog
	width := d.sizing.MinWidth
	height := d.sizing.MinHeight

	if d.HasName && d.NameAffectsLayout {
		nameSize := d.nameLabel.Size()
		// Initialize calculatedPadding with user-set externalPadding
		calculatedPadding := NFWidget.NewPadding(d.externalPadding.Top, d.externalPadding.Bottom, d.externalPadding.Left, d.externalPadding.Right)

		// Only add padding to the sides that nameBorder extends over
		if d.NamePosition.X < 0 {
			calculatedPadding.Left = max(calculatedPadding.Left, -d.NamePosition.X)
		}
		if d.NamePosition.Y < 0 {
			calculatedPadding.Top = max(calculatedPadding.Top, -d.NamePosition.Y)
		}
		if d.NamePosition.X+nameSize.Width > width {
			calculatedPadding.Right = max(calculatedPadding.Right, d.NamePosition.X+nameSize.Width-width)
		}
		if d.NamePosition.Y+nameSize.Height > height {
			calculatedPadding.Bottom = max(calculatedPadding.Bottom, d.NamePosition.Y+nameSize.Height-height)
		}

		// Adjust the MinSize based on the effective externalPadding
		width += calculatedPadding.Left + calculatedPadding.Right
		height += calculatedPadding.Top + calculatedPadding.Bottom
	}

	return fyne.NewSize(width, height)
}

func (r *dialogRenderer) Layout(size fyne.Size) {
	d := r.dialog

	d.debug.Resize(size)
	d.debug.Move(fyne.NewPos(0, 0))

	dialogSize := fyne.NewSize(min(d.sizing.MaxWidth, size.Width), min(d.sizing.MaxHeight, size.Height))
	if d.sizing.FitWidth {
		dialogSize.Width = size.Width
	}
	if d.sizing.FitHeight {
		dialogSize.Height = size.Height
	}

	// Adjust the main dialog border size and position
	borderSize := fyne.NewSize(dialogSize.Width-d.externalPadding.Horizontal(), dialogSize.Height-d.externalPadding.Vertical())
	d.border.Resize(borderSize)
	d.border.Move(fyne.NewPos(d.externalPadding.Left, d.externalPadding.Top))

	// Configure text label
	d.label.TextStyle = d.contentStyle.TextStyle
	d.label.Wrapping = d.contentStyle.Wrapping
	d.border.FillColor = d.Fill
	d.border.StrokeColor = d.StrokeColor
	d.border.StrokeWidth = d.Stroke

	// Configure scroll container
	scrollSize := fyne.NewSize(borderSize.Width-(2*d.Stroke+d.padding.Horizontal()), borderSize.Height-(2*d.Stroke+d.padding.Vertical()))
	d.scroll.Resize(scrollSize)
	d.scroll.Move(fyne.NewPos(d.Stroke+d.padding.Left+d.externalPadding.Left, d.Stroke+d.padding.Top+d.externalPadding.Top))

	if d.HasName {
		d.nameBorder.Show()
		d.nameLabel.Show()
		// Configure Name label and border
		d.nameLabel.TextStyle = d.nameStyle.TextStyle
		d.nameLabel.Wrapping = d.nameStyle.Wrapping
		d.nameBorder.FillColor = d.NameFill
		d.nameBorder.StrokeColor = d.NameStrokeColor
		d.nameBorder.StrokeWidth = d.NameStroke

		nameSize := d.nameLabel.Size()
		if nameSize.Width > d.nameSizing.MaxWidth {
			nameSize.Width = d.nameSizing.MaxWidth
		} else if nameSize.Width < d.nameSizing.MinWidth && d.nameSizing.MinWidth < d.nameSizing.MaxWidth {
			nameSize.Width = d.nameSizing.MinWidth
		}

		if nameSize.Height > d.nameSizing.MaxHeight {
			nameSize.Height = d.nameSizing.MaxHeight
		} else if nameSize.Height < d.nameSizing.MinHeight && d.nameSizing.MinHeight < d.nameSizing.MaxHeight {
			nameSize.Height = d.nameSizing.MinHeight
		}

		d.nameLabel.Resize(nameSize)
		d.nameBorder.Resize(fyne.NewSize(nameSize.Width+d.namePadding.Left+d.namePadding.Right, nameSize.Height+d.namePadding.Top+d.namePadding.Bottom))

		if d.NameAffectsLayout {
			//Shift struct containing the amount to shift each axis
			shift := fyne.NewPos(0, 0)
			newNamePosition := fyne.NewPos(d.NamePosition.X, d.NamePosition.Y)
			if d.NamePosition.X < 0 {
				shift.X = -d.NamePosition.X
				newNamePosition.X = 0 + d.externalPadding.Left
			}
			if d.NamePosition.Y < 0 {
				shift.Y = -d.NamePosition.Y
				newNamePosition.Y = 0 + d.externalPadding.Top
			}
			d.nameBorder.Move(newNamePosition)
			d.nameLabel.Move(fyne.NewPos(newNamePosition.X+d.namePadding.Left, newNamePosition.Y+d.namePadding.Top))

			//Resize the main dialog border and scrollbox to account for the shifted nameBorder
			d.border.Resize(fyne.NewSize(d.border.Size().Width-shift.X, d.border.Size().Height-shift.Y))
			d.scroll.Resize(fyne.NewSize(d.scroll.Size().Width-shift.X, d.scroll.Size().Height-shift.Y))
			//Move the main dialog border and scrollbox to account for the shifted nameBorder
			d.border.Move(d.border.Position().Add(shift))
			d.scroll.Move(d.scroll.Position().Add(shift))

		} else {
			d.nameBorder.Move(fyne.NewPos(d.NamePosition.X+d.externalPadding.Left, d.NamePosition.Y+d.externalPadding.Top))
			d.nameLabel.Move(fyne.NewPos(d.NamePosition.X+d.namePadding.Left+d.externalPadding.Left, d.NamePosition.Y+d.namePadding.Top+d.externalPadding.Top))
		}
	} else {
		d.nameBorder.Hide()
		d.nameLabel.Hide()
	}
}

func (r *dialogRenderer) Refresh() {
	d := r.dialog
	if d.TextOnStateChange {
		if len(d.AllText) == 0 {
			d.AllText = append(d.AllText, "")
		}
		d.MaxState = len(d.AllText) - 1
		if d.State > d.MaxState {
			d.State = d.MaxState
		} else if d.State < 0 {
			d.State = 0
		}
	}
	d.label.TextStyle = d.contentStyle.TextStyle
	d.label.Wrapping = d.contentStyle.Wrapping
	d.label.Refresh()
	d.UpdateText()
	d.border.FillColor = d.Fill
	d.border.StrokeColor = d.StrokeColor
	d.border.StrokeWidth = d.Stroke
	d.border.Refresh()
	if d.HasName {
		d.nameLabel.TextStyle = d.nameStyle.TextStyle
		d.nameLabel.Wrapping = d.nameStyle.Wrapping
		d.nameBorder.FillColor = d.NameFill
		d.nameBorder.StrokeColor = d.NameStrokeColor
		d.nameBorder.StrokeWidth = d.NameStroke
		d.nameBorder.Refresh()
		d.nameLabel.Refresh()
	}
	d.scroll.Refresh()
}

func (r *dialogRenderer) ApplyTheme() {}

func (r *dialogRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *dialogRenderer) Destroy() {}

func (r *dialogRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// UpdateText updates the text displayed by the dialog
func (d *Dialog) UpdateText(args ...int) {
	stringToDisplay := ""
	var newContent []string
	if d.ConcatText {
		for i := 0; i < len(d.AllText); i++ {
			if i == 0 {
				newContent = append(newContent, d.AllText[i])
			} else {
				newContent = append(newContent, strings.Join(d.AllText[:i], "\n\n"))
				newContent = append(newContent, d.AllText[i])
			}
		}
	}
	if d.TextOnStateChange {
		if d.ConcatText {
			stringToDisplay = newContent[d.State]
		} else {
			stringToDisplay = d.AllText[d.State]
		}
	} else if len(args) > 0 {
		if d.ConcatText {
			stringToDisplay = newContent[args[0]]
		} else {
			stringToDisplay = d.AllText[args[0]]
		}
	}
	//check if the current state is in animatedStates
	isAnimated := false
	for _, state := range d.animatedStates {
		if state == d.State {
			isAnimated = true
		}
	}
	d.curText = stringToDisplay
	if d.AnimateText && !d.animating && !isAnimated {
		d.animating = true
		d.animatedStates = append(d.animatedStates, d.State)
		go func(d *Dialog) {
			for i := 0; i < len(d.curText); i++ {
				//Make sure the slice hasn't changed
				if d.curText != stringToDisplay {
					d.animating = false
					return
				}
				d.SetText(d.curText[:i+1])
				time.Sleep(time.Duration(d.TextDelay) * time.Millisecond)
				if d.skipAnim {
					d.skipAnim = false
					d.SetText(d.curText)
					d.animating = false
					return
				}
			}
			d.animating = false
		}(d)
	} else {
		d.SetText(d.curText)
	}
}

func (d *Dialog) stateChanged() {
	if d.State > d.MaxState {
		d.MaxState = d.State
	} else if d.TextOnStateChange {
		d.UpdateText()
		if d.OnStateChange != nil {
			d.OnStateChange(d.State)
		}
	}
}

func (d *Dialog) Tapped(*fyne.PointEvent) {
	if d.Disabled() {
		return
	}
	if d.StateOnTap && !d.animating {
		d.AddState()
	} else if d.animating && d.CanSkip {
		d.skipAnim = true
	}

	if d.OnTapped != nil {
		d.OnTapped()
	}
}

func (d *Dialog) TappedSecondary(*fyne.PointEvent) {
	if d.Disabled() {
		return
	}
	if d.OnSecondaryTapped != nil {
		d.OnSecondaryTapped()
	}
}

func (d *Dialog) DoubleTapped(*fyne.PointEvent) {
	if d.Disabled() {
		return
	}
	if d.OnDoubleTapped != nil {
		d.OnDoubleTapped()
	}
}

func (d *Dialog) MouseIn(*desktop.MouseEvent) {
	if d.Disabled() {
		return
	}
	if d.OnHover != nil {
		d.OnHover()
	}
}

func (d *Dialog) MouseMoved(*desktop.MouseEvent) {
	if d.Disabled() {
		return
	}
	if d.WhileHover != nil {
		d.WhileHover()
	}
}

func (d *Dialog) MouseOut() {
	if d.Disabled() {
		return
	}
	if d.OnEndHover != nil {
		d.OnEndHover()
	}
}

//Getters and Setters

func (d *Dialog) Padding() *NFWidget.Padding {
	return d.padding
}

func (d *Dialog) NamePadding() *NFWidget.Padding {
	return d.namePadding
}

func (d *Dialog) ExternalPadding() *NFWidget.Padding {
	return d.externalPadding
}

func (d *Dialog) ContentStyle() *NFWidget.TextStyling {
	return d.contentStyle
}

func (d *Dialog) NameStyle() *NFWidget.TextStyling {
	return d.nameStyle
}

func (d *Dialog) Sizing() *NFWidget.Sizing {
	return d.sizing
}

func (d *Dialog) NameSizing() *NFWidget.Sizing {
	return d.nameSizing
}

func (d *Dialog) SetStateOnTap(stateOnTap bool) {
	d.StateOnTap = stateOnTap
	d.Refresh()
}

func (d *Dialog) SetTextOnStateChange(textOnStateChange bool) {
	d.TextOnStateChange = textOnStateChange
	d.Refresh()
}

func (d *Dialog) SetConcatText(concatText bool) {
	d.ConcatText = concatText
	d.Refresh()
}

func (d *Dialog) SetAnimateText(animateText bool) {
	d.AnimateText = animateText
	d.Refresh()
}

func (d *Dialog) SetTextAnimDelay(textAnimDelay float32) {
	d.TextDelay = textAnimDelay
	d.Refresh()
}

func (d *Dialog) SetText(text string) {
	d.label.SetText(text)
}

func (d *Dialog) SetName(name string) {
	d.Name = name
	d.Refresh()
}

func (d *Dialog) SetHasName(hasName bool) {
	d.HasName = hasName
	d.Refresh()
}

func (d *Dialog) SetStrokeColor(strokeColor color.Color) {
	d.StrokeColor = strokeColor
	d.Refresh()
}

func (d *Dialog) SetNameStrokeColor(nameStrokeColor color.Color) {
	d.NameStrokeColor = nameStrokeColor
	d.Refresh()
}

func (d *Dialog) SetFill(fill color.Color) {
	d.Fill = fill
	d.Refresh()
}

func (d *Dialog) SetNameFill(nameFill color.Color) {
	d.NameFill = nameFill
	d.Refresh()
}

func (d *Dialog) SetStroke(stroke float32) {
	d.Stroke = stroke
	d.Refresh()
}

func (d *Dialog) SetNameStroke(nameStroke float32) {
	d.NameStroke = nameStroke
	d.Refresh()
}

func (d *Dialog) SetNamePosition(namePosition fyne.Position) {
	d.NamePosition = namePosition
	d.Refresh()
}

func (d *Dialog) SetNameAffectsLayout(nameAffectsLayout bool) {
	d.NameAffectsLayout = nameAffectsLayout
	d.Refresh()
}

// SetContent sets the AllText field of the dialog; however, it does not update the label, UpdateText() must be called to do that
func (d *Dialog) SetContent(content []string) {
	d.AllText = content
	d.Refresh()
}

// AddContent adds a string to the AllText field of the dialog; however, it does not update the displayed text, UpdateText() must be called to do that
func (d *Dialog) AddContent(content string) {
	d.AllText = append(d.AllText, content)
	d.Refresh()
}

func (d *Dialog) SetState(state int) {
	d.State = state
	d.stateChanged()
}

func (d *Dialog) AddState() {
	d.State++
	d.stateChanged()

}

func (d *Dialog) ChangeState(incr int) {
	d.State += incr
	d.stateChanged()
}
