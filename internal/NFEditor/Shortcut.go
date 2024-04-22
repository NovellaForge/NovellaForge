package NFEditor

import (
	"fyne.io/fyne/v2"
)

type CustomShortcutHandlerFunc func(window fyne.Window, args ...interface{}) error

type CustomShortcut struct {
	shortcutName string
	key          fyne.KeyName
	mod          fyne.KeyModifier
}

func NewCustomShortcut(shortcutName string, key fyne.KeyName, mod fyne.KeyModifier) fyne.Shortcut {
	c := &CustomShortcut{
		shortcutName: shortcutName,
		key:          key,
		mod:          mod,
	}
	return c
}

func (c *CustomShortcut) SetShortcutName(shortcutName string) {
	c.shortcutName = shortcutName
}

func (c *CustomShortcut) SetKey(key fyne.KeyName) {
	c.key = key
}

func (c *CustomShortcut) SetMod(mod fyne.KeyModifier) {
	c.mod = mod
}

func (c *CustomShortcut) ShortcutName() string {
	return c.shortcutName
}

func (c *CustomShortcut) Key() fyne.KeyName {
	return c.key
}

func (c *CustomShortcut) Mod() fyne.KeyModifier {
	return c.mod
}
