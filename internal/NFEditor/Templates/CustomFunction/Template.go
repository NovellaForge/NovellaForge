package CustomFunction

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/NovellaForge/NovellaForge/pkg/NFFunction"
	"log"
)

func init() {
	ExampleFunction := NFFunction.Function{
		Name:         "ExampleFunction",
		Type:         "CustomFunction.ExampleFunction",
		RequiredArgs: map[string]interface{}{},
		OptionalArgs: map[string]interface{}{},
	}
	NFFunction.Register(ExampleFunction, ExampleFunctionHandler)
}

func Register() {
	//init is run when the package is imported, so this is just a dummy function to make sure the init function is run
	log.Printf("Registering ExampleFunction")
}

func ExampleFunctionHandler(window fyne.Window, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error) {
	//Do something
	log.Println("Example button was pressed!")
	dialog.ShowInformation("Example", "Example button was pressed!", window)
	return nil, nil, nil
}
