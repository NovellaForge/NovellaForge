package NFFunction

import (
	"encoding/json"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
)

type Function struct {
	//Name is the name of the function for later reference in editing
	Name string `json:"Name"`
	//Type is the type of function that is used to parse the function this should be Globally Unique, so when making
	//custom ones prefix it with your package name like "MyPackage.MyFunction"
	Type string `json:"Type"`
	//RequiredArgs is a list of arguments that are required for the function to run this should be a map of strings to an interface that can be parsed to determine the type
	RequiredArgs map[string]interface{} `json:"RequiredArgs"`
	//OptionalArgs is a list of arguments that are optional for the function to run this should be a map of strings to an interface that can be parsed to determine the type
	OptionalArgs map[string]interface{} `json:"OptionalArgs"`
}

// Export is a function that is used to export the function to a json file
// These files are used in the main editor to determine inputs needed to call a function in a scene
func (f Function) Export() error {
	jsonBytes, err := json.MarshalIndent(f, "", "	")
	if err != nil {
		return err
	}
	err = os.MkdirAll("exports/functions", 0755)
	err = os.WriteFile("exports/functions/"+f.Type+".json", jsonBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

// This map contains a reference to all registered functions, so they can easily be found and
// called, no matter which files they are registered in
var functions = map[string]functionHandler{}

// All functions are handled in a standard way, with two inputs, a window and a map of arguments
// To determine what goes into the map of arguments, you can look at the RequiredArgs and OptionalArgs fields of the Function struct
// These are also exported to a json file in the exports/functions directory
// The functionHandler returns a map of strings to interfaces and a map of strings to fyne.CanvasObjects
type functionHandler func(window fyne.Window, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error)

// ParseAndRun parses a function from its string name runs it and returns the results
func ParseAndRun(window fyne.Window, function string, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error) {
	//Set the window in the args
	args["window"] = window
	if handler, ok := functions[function]; ok {
		return handler(window, args)
	} else {
		return nil, nil, NFError.ErrNotImplemented
	}
}

// Register adds a custom function to the Functions map and regenerates the function json file
func Register(function Function, handler functionHandler) {
	//Check if the name is already registered, if it is, return
	if _, ok := functions[function.Type]; ok {
		return
	}
	functions[function.Type] = handler
	err := function.Export()
	if err != nil {
		log.Println(err)
	}
}
