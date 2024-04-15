package EditorWidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"strconv"
)

type TypedEntryRenderer struct {
	entry   *TypedEntry
	objects []fyne.CanvasObject
}

func (t TypedEntryRenderer) Destroy() {}

func (t TypedEntryRenderer) Layout(size fyne.Size) {
	t.entry.scroll.SetMinSize(t.MinSize())
	t.entry.scroll.Resize(size)
}

func (t TypedEntryRenderer) MinSize() fyne.Size {
	if !t.entry.editing {
		return t.entry.object.MinSize()
	} else {
		switch t.entry.valType {
		case IntType, FloatType, BooleanType, StringType:
			return t.entry.object.MinSize()
		default:
			return fyne.NewSize(200, 500)
		}
	}
}

func (t TypedEntryRenderer) Objects() []fyne.CanvasObject {
	return t.objects
}

func (t TypedEntryRenderer) Refresh() {
	t.entry.scroll.Refresh()
	t.Layout(t.MinSize())
}

type TypedEntry struct {
	object  fyne.CanvasObject
	val     interface{}
	valType ValueType
	editing bool
	scroll  *container.Scroll
	widget.BaseWidget
}

func (t *TypedEntry) CreateRenderer() fyne.WidgetRenderer {
	return &TypedEntryRenderer{entry: t}

}

func NewTypedEntry(val interface{}) *TypedEntry {
	entry := &TypedEntry{
		object: widget.NewLabel("Unknown"),
	}
	entry.scroll = container.NewScroll(entry.object)
	entry.SetVal(val)
	entry.ExtendBaseWidget(entry)
	return entry
}

func (t *TypedEntry) SetVal(val interface{}) {
	t.val = val
	t.valType, _ = DetectValueType(val)
	entry := t.createVal()
	t.object = entry
	t.Refresh()
}

func (t *TypedEntry) SetEditing(editing bool) {
	t.editing = editing
	t.SetVal(t.val)
}

func (t *TypedEntry) createVal() fyne.CanvasObject {
	valType, valString := DetectValueType(t.val)
	if t.editing {
		switch valType {
		case SliceType:
			valSlice := t.val.([]interface{})
			var tree *widget.Tree
			tree = widget.NewTree(
				func(id widget.TreeNodeID) []widget.TreeNodeID {
					switch id {
					case "":
						return []widget.TreeNodeID{"-1"}
					case "-1":
						c := make([]widget.TreeNodeID, len(valSlice))
						for i := range valSlice {
							c[i] = strconv.Itoa(i)
						}
						return c
					}
					return []string{}
				},
				func(id string) bool {
					return id == "-1"
				},
				func(b bool) fyne.CanvasObject {
					if b {
						return container.NewHBox(widget.NewLabel("Slice..."), widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {}))
					} else {
						return widget.NewLabel("Unknown")
					}
				},
				func(id string, b bool, obj fyne.CanvasObject) {
					if b {
						addButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
							valSlice = append(valSlice, "")
							tree.Refresh()
						})
						obj.(*fyne.Container).Objects[1] = addButton
						return
					}
					i, err := strconv.Atoi(id)
					if err != nil {
						log.Println(err)
					}
					obj = NewTypedEntry(valSlice[i])
				},
			)
			tree.Root = "-1"
			return tree
		case StringType, IntType, FloatType, BooleanType:
			entry := widget.NewEntry()
			entry.SetText(valString)
			entry.Validator = SelectValidator(valType)
			return entry
		default:
			return widget.NewLabel("Unknown")
		}
	} else {
		return NewTypedLabel(t.val)
	}
}

func SelectValidator(valType ValueType) fyne.StringValidator {
	switch valType {
	case IntType:
		return func(s string) error {
			_, err := strconv.Atoi(s)
			return err
		}
	case FloatType:
		return func(s string) error {
			_, err := strconv.ParseFloat(s, 64)
			return err
		}
	case BooleanType:
		return func(s string) error {
			_, err := strconv.ParseBool(s)
			return err
		}
	default:
		return nil
	}
}
