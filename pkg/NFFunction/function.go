package NFFunction

import (
	"fyne.io/fyne/v2"
	"github.com/NovellaForge/NovellaForge/pkg/NFError"
)

var defaultFunctions = map[string]functionHandler{
	"Error":        CustomError,
	"NewGame":      NewGame,
	"ContinueGame": ContinueGame,
	"Quit":         Quit,
	"SaveAs":       SaveAs,
}

var customFunctions = map[string]functionHandler{}

type functionHandler func(window fyne.Window, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error)

func FunctionParser(window fyne.Window, function string, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error) {
	//Set the window in the args
	args["window"] = window
	if handler, ok := defaultFunctions[function]; ok {
		return handler(window, args)
	} else if handler, ok = customFunctions[function]; ok {
		return handler(window, args)
	} else {
		return nil, nil, NFError.ErrNotImplemented
	}
}

// Add adds a custom function to the customFunctions map
func Add(name string, handler functionHandler) {
	customFunctions[name] = handler
}
