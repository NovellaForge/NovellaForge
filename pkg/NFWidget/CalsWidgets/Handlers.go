package CalsWidgets

import "fyne.io/fyne/v2"

// DialogHandler creates a dialog with the given settings
func DialogHandler(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error) {
	fullDialog := NewNarrativeBox(false)
	var dialog SafeNarrativeBox
	for _, value := range args {
		switch value.(type) {
		case SafeNarrativeBox:
			dialog = value.(SafeNarrativeBox)
		}
	}
	fullDialog.SafeNarrativeBox = dialog
	fullDialog.Refresh()

	return fullDialog, nil
}
