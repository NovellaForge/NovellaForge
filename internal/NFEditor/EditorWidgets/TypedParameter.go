package EditorWidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"strings"
)

type SupportedType string

/*
TODO:
 [ ] Array support
 [ ] Map support
 [ ] Custom type support (Must be
*/

const (
	// String is a string type
	String SupportedType = "string"
	// Int is an integer type
	Int SupportedType = "int"
	// Float is a float type
	Float SupportedType = "float"
	// Bool is a boolean type
	Bool SupportedType = "bool"
)

// TypedParameter is a widget that represents a parameter with a specific type
type TypedParameter struct {
	widget.BaseWidget
	window       fyne.Window
	keyEntry     *widget.Entry
	valueEntry   fyne.CanvasObject
	longEntry    bool
	selectedType SupportedType
}

func (t *TypedParameter) LongEntry() (bool, error) {
	return t.longEntry, nil
}

func (t *TypedParameter) SetLongEntry(longEntry bool) {
	t.longEntry = longEntry
}

func (t *TypedParameter) GetKey() string {
	return t.keyEntry.Text
}

func (t *TypedParameter) GetValue() interface{} {
	return t.castType(t.valueEntry)
}

func (t *TypedParameter) SetType(isType SupportedType) {
	t.selectedType = isType
}

func (t *TypedParameter) GetType() SupportedType {
	return t.selectedType
}

type typedParameterRenderer struct {
	objects        []fyne.CanvasObject
	TypedParameter *TypedParameter
}

func (t typedParameterRenderer) Destroy() {
	//IF needed clean up go routines or other resources
	return
}

func (t typedParameterRenderer) Layout(size fyne.Size) {
	grid := container.NewGridWithColumns(3)
	t.TypedParameter.keyEntry.SetPlaceHolder("Key")
	vbox := container.NewVBox()
	switch valueField := t.TypedParameter.valueEntry.(type) {
	case *widget.Entry:
		valueField.SetPlaceHolder("Value")
		if t.TypedParameter.longEntry {
			valueField.SetMinRowsVisible(5)
		} else {
			valueField.SetMinRowsVisible(1)
		}
		longEntryButton := widget.NewButton("Expand Entry", func() {
			t.TypedParameter.longEntry = !t.TypedParameter.longEntry
			if t.TypedParameter.longEntry {
				t.TypedParameter.valueEntry.(*widget.Entry).MultiLine = true
			} else {
				t.TypedParameter.valueEntry.(*widget.Entry).MultiLine = false
			}
		})
		vbox.Add(longEntryButton)
	case *fyne.Container:
		//TODO: Add reference and binding support
		//TODO Add custom Object widget that can allow object creation

	}

	hbox := container.NewHBox()
	settingsButton := widget.NewButtonWithIcon("", theme.MenuIcon(), func() {
		var menuPopup *widget.PopUp
		closeButton := widget.NewButton("Close", func() {
			menuPopup.Hide()
		})
		vbox.Add(closeButton)
		menuPopup = widget.NewPopUp(vbox, t.TypedParameter.window.Canvas())
		menuPopup.Show()
	})
	hbox.Add(settingsButton)
	hbox.Add(layout.NewSpacer())
	grid.Add(t.TypedParameter.keyEntry)
	grid.Add(t.TypedParameter.valueEntry)
	grid.Add(hbox)
	grid.Resize(size)
	t.objects = []fyne.CanvasObject{grid}
}

func (t typedParameterRenderer) MinSize() fyne.Size {
	return t.Objects()[0].MinSize()
}

func (t typedParameterRenderer) Objects() []fyne.CanvasObject {
	return t.objects
}

func (t typedParameterRenderer) Refresh() {
	t.Layout(t.MinSize())
}

func (t *TypedParameter) CreateRenderer() fyne.WidgetRenderer {
	return &typedParameterRenderer{TypedParameter: t}
}

func (t *TypedParameter) castType(value interface{}) interface{} {
	text := ""
	switch v := value.(type) {
	case *widget.Entry:
		text = v.Text
	}
	switch t.selectedType {
	case Int:
		//Int from string
		i, err := strconv.Atoi(text)
		if err != nil {
			return nil
		}
		return i
	case Float:
		f, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil
		}
		return f
	case Bool:
		b, err := strconv.ParseBool(text)
		if err != nil {
			return nil
		}
		return b
	case String:
		return text
	}
	return nil
}

// NewTypedParameter creates a new TypedParameter widget
func NewTypedParameter(key string, value string, selectedType SupportedType, window fyne.Window) *TypedParameter {
	tp := &TypedParameter{
		keyEntry:     widget.NewEntry(),
		valueEntry:   widget.NewEntry(),
		selectedType: selectedType,
	}
	tp.ExtendBaseWidget(tp)
	tp.keyEntry.SetText(key)
	//If the value contains a newline character, set the long entry to true
	if len(value) > 0 {
		if strings.Contains(value, "\n") {
			tp.longEntry = true
			tp.valueEntry.(*widget.Entry).MultiLine = true
		} else {
			tp.longEntry = false
			tp.valueEntry.(*widget.Entry).MultiLine = false
		}
	}
	tp.valueEntry.(*widget.Entry).SetText(value)
	return tp
}

// NewEmptyTypedParameter creates a new TypedParameter widget with empty values
func NewEmptyTypedParameter(isType SupportedType, window fyne.Window) *TypedParameter {
	return NewTypedParameter("", "", isType, window)
}
