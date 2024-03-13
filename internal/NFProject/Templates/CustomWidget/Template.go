package CustomWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/pkg/NFFunction"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
)

func init() {
	NFWidget.Register("ExampleWidget", ExampleWidgetHandler)
}

func ExampleWidgetHandler(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error) {
	//Get the action from the args
	action := args[0].(map[string]interface{})["action"].(string)
	message := args[0].(map[string]interface{})["message"].(string)

	button := widget.NewButton("Example Button", func() {
		//Do something
		_, _, _ = NFFunction.ParseAndRun(window, action, map[string]interface{}{"message": message})
	})
	return button, nil
}
