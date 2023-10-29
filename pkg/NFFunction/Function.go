package NFFunction

import (
	"fyne.io/fyne/v2"
	"github.com/NovellaForge/NovellaForge/pkg/NFError"
	"github.com/NovellaForge/NovellaForge/pkg/NFFunction/DefaultFunctions"
)

var defaultFunctions = map[string]functionHandler{
	"Error":        DefaultFunctions.CustomError,
	"NewGame":      DefaultFunctions.NewGame,
	"ContinueGame": DefaultFunctions.ContinueGame,
	"Quit":         DefaultFunctions.Quit,
	"SaveAs":       DefaultFunctions.SaveAs,
}

var customFunctions = map[string]functionHandler{}

type functionHandler func(window fyne.Window, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error)

// ParseAndRun parses a function from its string name runs it and returns the results
func ParseAndRun(window fyne.Window, function string, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error) {
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

// Register adds a custom function to the customFunctions map
func Register(name string, handler functionHandler) {
	//Check if the name is already registered, if it is, return
	if _, ok := customFunctions[name]; ok {
		return
	}
	customFunctions[name] = handler
}
