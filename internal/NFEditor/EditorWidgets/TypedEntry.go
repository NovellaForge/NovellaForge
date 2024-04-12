package EditorWidgets

import (
	"errors"
	"fyne.io/fyne/v2/widget"
	"strconv"
)

type TypedEntry struct {
	entryType ValueType
	*widget.Entry
}

func (t *TypedEntry) ParsedValue() (interface{}, error) {
	switch t.Type() {
	case IntType:
		return strconv.Atoi(t.Text)
	case FloatType:
		return strconv.ParseFloat(t.Text, 64)
	case BooleanType:
		return strconv.ParseBool(t.Text)
	case StringType:
		return t.Text, nil
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
		t.Validator = func(s string) error {
			_, err := strconv.Atoi(s)
			return err
		}
	case FloatType:
		t.Validator = func(s string) error {
			_, err := strconv.ParseFloat(s, 64)
			return err
		}
	case BooleanType:
		t.Validator = func(s string) error {
			_, err := strconv.ParseBool(s)
			return err
		}
	case StringType:
		t.Validator = func(s string) error {
			return nil
		}
	default:
		t.Validator = nil
	}
}

func NewTypedEntry(t ValueType) *TypedEntry {
	entry := &TypedEntry{
		entryType: t,
		Entry:     widget.NewEntry(),
	}
	return entry
}

func NewTypedEntryWithText(t ValueType, text string) *TypedEntry {
	entry := &TypedEntry{
		entryType: t,
		Entry:     widget.NewEntry(),
	}
	entry.SetText(text)
	return entry
}
