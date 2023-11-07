package NFWidget

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"github.com/NovellaForge/NovellaForge/pkg/NFError"
	"os"
	"reflect"
	"strings"
)

// Widget is the struct that holds all the information about a widget
type Widget struct {
	//ID is the unique ID of the widget for later reference in editing
	ID string `json:"ID"`
	//Type is the type of widget that is used to parse the widget this should be Globally Unique, so when making
	//custom ones prefix it with your package name like "MyPackage.MyWidget"
	Type string `json:"Type"`
	//Children is a list of widgets that are children of this widget
	Children []Widget `json:"Children"`
	//Properties is a map of properties that are used to parse the widget
	Properties map[string]interface{} `json:"Properties"`
	//JsonSafe is a json safe version of the widget being parsed to allow quick setting of properties; this is not required
	JsonSafe interface{} `json:"JsonSafe"`
}

type widgetHandler func(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error)

// Widgets is a map of all the widgets that are registered and can be used by the engine
var Widgets = map[string]widgetHandler{}

// Register adds a custom widget to the customWidgets map
func Register(name string, handler widgetHandler) {
	//Check if the name is already registered
	if _, ok := Widgets[name]; ok {

	}
	Widgets[name] = handler
}

func (w Widget) Parse(window fyne.Window) (fyne.CanvasObject, error) {
	if handler, ok := Widgets[w.Type]; ok {
		return handler(window, w.Properties, w.Children, w.JsonSafe)
	} else {
		return nil, NFError.ErrNotImplemented
	}
}

type TextStyling struct {
	fyne.TextStyle
	Wrapping fyne.TextWrap
}

func NewTextStyling() TextStyling {
	return TextStyling{Wrapping: fyne.TextWrapWord, TextStyle: fyne.TextStyle{Bold: false, Italic: false, Monospace: false, Symbol: false, TabWidth: 4}}
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

func NewSizing(minWidth float32, minHeight float32, maxWidth float32, maxHeight float32, fitWidth bool, fitHeight bool) Sizing {
	return Sizing{MinWidth: minWidth, MinHeight: minHeight, MaxWidth: maxWidth, MaxHeight: maxHeight, FitWidth: fitWidth, FitHeight: fitHeight}
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

func NewPadding(top float32, bottom float32, left float32, right float32) Padding {
	return Padding{Top: top, Bottom: bottom, Left: left, Right: right}
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

// Origin returns the top left combination of the padding, but it can accept float args to offset the origin
//
// One arg will offset the origin by the arg in the x direction,
//
// # Two args will offset the origin by the first arg in the x direction and the second arg in the y direction
//
// # Three args will offset the origin by the first arg in the x direction, the second arg in the y direction, and the third added to both
//
// Example: padding.Origin() will return a fyne.Position with the x value of padding.Left and the y value of padding.Top
//
// Example: padding.Origin(10) will return a fyne.Position with the x value of padding.Left+10 and the y value of padding.Top
//
// Example: padding.Origin(10, 20) will return a fyne.Position with the x value of padding.Left+10 and the y value of padding.Top+20
//
// Example: padding.Origin(10, 20, 30) will return a fyne.Position with the x value of padding.Left+10+30 and the y value of padding.Top+20+30
func (p *Padding) Origin(args ...float32) fyne.Position {
	//Switch on the number of args
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

func (p *Padding) SetOrigin(x, y float32, parent fyne.CanvasObject) {
	p.Left = x
	p.Top = y
	if parent != nil {
		parent.Refresh()
	}
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

func processFields(v reflect.Value) map[string]interface{} {
	t := v.Type()
	fields := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		if value.CanInterface() {
			fields[field.Name] = value.Interface()
		}
	}
	return fields
}

type exportableWidget struct {
	Type     string                 `json:"Type"`
	JsonSafe map[string]interface{} `json:"JsonSafe"`
}

// WidgetExport exports a widget to a json file to allow for use in the editor
func WidgetExport(widget interface{}, widgetType string) error {
	v := reflect.ValueOf(widget)
	jsonSafeMap := processFields(v)
	newWidget := exportableWidget{
		Type:     widgetType,
		JsonSafe: jsonSafeMap,
	}
	//Make a regex to remove '.' and '/' from the widget type
	FileName := strings.ReplaceAll(widgetType, ".", "")
	FileName = strings.ReplaceAll(FileName, "/", "")
	FileName = "export/widgets/" + FileName + ".json"
	bytes, err := json.MarshalIndent(newWidget, "", "  ")
	if err != nil {
		return err
	}

	//Check or make the build/exports directory
	_, err = os.Stat("export/widgets")
	if os.IsNotExist(err) {
		err = os.MkdirAll("export/widgets", 0755)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(FileName, bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}
