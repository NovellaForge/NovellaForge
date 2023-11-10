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

type SafeNarrativeBox struct {
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

// NarrativeBox is a widget that displays a message to the user while also having interactive elements
type NarrativeBox struct {
	//Interfaces for the dialog
	widget.BaseWidget
	widget.DisableableWidget

	//JSON-safe variables for the dialog and use in the editor
	SafeNarrativeBox

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

type renderer struct {
	objects      []fyne.CanvasObject
	narrativeBox *NarrativeBox
}

func NewJsonSafeDialog() SafeNarrativeBox {
	return SafeNarrativeBox{
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

func NewNarrativeBox(hasName bool, text ...string) *NarrativeBox {
	n := &NarrativeBox{
		SafeNarrativeBox: NewJsonSafeDialog(),
	}
	if hasName {
		n.HasName = true
		if len(text) > 0 {
			n.Name = text[0]
			n.AllText = text[1:]
		}
	} else {
		n.AllText = text
	}
	n.ExtendBaseWidget(n)
	return n
}

// Export exports the widget to a json file in the export folder relative to the parent application,
// to allow for use in the editor
func (n *NarrativeBox) Export() error {
	err := NFWidget.WidgetExport(n.SafeNarrativeBox, "NarrativeBox")
	if err != nil {
		return err
	}
	return nil
}

func (n *NarrativeBox) CreateRenderer() fyne.WidgetRenderer {
	debugBorder := canvas.NewRectangle(color.RGBA{R: 255, A: 255})
	debugBorder.StrokeColor = color.RGBA{R: 255, A: 255}
	debugBorder.StrokeWidth = 1
	n.debug = debugBorder

	border := canvas.NewRectangle(color.Transparent)
	border.FillColor = n.Fill
	border.StrokeColor = n.StrokeColor
	border.StrokeWidth = n.Stroke
	label := widget.NewLabel("")
	label.TextStyle = n.ContentStyle.TextStyle
	label.Wrapping = n.ContentStyle.Wrapping
	scroll := container.NewScroll(label)
	nameLabel := widget.NewLabel(n.Name)
	nameLabel.TextStyle = n.NameStyle.TextStyle
	nameLabel.Wrapping = n.NameStyle.Wrapping
	nameBorder := canvas.NewRectangle(color.Transparent)
	nameBorder.FillColor = n.NameFill
	nameBorder.StrokeColor = n.NameStrokeColor
	nameBorder.StrokeWidth = n.NameStroke
	n.border = border
	n.label = label
	n.scroll = scroll
	n.nameLabel = nameLabel
	n.nameBorder = nameBorder
	//objects := []fyne.CanvasObject{debugBorder, border, scroll, nameLabel, nameBorder}
	objects := []fyne.CanvasObject{border, scroll, nameLabel, nameBorder}
	r := &renderer{
		objects:      objects,
		narrativeBox: n,
	}
	return r
}

func (r *renderer) MinSize() fyne.Size {
	n := r.narrativeBox
	width := n.Sizing.MinWidth
	height := n.Sizing.MinHeight

	if n.HasName && n.NameAffectsLayout {
		nameSize := n.nameLabel.Size()
		// Initialize calculatedPadding with user-set externalPadding
		calculatedPadding := NFWidget.NewPadding(n.ExternalPadding.Top, n.ExternalPadding.Bottom, n.ExternalPadding.Left, n.ExternalPadding.Right)

		// Only add padding to the sides that nameBorder extends over
		if n.NamePosition.X < 0 {
			calculatedPadding.Left = max(calculatedPadding.Left, -n.NamePosition.X)
		}
		if n.NamePosition.Y < 0 {
			calculatedPadding.Top = max(calculatedPadding.Top, -n.NamePosition.Y)
		}
		if n.NamePosition.X+nameSize.Width > width {
			calculatedPadding.Right = max(calculatedPadding.Right, n.NamePosition.X+nameSize.Width-width)
		}
		if n.NamePosition.Y+nameSize.Height > height {
			calculatedPadding.Bottom = max(calculatedPadding.Bottom, n.NamePosition.Y+nameSize.Height-height)
		}

		// Adjust the MinSize based on the effective externalPadding
		width += calculatedPadding.Left + calculatedPadding.Right
		height += calculatedPadding.Top + calculatedPadding.Bottom
	}

	return fyne.NewSize(width, height)
}

func (r *renderer) Layout(size fyne.Size) {
	n := r.narrativeBox

	n.debug.Resize(size)
	n.debug.Move(fyne.NewPos(0, 0))

	dialogSize := fyne.NewSize(min(n.Sizing.MaxWidth, size.Width), min(n.Sizing.MaxHeight, size.Height))
	if n.Sizing.FitWidth {
		dialogSize.Width = size.Width
	}
	if n.Sizing.FitHeight {
		dialogSize.Height = size.Height
	}

	// Adjust the main dialog border size and position
	borderSize := fyne.NewSize(dialogSize.Width-n.ExternalPadding.Horizontal(), dialogSize.Height-n.ExternalPadding.Vertical())
	n.border.Resize(borderSize)
	n.border.Move(fyne.NewPos(n.ExternalPadding.Left, n.ExternalPadding.Top))

	// Configure text label
	n.label.TextStyle = n.ContentStyle.TextStyle
	n.label.Wrapping = n.ContentStyle.Wrapping
	n.border.FillColor = n.Fill
	n.border.StrokeColor = n.StrokeColor
	n.border.StrokeWidth = n.Stroke

	// Configure scroll container
	scrollSize := fyne.NewSize(borderSize.Width-(2*n.Stroke+n.Padding.Horizontal()), borderSize.Height-(2*n.Stroke+n.Padding.Vertical()))
	n.scroll.Resize(scrollSize)
	n.scroll.Move(n.Padding.Origin(0, 0, n.Stroke))

	if n.HasName {
		n.nameBorder.Show()
		n.nameLabel.Show()
		// Configure Name label and border
		n.nameLabel.TextStyle = n.NameStyle.TextStyle
		n.nameLabel.Wrapping = n.NameStyle.Wrapping
		n.nameBorder.FillColor = n.NameFill
		n.nameBorder.StrokeColor = n.NameStrokeColor
		n.nameBorder.StrokeWidth = n.NameStroke

		nameSize := n.nameLabel.Size()
		if nameSize.Width > n.NameSizing.MaxWidth {
			nameSize.Width = n.NameSizing.MaxWidth
		} else if nameSize.Width < n.NameSizing.MinWidth && n.NameSizing.MinWidth < n.NameSizing.MaxWidth {
			nameSize.Width = n.NameSizing.MinWidth
		}

		if nameSize.Height > n.NameSizing.MaxHeight {
			nameSize.Height = n.NameSizing.MaxHeight
		} else if nameSize.Height < n.NameSizing.MinHeight && n.NameSizing.MinHeight < n.NameSizing.MaxHeight {
			nameSize.Height = n.NameSizing.MinHeight
		}

		n.nameLabel.Resize(nameSize)
		n.nameBorder.Resize(fyne.NewSize(nameSize.Width+n.NamePadding.Left+n.NamePadding.Right, nameSize.Height+n.NamePadding.Top+n.NamePadding.Bottom))

		if n.NameAffectsLayout {
			//Shift struct containing the amount to shift each axis
			shift := fyne.NewPos(0, 0)
			newNamePosition := fyne.NewPos(n.NamePosition.X, n.NamePosition.Y)
			if n.NamePosition.X < 0 {
				shift.X = -n.NamePosition.X
				newNamePosition.X = 0 + n.ExternalPadding.Left
			}
			if n.NamePosition.Y < 0 {
				shift.Y = -n.NamePosition.Y
				newNamePosition.Y = 0 + n.ExternalPadding.Top
			}
			n.nameBorder.Move(newNamePosition)
			n.nameLabel.Move(fyne.NewPos(newNamePosition.X+n.NamePadding.Left, newNamePosition.Y+n.NamePadding.Top))

			//Resize the main dialog border and scrollbox to account for the shifted nameBorder
			n.border.Resize(fyne.NewSize(n.border.Size().Width-shift.X, n.border.Size().Height-shift.Y))
			n.scroll.Resize(fyne.NewSize(n.scroll.Size().Width-shift.X, n.scroll.Size().Height-shift.Y))
			//Move the main dialog border and scrollbox to account for the shifted nameBorder
			n.border.Move(n.border.Position().Add(shift))
			n.scroll.Move(n.scroll.Position().Add(shift))

		} else {
			n.nameBorder.Move(fyne.NewPos(n.NamePosition.X+n.ExternalPadding.Left, n.NamePosition.Y+n.ExternalPadding.Top))
			n.nameLabel.Move(fyne.NewPos(n.NamePosition.X+n.NamePadding.Left+n.ExternalPadding.Left, n.NamePosition.Y+n.NamePadding.Top+n.ExternalPadding.Top))
		}
	} else {
		n.nameBorder.Hide()
		n.nameLabel.Hide()
	}
}

func (r *renderer) Refresh() {
	n := r.narrativeBox
	if n.TextOnStateChange {
		if len(n.AllText) == 0 {
			n.AllText = append(n.AllText, "")
		}
		n.MaxState = len(n.AllText) - 1
		if n.State > n.MaxState {
			n.State = n.MaxState
		} else if n.State < 0 {
			n.State = 0
		}
	}
	n.label.TextStyle = n.ContentStyle.TextStyle
	n.label.Wrapping = n.ContentStyle.Wrapping
	n.label.Refresh()
	n.UpdateText()
	n.border.FillColor = n.Fill
	n.border.StrokeColor = n.StrokeColor
	n.border.StrokeWidth = n.Stroke
	n.border.Refresh()
	if n.HasName {
		n.nameLabel.TextStyle = n.NameStyle.TextStyle
		n.nameLabel.Wrapping = n.NameStyle.Wrapping
		n.nameBorder.FillColor = n.NameFill
		n.nameBorder.StrokeColor = n.NameStrokeColor
		n.nameBorder.StrokeWidth = n.NameStroke
		n.nameBorder.Refresh()
		n.nameLabel.Refresh()
	}
	n.scroll.Refresh()
}

func (r *renderer) ApplyTheme() {}

func (r *renderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *renderer) Destroy() {}

func (r *renderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// UpdateText updates the text displayed by the dialog
func (n *NarrativeBox) UpdateText(args ...int) {
	stringToDisplay := ""
	var newContent []string
	if n.ConcatText {
		for i := 0; i < len(n.AllText); i++ {
			if i == 0 {
				newContent = append(newContent, n.AllText[i])
			} else {
				newContent = append(newContent, strings.Join(n.AllText[:i], "\n\n"))
				newContent = append(newContent, n.AllText[i])
			}
		}
	}
	if n.TextOnStateChange {
		if n.ConcatText {
			stringToDisplay = newContent[n.State]
		} else {
			stringToDisplay = n.AllText[n.State]
		}
	} else if len(args) > 0 {
		if n.ConcatText {
			stringToDisplay = newContent[args[0]]
		} else {
			stringToDisplay = n.AllText[args[0]]
		}
	}
	//check if the current state is in animatedStates
	isAnimated := false
	for _, state := range n.animatedStates {
		if state == n.State {
			isAnimated = true
		}
	}
	n.curText = stringToDisplay
	if n.AnimateText && !n.animating && !isAnimated {
		n.animating = true
		n.animatedStates = append(n.animatedStates, n.State)
		go func(d *NarrativeBox) {
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
		}(n)
	} else {
		n.SetText(n.curText)
	}
}

func (n *NarrativeBox) stateChanged() {
	if n.State > n.MaxState {
		n.MaxState = n.State
	} else if n.TextOnStateChange {
		n.UpdateText()
		if n.OnStateChange != nil {
			n.OnStateChange(n.State)
		}
	}
}

func (n *NarrativeBox) Tapped(*fyne.PointEvent) {
	if n.Disabled() {
		return
	}
	if n.StateOnTap && !n.animating {
		n.AddState()
	} else if n.animating && n.CanSkip {
		n.skipAnim = true
	}

	if n.OnTapped != nil {
		n.OnTapped()
	}
}

func (n *NarrativeBox) TappedSecondary(*fyne.PointEvent) {
	if n.Disabled() {
		return
	}
	if n.OnSecondaryTapped != nil {
		n.OnSecondaryTapped()
	}
}

func (n *NarrativeBox) DoubleTapped(*fyne.PointEvent) {
	if n.Disabled() {
		return
	}
	if n.OnDoubleTapped != nil {
		n.OnDoubleTapped()
	}
}

func (n *NarrativeBox) MouseIn(*desktop.MouseEvent) {
	if n.Disabled() {
		return
	}
	if n.OnHover != nil {
		n.OnHover()
	}
}

func (n *NarrativeBox) MouseMoved(*desktop.MouseEvent) {
	if n.Disabled() {
		return
	}
	if n.WhileHover != nil {
		n.WhileHover()
	}
}

func (n *NarrativeBox) MouseOut() {
	if n.Disabled() {
		return
	}
	if n.OnEndHover != nil {
		n.OnEndHover()
	}
}

//Getters and Setters

func (n *NarrativeBox) SetStateOnTap(stateOnTap bool) {
	n.StateOnTap = stateOnTap
	n.Refresh()
}

func (n *NarrativeBox) SetTextOnStateChange(textOnStateChange bool) {
	n.TextOnStateChange = textOnStateChange
	n.Refresh()
}

func (n *NarrativeBox) SetConcatText(concatText bool) {
	n.ConcatText = concatText
	n.Refresh()
}

func (n *NarrativeBox) SetAnimateText(animateText bool) {
	n.AnimateText = animateText
	n.Refresh()
}

func (n *NarrativeBox) SetTextAnimDelay(textAnimDelay float32) {
	n.TextDelay = textAnimDelay
	n.Refresh()
}

func (n *NarrativeBox) SetText(text string) {
	n.label.SetText(text)
}

func (n *NarrativeBox) SetName(name string) {
	n.Name = name
	n.Refresh()
}

func (n *NarrativeBox) SetHasName(hasName bool) {
	n.HasName = hasName
	n.Refresh()
}

func (n *NarrativeBox) SetStrokeColor(strokeColor color.Color) {
	n.StrokeColor = strokeColor
	n.Refresh()
}

func (n *NarrativeBox) SetNameStrokeColor(nameStrokeColor color.Color) {
	n.NameStrokeColor = nameStrokeColor
	n.Refresh()
}

func (n *NarrativeBox) SetFill(fill color.Color) {
	n.Fill = fill
	n.Refresh()
}

func (n *NarrativeBox) SetNameFill(nameFill color.Color) {
	n.NameFill = nameFill
	n.Refresh()
}

func (n *NarrativeBox) SetStroke(stroke float32) {
	n.Stroke = stroke
	n.Refresh()
}

func (n *NarrativeBox) SetNameStroke(nameStroke float32) {
	n.NameStroke = nameStroke
	n.Refresh()
}

func (n *NarrativeBox) SetNamePosition(namePosition fyne.Position) {
	n.NamePosition = namePosition
	n.Refresh()
}

func (n *NarrativeBox) SetNameAffectsLayout(nameAffectsLayout bool) {
	n.NameAffectsLayout = nameAffectsLayout
	n.Refresh()
}

// SetContent sets the AllText field of the dialog; however, it does not update the label, UpdateText() must be called to do that
func (n *NarrativeBox) SetContent(content []string) {
	n.AllText = content
	n.Refresh()
}

// AddContent adds a string to the AllText field of the dialog; however, it does not update the displayed text, UpdateText() must be called to do that
func (n *NarrativeBox) AddContent(content string) {
	n.AllText = append(n.AllText, content)
	n.Refresh()
}

func (n *NarrativeBox) SetState(state int) {
	n.State = state
	n.stateChanged()
}

func (n *NarrativeBox) AddState() {
	n.State++
	n.stateChanged()

}

func (n *NarrativeBox) ChangeState(incr int) {
	n.State += incr
	n.stateChanged()
}
