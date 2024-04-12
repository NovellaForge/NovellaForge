package EditorWidgets

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"sort"
)

type TypedParamListRenderer struct {
	list *TypedParamList
}

func (t TypedParamListRenderer) Destroy() {}

func (t TypedParamListRenderer) Layout(size fyne.Size) {
	t.list.container.Layout.Layout(t.list.container.Objects, size)
}

func (t TypedParamListRenderer) MinSize() fyne.Size {
	return t.list.container.MinSize()
}

func (t TypedParamListRenderer) Objects() []fyne.CanvasObject {
	return t.list.container.Objects
}

func (t TypedParamListRenderer) Refresh() {
	t.list.container.Refresh()
}

type TypedParamList struct {
	widget.BaseWidget
	container  *fyne.Container
	data       *map[string]interface{}
	editingKey string
	window     fyne.Window
}

func (l *TypedParamList) CreateRenderer() fyne.WidgetRenderer {
	return &TypedParamListRenderer{list: l}
}

func NewTypedParamList(data *map[string]interface{}, window fyne.Window) *TypedParamList {
	list := &TypedParamList{
		container: container.NewVBox(),
		data:      data,
		window:    window,
	}
	list.SetData(data)
	list.ExtendBaseWidget(list)
	return list
}

func (l *TypedParamList) SetData(data *map[string]interface{}) {
	l.container.Objects = nil

	// Extract the keys and sort them
	keys := make([]string, 0, len(*data))
	for key := range *data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Iterate over sorted keys
	for _, key := range keys {
		value := (*data)[key]
		box := NewKeyValBox(key, value)
		box.OnEditChange = func(b *KeyValBox) {
			if l.editingKey != b.key {
				l.ChangeEditingKey(l.editingKey, b.key)
			}
		}
		l.container.Add(box)
	}
	l.Refresh()
}

func (l *TypedParamList) ChangeEditingKey(oldKey string, newKey string) {
	//Pop up a dialog asking if the user wants to save the changes made to the old key before switching
	if oldKey != "" {
		dialog.ShowConfirm("Save Changes?", "Do you want to save the changes made to "+oldKey+"?", func(b bool) {
			if b {
				l.SaveChanges(oldKey)
			}
			l.editingKey = newKey
		}, l.window)
	} else {
		l.editingKey = newKey
	}
}

func (l *TypedParamList) SaveChanges(key string) {
	for _, obj := range l.container.Objects {
		box := obj.(*KeyValBox)
		if box.key == key {
			//Check if the key is the same as the entryKey.Text
			if box.key != box.keyEntry.Text {
				//Check if the new key already exists
				if _, ok := (*l.data)[box.keyEntry.Text]; ok {
					dialog.ShowError(errors.New("key already exists"), l.window)
					return
				} else {
					//Delete the old key and add the new key
					(*l.data)[box.keyEntry.Text] = (*l.data)[box.key]
					delete(*l.data, box.key)
					box.key = box.keyEntry.Text
				}
			} else {
				//Check if the value has changed
				if box.val != box.valEntry.Text {
					val, err := box.valEntry.ParsedValue()
					if err != nil {
						dialog.ShowError(err, l.window)
						return
					}
					(*l.data)[box.key] = val
				}
			}
		}
	}
}
