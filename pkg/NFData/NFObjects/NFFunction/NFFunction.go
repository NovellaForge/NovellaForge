package NFFunction

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects"
	"log"
	"os"
	"path/filepath"
)

//TODO turn this into an Object that can be created in the editor and then run in the scene, it will function
// Almost identically to widgets and layouts just without children and it won't return a fyne.CanvasObject
// But it will act as a child to the widget/layout that it is attached to so that it can be given specific traits
// Like running on scene load or just on a button press
// This will include Name validation since it will be used in the editor and will need an ID to be referenced
// Functions will be allowed to be called in two ways, direct parsing via Type, as well as via child invocation
// By specifying which actions the functions will be called on, this will allow for more complex interactions
// For example we could have a function that is called by name, like error. Or we could have a button call its OnTapped
// Which will execute all child functions that have their ActionField set to OnTapped
// Widgets and layouts will be updated to include which action fields they support in their exported types

type Function struct {
	//Name is the name of the function for later reference in editing
	Name string `json:"Name"`
	//Type is the type of function that is used to parse the function this should be Globally Unique, so when making
	//custom ones prefix it with your package name like "MyPackage.MyFunction"
	Type string `json:"Type"`
	//RequiredArgs is a list of arguments that are required for the function to run
	RequiredArgs *NFData.NFInterfaceMap `json:"-"` //This is not exported to json
	//OptionalArgs is a list of arguments that are optional for the function to run
	OptionalArgs *NFData.NFInterfaceMap `json:"-"` //This is not exported to json
	//Args is a list of arguments that are passed to the function through the scene
	Args *NFData.NFInterfaceMap `json:"Args"`
}

func (f *Function) FetchChildren() []NFObjects.NFObject {
	log.Println("Functions cannot have children")
	return nil
}

func (f *Function) DeleteChild(name string) error {
	//Find a child by name and delete it
	return errors.New("functions cannot have children")
}

func (f *Function) AddChild(_ NFObjects.NFObject) {
	log.Println("Functions cannot have children")
}

//Functions for the NFData.NFObject interface

func (f *Function) GetType() string {
	return f.Type
}

func (f *Function) SetType(t string) {
	f.Type = t
}

func (f *Function) GetArgs() *NFData.NFInterfaceMap {
	return f.Args
}

func (f *Function) SetArgs(args *NFData.NFInterfaceMap) {
	f.Args = args
}

//End of Functions for the NFData.NFObject interface

// CheckArgs checks if the function has all the required arguments
func (f *Function) CheckArgs() error {
	info, err := GetFunctionInfo(f.Type)
	if err != nil {
		return err
	}
	f.RequiredArgs = info.RequiredArgs
	ok, miss := f.Args.HasAllKeys(f.RequiredArgs)
	if ok {
		return nil
	}
	var missArgs error
	for _, m := range miss {
		errors.Join(missArgs, NFError.NewErrMissingArgument(f.Name, m))
	}
	return missArgs
}

// SetAndCheckArgs sets the arguments and checks if the function has all the required arguments
func (f *Function) SetAndCheckArgs(args *NFData.NFInterfaceMap) error {
	f.Args = args
	return f.CheckArgs()
}

type functionWithHandler struct {
	Info    Function
	Handler functionHandler
}

// This map contains a reference to all registered functions, so they can easily be found and
// called, no matter which files they are registered in
var functionMap = map[string]functionWithHandler{}

func GetFunctionInfo(function string) (Function, error) {
	if f, ok := functionMap[function]; ok {
		return f.Info, nil
	} else {
		return Function{}, NFError.NewErrNotImplemented("Function Type: " + function + " is not implemented")
	}
}

// All functions are handled in a standard way, with two inputs, a window and a map of arguments
// To determine what goes into the map of arguments, you can look at the RequiredArgs and OptionalArgs fields of the Function struct
// These are also exported to a json file in the exports/functions directory
// The functionHandler returns a map of strings to interfaces and a map of strings to fyne.CanvasObjects
type functionHandler func(window fyne.Window, args *NFData.NFInterfaceMap) (*NFData.NFInterfaceMap, error)

// ParseAndRun parses a function from its string name runs it and returns the results
func ParseAndRun(window fyne.Window, function string, args *NFData.NFInterfaceMap) (*NFData.NFInterfaceMap, error) {
	//Checks if the function is registered, if it is, run it, if not, return an error
	if f, ok := functionMap[function]; ok {
		if err := f.Info.SetAndCheckArgs(args); err != nil {
			return args, err
		}
		return f.Handler(window, args)
	} else {
		return args, NFError.NewErrNotImplemented("Function Type: " + function + " is not implemented")
	}
}

// Register adds a custom function to the Functions map
func (f *Function) Register(handler functionHandler) {
	//Check if the name is already registered, if it is, return
	if _, ok := functionMap[f.Type]; ok {
		log.Println("Error registering function: " + f.Type + " is already registered. If you are using third party functions yell at the developer that they need to properly namespace their functions so they are unique")
		return
	}
	functionMap[f.Type] = functionWithHandler{Info: *f, Handler: handler}
}

// ExportPath is the path where the function will be exported
var ExportPath = "exports/functions"

// ExportRegistered exports all registered functions to json files
func ExportRegistered() {
	for _, function := range functionMap {
		err := function.Info.Export()
		if err != nil {
			log.Println(err)
		}
	}
}

// Export is a function that is used to export the function to a json file
// These files are used in the main editor to determine inputs needed to call a function in a scene
func (f *Function) Export() error {
	//Check if the export path has a trailing slash and add one if it doesn't
	if ExportPath[len(ExportPath)-1] != '/' {
		ExportPath += "/"
	}
	fBytes := struct {
		Type         string              `json:"Type"`
		RequiredArgs map[string][]string `json:"RequiredArgs"`
		OptionalArgs map[string][]string `json:"OptionalArgs"`
	}{
		Type:         f.Type,
		RequiredArgs: f.RequiredArgs.Export(),
		OptionalArgs: f.OptionalArgs.Export(),
	}

	//Export the function to a json file
	jsonBytes, err := json.MarshalIndent(fBytes, "", "  ")
	if err != nil {
		return err
	}

	//Check or make the build/exports directory
	_, err = os.Stat(ExportPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(ExportPath, 0755)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(ExportPath+f.Type+".NFFunction", jsonBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Load is a function that is used to load a function from a .NFFunction file
func Load(path string) (a NFData.AssetProperties, err error) {
	if filepath.Ext(path) != ".NFFunction" {
		return NFData.AssetProperties{}, errors.New("invalid file type")
	}
	err = a.Load(path)
	return a, err
}
