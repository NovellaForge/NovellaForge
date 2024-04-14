package EditorWidgets

import (
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
	container *fyne.Container
	data      *map[string]interface{}
	window    fyne.Window
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
		box := NewKeyValBox(key, data)
		l.container.Add(box)
	}
	l.Refresh()
}

func (l *TypedParamList) CheckAll(window fyne.Window, callBack func(bool)) {
	editingKeys := make([]string, 0)
	for _, object := range l.container.Objects {
		box := object.(*KeyValBox)
		if box.isEditing {
			editingKeys = append(editingKeys, box.key)
		}
	}
	if len(editingKeys) > 0 {
		var saveDialog *dialog.CustomDialog
		contentBox := container.NewVBox()
		buttonBox := container.NewHBox()
		label := widget.NewLabel("You have unsaved changes in the following keys:")
		keyBox := container.NewVBox()
		for _, key := range editingKeys {
			keyBox.Add(widget.NewLabel(key))
		}
		scrollBox := container.NewVScroll(keyBox)
		scrollBox.SetMinSize(fyne.NewSize(200, 100))
		cancelButton := widget.NewButton("Cancel", func() {
			saveDialog.Hide()
			callBack(false)
		})
		saveAllButton := widget.NewButton("SaveAll", func() {
			saveDialog.Hide()
			for _, object := range l.container.Objects {
				box := object.(*KeyValBox)
				if box.isEditing {
					box.Save()
				}
			}
			callBack(true)
		})
		revertAllButton := widget.NewButton("RevertAll", func() {
			saveDialog.Hide()
			for _, object := range l.container.Objects {
				box := object.(*KeyValBox)
				if box.isEditing {
					box.Revert()
				}
			}
			callBack(true)
		})
		buttonBox.Add(cancelButton)
		buttonBox.Add(saveAllButton)
		buttonBox.Add(revertAllButton)
		contentBox.Add(label)
		contentBox.Add(scrollBox)
		contentBox.Add(buttonBox)
		saveDialog = dialog.NewCustomWithoutButtons("Unsaved Changes", contentBox, window)
	}
}
