package EditorWidgets

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"strconv"
)

type TypedEntry struct {
	entryType ValueType
	container *fyne.Container
	valEntry  *widget.Entry
	widget.BaseWidget
}

type TypedEntryRenderer struct {
	entry *TypedEntry
}

func (t TypedEntryRenderer) Destroy() {}

func (t TypedEntryRenderer) Layout(size fyne.Size) {
	t.entry.container.Layout.Layout(t.entry.container.Objects, size)
}

func (t TypedEntryRenderer) MinSize() fyne.Size {
	return t.entry.container.MinSize()
}

func (t TypedEntryRenderer) Objects() []fyne.CanvasObject {
	return t.entry.container.Objects
}

func (t TypedEntryRenderer) Refresh() {
	t.entry.container.Refresh()
}

func (t *TypedEntry) CreateRenderer() fyne.WidgetRenderer {
	return &TypedEntryRenderer{entry: t}
}

func (t *TypedEntry) ParsedValue() (interface{}, error) {
	switch t.Type() {
	case IntType:
		return strconv.Atoi(t.valEntry.Text)
	case FloatType:
		return strconv.ParseFloat(t.valEntry.Text, 64)
	case BooleanType:
		return strconv.ParseBool(t.valEntry.Text)
	case StringType:
		return t.valEntry.Text, nil
	default:
		return nil, errors.New("unknown type")
	}
}

func (t *TypedEntry) Type() ValueType {
	return t.entryType
}

func (t *TypedEntry) SetType(entryType ValueType) {
	t.entryType = entryType
	switch entryType {
	case IntType:
		t.valEntry.Validator = func(s string) error {
			_, err := strconv.Atoi(s)
			return err
		}
	case FloatType:
		t.valEntry.Validator = func(s string) error {
			_, err := strconv.ParseFloat(s, 64)
			return err
		}
	case BooleanType:
		t.valEntry.Validator = func(s string) error {
			_, err := strconv.ParseBool(s)
			return err
		}
	case StringType:
		t.valEntry.Validator = func(s string) error {
			return nil
		}
	default:
		t.valEntry.Validator = nil
	}
}

func (t *TypedEntry) Text() string {
	return t.valEntry.Text
}

func (t *TypedEntry) SetText(val string) {
	t.valEntry.SetText(val)
}

func NewTypedEntry(t ValueType) *TypedEntry {
	entry := &TypedEntry{
		entryType: t,
		valEntry:  widget.NewEntry(),
		container: container.NewStack(),
	}
	entry.ExtendBaseWidget(entry)
	entry.container.Add(entry.valEntry)
	return entry
}

func NewTypedEntryWithText(t ValueType, text string) *TypedEntry {
	entry := NewTypedEntry(t)
	entry.valEntry.SetText(text)
	return entry
}
