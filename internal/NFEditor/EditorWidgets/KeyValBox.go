package EditorWidgets

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
)

type keyValBoxRenderer struct {
	box     *KeyValBox
	objects []fyne.CanvasObject
}

func (k *keyValBoxRenderer) Destroy() {}

func (k *keyValBoxRenderer) Layout(size fyne.Size) {
	box := container.NewWithoutLayout(k.box.editButton, k.box.revertButton, k.box.keyField, k.box.valField)
	box.Resize(size)
	availableSize := fyne.NewSize(size.Width, size.Height)

	k.box.editButton.Resize(k.box.editButton.MinSize())
	k.box.editButton.Move(fyne.NewPos(size.Width-availableSize.Width, 0))
	availableSize.Width -= k.box.editButton.MinSize().Width

	k.box.revertButton.Resize(k.box.revertButton.MinSize())
	k.box.revertButton.Move(fyne.NewPos(size.Width-availableSize.Width, 0))
	availableSize.Width -= k.box.revertButton.MinSize().Width

	k.box.keyField.Resize(k.box.keyField.MinSize())
	k.box.keyField.Move(fyne.NewPos(size.Width-availableSize.Width, 0))
	availableSize.Width -= k.box.keyField.MinSize().Width

	k.box.valField.Resize(k.box.valField.MinSize())
	k.box.valField.Move(fyne.NewPos(size.Width-availableSize.Width, 0))
	availableSize.Width -= k.box.valField.MinSize().Width

	k.objects = []fyne.CanvasObject{box}
}

func (k *keyValBoxRenderer) MinSize() fyne.Size {
	var minWidth float32 = 0
	minWidth += k.box.editButton.MinSize().Width
	minWidth += k.box.revertButton.MinSize().Width
	minWidth += k.box.keyField.MinSize().Width
	minWidth += k.box.valField.MinSize().Width
	var minHeight float32 = 0
	minHeight = max(minHeight, k.box.editButton.MinSize().Height)
	minHeight = max(minHeight, k.box.revertButton.MinSize().Height)
	minHeight = max(minHeight, k.box.keyField.MinSize().Height)
	minHeight = max(minHeight, k.box.valField.MinSize().Height)
	return fyne.NewSize(minWidth, minHeight)
}

func (k *keyValBoxRenderer) Objects() []fyne.CanvasObject {
	return k.objects
}

func (k *keyValBoxRenderer) Refresh() {
	k.Layout(k.MinSize())
}

func (b *KeyValBox) CreateRenderer() fyne.WidgetRenderer {
	return &keyValBoxRenderer{box: b}
}

type KeyValBox struct {
	widget.BaseWidget
	key          string
	val          interface{}
	editButton   *widget.Button
	revertButton *widget.Button
	keyField     *KeyEntry
	valField     *TypedEntry
	isEditing    bool
	data         *map[string]interface{}
	window       fyne.Window
	OnSave       func()
}

func NewKeyValBox(key string, data *map[string]interface{}, window fyne.Window) *KeyValBox {
	val := (*data)[key]
	box := &KeyValBox{
		key:          key,
		editButton:   widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {}),
		revertButton: widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {}),
		keyField:     NewKeyEntry(key),
		valField:     NewTypedEntry(val),
		val:          val,
		isEditing:    false,
		data:         data,
		window:       window,
	}
	box.ExtendBaseWidget(box)
	box.editButton.OnTapped = func() {
		if box.isEditing {
			box.Save()
		} else {
			box.StartEditing()
		}
	}
	box.revertButton.OnTapped = func() {
		if box.isEditing {
			box.Revert()
		} else {
			dialog.ShowConfirm("Confirm Deletion", "Are you sure you want to delete the key: "+box.key, func(response bool) {
				if response {
					delete(*box.data, box.key)
				}
			}, box.window)
		}
	}
	return box
}

func (b *KeyValBox) StartEditing() {
	if !b.isEditing {
		b.isEditing = true
		b.editButton.SetIcon(theme.DocumentSaveIcon())
		b.revertButton.SetIcon(theme.CancelIcon())
		b.valField.SetEditing(true)
	}
}

func (b *KeyValBox) StopEditing(save bool) {
	if b.isEditing {
		b.isEditing = false
		b.editButton.SetIcon(theme.DocumentCreateIcon())
		b.revertButton.SetIcon(theme.DeleteIcon())
		b.valField.SetEditing(false)
		if save {
			if b.key != b.keyField.Key() {
				//Remove the old key and add the new key
				//Check if the new key already exists
				if _, ok := (*b.data)[b.keyField.Key()]; ok {
					dialog.ShowError(errors.New("key already exists"), b.window)
					b.keyField.SetKey(b.key)
					return
				} else {
					(*b.data)[b.keyField.Key()] = (*b.data)[b.key]
					delete(*b.data, b.key)
					b.key = b.keyField.Key()
				}
			}

			//Check if the value has changed
			if b.val != b.valField.val {
				(*b.data)[b.key] = b.valField.val
				b.val = b.valField.val
			} else {
				log.Println("No changes made")
			}
			if b.OnSave != nil {
				b.OnSave()
			}
		} else {
			b.keyField.SetKey(b.key)
			b.valField.SetVal(b.val)
		}
	}
}

func (b *KeyValBox) Save() {
	b.StopEditing(true)
}

func (b *KeyValBox) Revert() {
	b.StopEditing(false)
}
