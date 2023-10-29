package CalsWidgets

import "fyne.io/fyne/v2"

// DialogHandler creates a dialog with the given settings
func DialogHandler(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error) {
	fullDialog := NewDialog(false)
	var dialog JsonSafeDialog
	for _, value := range args {
		switch value.(type) {
		case JsonSafeDialog:
			dialog = value.(JsonSafeDialog)
		}
	}
	fullDialog.JsonSafeDialog = dialog
	fullDialog.Refresh()

	return fullDialog, nil
}
