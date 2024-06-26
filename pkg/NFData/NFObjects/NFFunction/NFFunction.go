package NFFunction

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"github.com/google/uuid"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// TODO add in a priority field so that action based functions can be run in a specific order

func New(action, functionType string, args *NFData.NFInterfaceMap) *Function {
	return &Function{
		Name:   action,
		UUID:   uuid.New(),
		Action: action,
		Type:   functionType,
		Args:   args,
	}
}

type Function struct {
	// Name is the name of the function for later reference in editing
	Name string `json:"Name"`
	// UUID is the Universally Unique Identifier of the function
	UUID uuid.UUID `json:"UUID"`
	// Action is the action that the function will be called on(OnTapped, OnHover, etc)
	Action string `json:"Action"`
	//Priority is the order in which the function will be called in relation to other action based functions
	Priority int `json:"Priority"`
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

func (f *Function) Validate() error {
	name := f.GetName()
	f.SetName(f.GetName())
	newName := f.GetName()
	var fullErr error
	if name != newName {
		errors.Join(fullErr, NFError.NewErrSceneValidation("Name was changed from "+name+" to "+newName))
	}
	if err := f.CheckArgs(); err != nil {
		errors.Join(fullErr, err)
	}
	return fullErr
}

func (f *Function) MakeId() {
	f.UUID = uuid.New()
}

func (f *Function) FetchIDs() []uuid.UUID {
	return []uuid.UUID{f.UUID}
}

func (f *Function) GetName() string {
	return f.Name
}

func (f *Function) SetName(newName string) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		f.Name = f.Type
	} else {
		f.Name = newName
	}
}

func (f *Function) GetID() uuid.UUID {
	return f.UUID
}

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

func (f *Function) Run(window fyne.Window, newValues *NFData.NFInterfaceMap) (*NFData.NFInterfaceMap, error) {
	return ParseAndRun(window, f.Type, f.Args.Merge(newValues))
}

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
		errors.Join(missArgs, NFError.NewErrMissingArgument(f.GetID().String(), m))
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
func Load(path string) (a NFObjects.AssetProperties, err error) {
	if filepath.Ext(path) != ".NFFunction" {
		return NFObjects.AssetProperties{}, errors.New("invalid file type")
	}
	err = a.Load(path)
	return a, err
}
