package EditorWidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type keyValBoxRenderer struct {
	box *KeyValBox
}

func (k *keyValBoxRenderer) Destroy() {}

func (k *keyValBoxRenderer) Layout(size fyne.Size) {
	k.box.Container.Layout.Layout(k.box.Container.Objects, size)
}

func (k *keyValBoxRenderer) MinSize() fyne.Size {
	return k.box.Container.MinSize()
}

func (k *keyValBoxRenderer) Objects() []fyne.CanvasObject {
	return k.box.Container.Objects
}

func (k *keyValBoxRenderer) Refresh() {
	k.box.Container.Refresh()
}

func (b *KeyValBox) CreateRenderer() fyne.WidgetRenderer {
	return &keyValBoxRenderer{b}
}

type KeyValBox struct {
	*fyne.Container
	key, val     string
	editButton   *widget.Button
	keyLabel     *widget.Label
	valLabel     *TypedLabel
	keyEntry     *widget.Entry
	valEntry     *TypedEntry
	isEditing    bool
	OnEditChange func(b *KeyValBox)
}

func NewKeyValBox(key string, val interface{}) *KeyValBox {
	box := &KeyValBox{
		Container:  container.NewHBox(),
		key:        key,
		editButton: widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {}),
		keyLabel:   widget.NewLabel(key),
		valLabel:   NewTypedLabel(DetectValueType(val)),
		keyEntry:   widget.NewEntry(),
		valEntry:   NewTypedEntryWithText(DetectValueType(val)),
		isEditing:  false,
	}
	box.editButton.OnTapped = func() {
		if box.isEditing {
			box.Save()
		} else {
			box.StartEditing()
		}
	}
	box.keyEntry.SetText(key)
	box.val = box.valEntry.Text
	box.Container.Add(box.editButton)
	box.Container.Add(box.keyLabel)
	box.Container.Add(box.valLabel)
	return box
}

func (b *KeyValBox) StartEditing() {
	if !b.isEditing {
		b.editButton.SetIcon(theme.DocumentSaveIcon())
		b.Container.Remove(b.keyLabel)
		b.Container.Remove(b.valLabel)
		b.Container.Add(b.keyEntry)
		b.Container.Add(b.valEntry)
		b.isEditing = true
		if b.OnEditChange != nil {
			b.OnEditChange(b)
		}
	}
}

func (b *KeyValBox) StopEditing(save bool) {
	if b.isEditing {
		b.Container.Remove(b.keyEntry)
		b.Container.Remove(b.valEntry)
		b.Container.Add(b.keyLabel)
		b.Container.Add(b.valLabel)
		if save {
			b.key = b.keyEntry.Text
			b.val = b.valEntry.Text
			b.keyLabel.SetText(b.key)
			b.valLabel.SetText(b.val)
		} else {
			b.keyEntry.SetText(b.key)
			b.valEntry.SetText(b.val)
		}
		b.isEditing = false
		if b.OnEditChange != nil {
			b.OnEditChange(b)
		}
	}
}

func (b *KeyValBox) Save() {
	b.StopEditing(true)
}

func (b *KeyValBox) Revert() {
	b.StopEditing(false)
}
