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

type JsonSafeDialog struct {
	AllText           []string             `json:"AllText"`
	State             int                  `json:"State"`
	MaxState          int                  `json:"MaxState"`
	Name              string               `json:"Name"`
	HasName           bool                 `json:"HasName"`
	StrokeColor       color.Color          `json:"StrokeColor"`
	NameStrokeColor   color.Color          `json:"NameStrokeColor"`
	Fill              color.Color          `json:"Fill"`
	NameFill          color.Color          `json:"NameFill"`
	Stroke            float32              `json:"Stroke"`
	NameStroke        float32              `json:"NameStroke"`
	NamePosition      fyne.Position        `json:"NamePosition"`
	NameAffectsLayout bool                 `json:"NameAffectsLayout"`
	StateOnTap        bool                 `json:"StateOnTap"`
	TextOnStateChange bool                 `json:"TextOnStateChange"`
	ConcatText        bool                 `json:"ConcatText"`
	AnimateText       bool                 `json:"AnimateText"`
	CanSkip           bool                 `json:"CanSkip"`
	TextDelay         float32              `json:"TextDelay"`
	OnStateChange     string               `json:"OnStateChange"`
	OnTapped          string               `json:"OnTapped"`
	OnSecondaryTapped string               `json:"OnSecondaryTapped"`
	OnDoubleTapped    string               `json:"OnDoubleTapped"`
	OnHover           string               `json:"OnHover"`
	OnEndHover        string               `json:"OnEndHover"`
	WhileHover        string               `json:"WhileHover"`
	Sizing            NFWidget.Sizing      `json:"Sizing"`
	Padding           NFWidget.Padding     `json:"Padding"`
	ExternalPadding   NFWidget.Padding     `json:"ExternalPadding"`
	ContentStyle      NFWidget.TextStyling `json:"ContentStyle"`
	NameStyle         NFWidget.TextStyling `json:"NameStyle"`
	NamePadding       NFWidget.Padding     `json:"NamePadding"`
	NameSizing        NFWidget.Sizing      `json:"NameSizing"`
}

// Dialog is a widget that displays a message to the user while also having interactive elements
type Dialog struct {
	//Interfaces for the dialog
	widget.BaseWidget
	widget.DisableableWidget

	//JSON-safe variables for the dialog and use in the editor
	JsonSafeDialog

	//Pointers to the objects that make up the dialog
	label, nameLabel              *widget.Label
	border, debug                 *canvas.Rectangle
	scroll                        *container.Scroll
	nameBorder                    *canvas.Rectangle
	tapAnim, hoverAnim, stateAnim *fyne.Animation

	//Functions for the dialog
	OnStateChange     func(int)
	OnTapped          func()
	OnSecondaryTapped func()
	OnDoubleTapped    func()
	OnHover           func()
	OnEndHover        func()
	WhileHover        func()

	//Non-JSON-safe variables omit them from the export
	curText        string
	animating      bool
	skipAnim       bool
	animatedStates []int
}

type dialogRenderer struct {
	objects []fyne.CanvasObject
	dialog  *Dialog
}

func NewJsonSafeDialog() JsonSafeDialog {
	return JsonSafeDialog{
		AllText:           []string{"Lorem Ipsum"},
		State:             0,
		MaxState:          0,
		StateOnTap:        true,
		TextOnStateChange: true,
		ConcatText:        false,
		AnimateText:       true,
		CanSkip:           true,
		TextDelay:         25,
		Name:              "",
		ContentStyle:      NFWidget.NewTextStyling(),
		Sizing:            NFWidget.NewSizing(100, 100, 200, 200, true, false),
		Padding:           NFWidget.NewPadding(5, 5, 5, 5),
		ExternalPadding:   NFWidget.NewPadding(5, 5, 5, 5),
		StrokeColor:       color.RGBA{A: 255},
		Fill:              color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Stroke:            1,
		HasName:           false,
		NameAffectsLayout: true,
		NameStyle:         NFWidget.NewTextStyling(),
		NamePadding:       NFWidget.NewPadding(1, 1, 1, 1),
		NameSizing:        NFWidget.NewSizing(100, 50, 200, 50, false, false),
		NameStrokeColor:   color.RGBA{A: 255},
		NameFill:          color.RGBA{B: 255, A: 255},
		NameStroke:        1,
		NamePosition:      fyne.NewPos(200, 50),
	}
}

func NewDialog(hasName bool, text ...string) *Dialog {
	dialog := &Dialog{
		JsonSafeDialog: NewJsonSafeDialog(),
	}
	if hasName {
		dialog.HasName = true
		if len(text) > 0 {
			dialog.Name = text[0]
			dialog.AllText = text[1:]
		}
	} else {
		dialog.AllText = text
	}
	dialog.ExtendBaseWidget(dialog)
	return dialog
}

