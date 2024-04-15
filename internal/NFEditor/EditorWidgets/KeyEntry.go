package EditorWidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type KeyEntryRenderer struct {
	objects []fyne.CanvasObject
	entry   *KeyEntry
}

func (k *KeyEntryRenderer) Destroy() {}

func (k *KeyEntryRenderer) Layout(size fyne.Size) {
	k.entry.object.Resize(size)
	scroll := container.NewScroll(k.entry.object)
	scroll.SetMinSize(size)
	k.objects = []fyne.CanvasObject{scroll}
	scroll.Refresh()
}

func (k *KeyEntryRenderer) MinSize() fyne.Size {
	return fyne.NewSize(150, 50)
}

func (k *KeyEntryRenderer) Objects() []fyne.CanvasObject {
	return k.objects
}

func (k *KeyEntryRenderer) Refresh() {
	k.entry.object.Refresh()
	k.Layout(k.MinSize())
}

type KeyEntry struct {
	widget.BaseWidget
	object  fyne.CanvasObject
	key     string
	editing bool
}

func (k *KeyEntry) Key() string {
	return k.key
}

func (k *KeyEntry) Editing() bool {
	return k.editing
}

func (k *KeyEntry) SetKey(key string) {
	k.key = key
	switch obj := k.object.(type) {
	case *widget.Label:
		obj.SetText(key)
	case *widget.Entry:
		obj.SetText(key)
	}
}

func (k *KeyEntry) SetEditing(editing bool) {
	k.editing = editing
	if k.editing {
		k.object = widget.NewEntry()
		k.object.(*widget.Entry).SetText(k.key)
		k.object.(*widget.Entry).OnChanged = func(s string) {
			k.key = s
		}
	} else {
		k.object = widget.NewLabel(k.key)
	}
	k.Refresh()
}

func (k *KeyEntry) CreateRenderer() fyne.WidgetRenderer {
	return &KeyEntryRenderer{entry: k}
}

func NewKeyEntry(key string) *KeyEntry {
	entry := &KeyEntry{
		object:  widget.NewLabel(key),
		key:     key,
		editing: false,
	}
	entry.ExtendBaseWidget(entry)
	return entry
}
