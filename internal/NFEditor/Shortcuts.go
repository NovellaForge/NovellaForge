package NFEditor

import "fyne.io/fyne/v2"

type CustomShortcut struct {
	name string
	key  fyne.KeyName
	mod  fyne.KeyModifier
}

func NewCustomShortcut(name string, key fyne.KeyName, mod fyne.KeyModifier) *CustomShortcut {
	return &CustomShortcut{
		name: name,
		key:  key,
		mod:  mod,
	}
}

func (c CustomShortcut) ShortcutName() string {
	return c.name
}

func (c CustomShortcut) Key() fyne.KeyName {
	return c.key
}

func (c CustomShortcut) Mod() fyne.KeyModifier {
	return c.mod
}
