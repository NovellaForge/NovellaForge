package CustomWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/pkg/NFFunction"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
	"log"
)

func init() {
	NFWidget.Register("ExampleWidget", ExampleWidgetHandler)
}

func Register() {
	//init is run when the package is imported, so this is just a dummy function to make sure the init function is run
	log.Printf("Registering ExampleWidget")
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
