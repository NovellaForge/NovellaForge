package CustomFunction

import (
	"fyne.io/fyne/v2"
	. "github.com/NovellaForge/NovellaForge/pkg/NFFunction"
	"log"
)

func init() {
	ExampleFunction := Function{
		Name:         "ExampleFunction",
		Type:         "",
		RequiredArgs: map[string]interface{}{},
		OptionalArgs: map[string]interface{}{},
	}
	Register(ExampleFunction, ExampleFunctionHandler)
}

func ExampleFunctionHandler(window fyne.Window, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error) {
	//Do something
	for argsKey, argsValue := range args {
		log.Printf("Key: %s, Value: %v", argsKey, argsValue)
	}
	return nil, nil, nil
}
