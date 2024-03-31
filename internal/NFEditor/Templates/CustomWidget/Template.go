package CustomWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFFunction"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
	"log"
)

// Import is a function used to allow importing of the custom widget package
// while still allowing the package to be used normally and not be yelled at by the compiler
func Import() {}

func init() {
	log.Printf("Registering ExampleWidget")
	NFWidget.Widget{
		Type: "ExampleWidget",
		RequiredArgs: NFData.NewNFInterface(
			NFData.NewKeyVal("action", ""),
			NFData.NewKeyVal("message", ""),
		),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(ExampleWidgetHandler)
}

func ExampleWidgetHandler(window fyne.Window, args NFData.NFInterface, w NFWidget.Widget) (fyne.CanvasObject, error) {
	//Get the action from the args
	var action string
	err := args.Get("action", &action)
	if err != nil {
		return nil, err
	}

	var message string
	err = args.Get("message", &message)
	if err != nil {
		return nil, err
	}

	button := widget.NewButton("Example Button", func() {
		//Do something
		_, _ = NFFunction.ParseAndRun(window, action, NFData.NewNFInterface(NFData.NewKeyVal("message", message)))
	})
	return button, nil
}
