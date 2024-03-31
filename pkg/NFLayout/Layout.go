package NFLayout

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
	"log"
	"os"
)

// layoutHandler is a function type that handles the layout logic.
type layoutHandler func(window fyne.Window, args NFData.NFInterface, l Layout) (fyne.CanvasObject, error)

// Layout struct holds all the information about a layout.
//
// Layouts are used to define how widgets are placed in a scene.
//
// They contain a list of widgets that are children of the layout.
//
// They also contain a list of arguments that are required for the layout to run.
type Layout struct {
	Type         string             `json:"Type"`    // Type of the layout
	Children     []NFWidget.Widget  `json:"Widgets"` // List of widgets that are children of the layout
	RequiredArgs NFData.NFInterface `json:"-"`       // List of arguments that are required for the layout to run
	OptionalArgs NFData.NFInterface `json:"-"`       // List of arguments that are optional for the layout to run
	Args         NFData.NFInterface `json:"Args"`    // List of arguments that are passed to the layout
}

// CheckArgs checks if all required arguments are present in Args.
// Returns an error if a required argument is missing.
func (l Layout) CheckArgs() error {
	ok, miss := l.Args.HasAllKeys(l.RequiredArgs)
	if ok {
		return nil
	}
	return NFError.ErrMissingArgument(l.Type, miss...)
}

// layouts is a map that holds all the registered layout handlers.
var layouts = map[string]layoutHandler{}

// Parse checks if the layout type exists in the registered layouts.
// If it does, it uses the assigned handler to parse the layout.
// If it doesn't, it returns an error.
func (l Layout) Parse(window fyne.Window) (fyne.CanvasObject, error) {
	if handler, ok := layouts[l.Type]; ok {
		if err := l.CheckArgs(); err != nil {
			return nil, err
		}
		return handler(window, l.Args, l)
	} else {
		return nil, errors.New("layout unable to be parsed")
	}
}

// Register registers a custom layout handler.
// If the name is already registered, it will not be registered again.
func (l Layout) Register(handler layoutHandler) {
	if _, ok := layouts[l.Type]; ok {
		log.Printf("Layout %s already registered\nIf you are using third party layouts, please yell at the developer to make sure they have properly namespaced their layouts", l.Type)
		return
	}
	layouts[l.Type] = handler
	if ShouldExport {
		err := l.Export()
		if err != nil {
			log.Println("Error exporting layout: " + l.Type)
		}
	}
}

// ExportPath is the path where the layout will be exported.
var ExportPath = "export/layouts"

// ShouldExport is a flag that determines if the layout should be exported.
var ShouldExport = false

// Export exports the layout to a json file.
// These files are used in the main editor to determine inputs needed to create the layout in a scene.
func (l Layout) Export() error {
	lBytes := struct {
		RequiredArgs map[string][]string `json:"RequiredArgs"` // List of required arguments for the layout
		OptionalArgs map[string][]string `json:"OptionalArgs"` // List of optional arguments for the layout
	}{
		RequiredArgs: l.RequiredArgs.Export(),
		OptionalArgs: l.OptionalArgs.Export(),
	}

	jsonBytes, err := json.MarshalIndent(lBytes, "", "  ")
	if err != nil {
		return err
	}

	_, err = os.Stat(ExportPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(ExportPath, 0755)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(ExportPath+"/"+l.Type+".json", jsonBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}
