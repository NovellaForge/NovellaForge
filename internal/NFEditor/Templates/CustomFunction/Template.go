package CustomFunction

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFFunction"
	"log"
)

// Import is a function used to allow importing of the custom function package
// while still allowing the package to be used normally and not be yelled at by the compiler
func Import() {}

func init() {
	log.Printf("Registering ExampleFunction")
	NFFunction.Function{
		Name:         "ExampleFunction",
		Type:         "CustomFunction.ExampleFunction",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(ExampleFunctionHandler)
}

func ExampleFunctionHandler(window fyne.Window, args NFData.NFInterface) (NFData.NFInterface, error) {
	//Do something
	log.Println("Example button was pressed!")
	dialog.ShowInformation("Example", "Example button was pressed!", window)
	return args, nil
}