// Export exports the widget to a json file in the export folder relative to the parent application,
// to allow for use in the editor
func (d *Dialog) Export() error {
	err := NFWidget.WidgetExport(d.JsonSafeDialog, "Dialog")
	if err != nil {
		return err
	}
	return nil
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
	label.TextStyle = d.ContentStyle.TextStyle
	label.Wrapping = d.ContentStyle.Wrapping
	scroll := container.NewScroll(label)
	nameLabel := widget.NewLabel(d.Name)
	nameLabel.TextStyle = d.NameStyle.TextStyle
	nameLabel.Wrapping = d.NameStyle.Wrapping
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
	width := d.Sizing.MinWidth
	height := d.Sizing.MinHeight

	if d.HasName && d.NameAffectsLayout {
		nameSize := d.nameLabel.Size()
		// Initialize calculatedPadding with user-set externalPadding
		calculatedPadding := NFWidget.NewPadding(d.ExternalPadding.Top, d.ExternalPadding.Bottom, d.ExternalPadding.Left, d.ExternalPadding.Right)

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

	dialogSize := fyne.NewSize(min(d.Sizing.MaxWidth, size.Width), min(d.Sizing.MaxHeight, size.Height))
	if d.Sizing.FitWidth {
		dialogSize.Width = size.Width
	}
	if d.Sizing.FitHeight {
		dialogSize.Height = size.Height
	}

	// Adjust the main dialog border size and position
	borderSize := fyne.NewSize(dialogSize.Width-d.ExternalPadding.Horizontal(), dialogSize.Height-d.ExternalPadding.Vertical())
	d.border.Resize(borderSize)
	d.border.Move(fyne.NewPos(d.ExternalPadding.Left, d.ExternalPadding.Top))

	// Configure text label
	d.label.TextStyle = d.ContentStyle.TextStyle
	d.label.Wrapping = d.ContentStyle.Wrapping
	d.border.FillColor = d.Fill
	d.border.StrokeColor = d.StrokeColor
	d.border.StrokeWidth = d.Stroke

	// Configure scroll container
	scrollSize := fyne.NewSize(borderSize.Width-(2*d.Stroke+d.Padding.Horizontal()), borderSize.Height-(2*d.Stroke+d.Padding.Vertical()))
	d.scroll.Resize(scrollSize)
	d.scroll.Move(d.Padding.Origin(0, 0, d.Stroke))

	if d.HasName {
		d.nameBorder.Show()
		d.nameLabel.Show()
		// Configure Name label and border
		d.nameLabel.TextStyle = d.NameStyle.TextStyle
		d.nameLabel.Wrapping = d.NameStyle.Wrapping
		d.nameBorder.FillColor = d.NameFill
		d.nameBorder.StrokeColor = d.NameStrokeColor
		d.nameBorder.StrokeWidth = d.NameStroke

		nameSize := d.nameLabel.Size()
		if nameSize.Width > d.NameSizing.MaxWidth {
			nameSize.Width = d.NameSizing.MaxWidth
		} else if nameSize.Width < d.NameSizing.MinWidth && d.NameSizing.MinWidth < d.NameSizing.MaxWidth {
			nameSize.Width = d.NameSizing.MinWidth
		}

		if nameSize.Height > d.NameSizing.MaxHeight {
			nameSize.Height = d.NameSizing.MaxHeight
		} else if nameSize.Height < d.NameSizing.MinHeight && d.NameSizing.MinHeight < d.NameSizing.MaxHeight {
			nameSize.Height = d.NameSizing.MinHeight
		}

		d.nameLabel.Resize(nameSize)
		d.nameBorder.Resize(fyne.NewSize(nameSize.Width+d.NamePadding.Left+d.NamePadding.Right, nameSize.Height+d.NamePadding.Top+d.NamePadding.Bottom))

		if d.NameAffectsLayout {
			//Shift struct containing the amount to shift each axis
			shift := fyne.NewPos(0, 0)
			newNamePosition := fyne.NewPos(d.NamePosition.X, d.NamePosition.Y)
			if d.NamePosition.X < 0 {
				shift.X = -d.NamePosition.X
				newNamePosition.X = 0 + d.ExternalPadding.Left
			}
			if d.NamePosition.Y < 0 {
				shift.Y = -d.NamePosition.Y
				newNamePosition.Y = 0 + d.ExternalPadding.Top
			}
			d.nameBorder.Move(newNamePosition)
			d.nameLabel.Move(fyne.NewPos(newNamePosition.X+d.NamePadding.Left, newNamePosition.Y+d.NamePadding.Top))

			//Resize the main dialog border and scrollbox to account for the shifted nameBorder
			d.border.Resize(fyne.NewSize(d.border.Size().Width-shift.X, d.border.Size().Height-shift.Y))
			d.scroll.Resize(fyne.NewSize(d.scroll.Size().Width-shift.X, d.scroll.Size().Height-shift.Y))
			//Move the main dialog border and scrollbox to account for the shifted nameBorder
			d.border.Move(d.border.Position().Add(shift))
			d.scroll.Move(d.scroll.Position().Add(shift))

		} else {
			d.nameBorder.Move(fyne.NewPos(d.NamePosition.X+d.ExternalPadding.Left, d.NamePosition.Y+d.ExternalPadding.Top))
			d.nameLabel.Move(fyne.NewPos(d.NamePosition.X+d.NamePadding.Left+d.ExternalPadding.Left, d.NamePosition.Y+d.NamePadding.Top+d.ExternalPadding.Top))
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
	d.label.TextStyle = d.ContentStyle.TextStyle
	d.label.Wrapping = d.ContentStyle.Wrapping
	d.label.Refresh()
	d.UpdateText()
	d.border.FillColor = d.Fill
	d.border.StrokeColor = d.StrokeColor
	d.border.StrokeWidth = d.Stroke
	d.border.Refresh()
	if d.HasName {
		d.nameLabel.TextStyle = d.NameStyle.TextStyle
		d.nameLabel.Wrapping = d.NameStyle.Wrapping
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
