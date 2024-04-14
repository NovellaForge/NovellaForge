package EditorWidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
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
	data         *map[string]interface{}
	revertButton *widget.Button
	gridBox      *fyne.Container
	dataBox      *fyne.Container
	buttonBox    *fyne.Container
}

func NewKeyValBox(key string, data *map[string]interface{}) *KeyValBox {
	val := (*data)[key]
	box := &KeyValBox{
		Container:    container.NewPadded(),
		gridBox:      container.NewGridWithColumns(2),
		buttonBox:    container.NewHBox(),
		dataBox:      container.NewGridWithColumns(2),
		key:          key,
		editButton:   widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {}),
		revertButton: widget.NewButtonWithIcon("", theme.CancelIcon(), func() {}),
		keyLabel:     widget.NewLabel(key),
		valLabel:     NewTypedLabel(DetectValueType(val)),
		keyEntry:     widget.NewEntry(),
		valEntry:     NewTypedEntryWithText(DetectValueType(val)),
		isEditing:    false,
		data:         data,
	}
	box.Container.Add(box.gridBox)
	box.editButton.OnTapped = func() {
		if box.isEditing {
			box.Save()
		} else {
			box.StartEditing()
		}
	}
	box.revertButton.OnTapped = func() {
		box.Revert()
	}
	box.keyEntry.SetText(key)
	box.val = box.valEntry.Text()
	box.buttonBox.Add(box.editButton)
	box.dataBox.Add(box.keyLabel)
	box.dataBox.Add(box.valLabel)
	box.gridBox.Add(box.dataBox)
	box.gridBox.Add(box.buttonBox)
	return box
}

func (b *KeyValBox) StartEditing() {
	if !b.isEditing {
		b.isEditing = true
		b.editButton.SetIcon(theme.DocumentSaveIcon())
		b.dataBox.Objects = nil
		b.gridBox.Objects = nil
		b.buttonBox.Objects = nil
		b.buttonBox.Add(b.revertButton)
		b.buttonBox.Add(b.editButton)
		b.dataBox.Add(b.keyEntry)
		b.dataBox.Add(b.valEntry)
		b.gridBox.Add(b.dataBox)
		b.gridBox.Add(b.buttonBox)
	}
}

func (b *KeyValBox) StopEditing(save bool) {
	if b.isEditing {
		b.isEditing = false
		b.editButton.SetIcon(theme.DocumentCreateIcon())
		b.buttonBox.Objects = nil
		b.dataBox.Objects = nil
		b.gridBox.Objects = nil
		b.buttonBox.Add(b.editButton)
		b.dataBox.Add(b.keyLabel)
		b.dataBox.Add(b.valLabel)
		b.gridBox.Add(b.dataBox)
		b.gridBox.Add(b.buttonBox)
		if save {
			if b.key != b.keyEntry.Text {
				//Remove the old key and add the new key
				//Check if the new key already exists
				if _, ok := (*b.data)[b.keyEntry.Text]; ok {
					//dialog.ShowError(errors.New("key already exists"), l.window)
					b.keyEntry.SetText(b.key)
					return
				} else {
					(*b.data)[b.keyEntry.Text] = (*b.data)[b.key]
					delete(*b.data, b.key)
					b.key = b.keyEntry.Text
					b.keyLabel.SetText(b.key)
				}
			}

			//Check if the value has changed
			if b.val != b.valEntry.Text() {
				val, err := b.valEntry.ParsedValue()
				if err != nil {
					log.Println(err)
					b.valEntry.SetText(b.val)
					return
				}
				log.Println("Changes made")
				log.Println("old value: ", b.val)
				log.Println("new value: ", b.valEntry.Text())
				(*b.data)[b.key] = val
				b.val = b.valEntry.Text()
				b.valLabel.SetText(b.val)
			} else {
				log.Println("No changes made")
			}
		} else {
			b.keyEntry.SetText(b.key)
			b.keyLabel.SetText(b.key)
			b.valEntry.SetText(b.val)
			b.valLabel.SetText(b.val)
		}
	}
}

func (b *KeyValBox) Save() {
	b.StopEditing(true)
}

func (b *KeyValBox) Revert() {
	b.StopEditing(false)
}
