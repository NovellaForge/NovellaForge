package NFWidget

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
	"log"
	"os"
)

// Widget is the struct that holds all the information about a widget
type Widget struct {
	// ID is the unique ID of the widget for later reference in editing
	ID string `json:"ID"`
	// Type is the type of widget that is used to parse the widget this should be Globally Unique, so when making
	//custom ones prefix it with your package name like "MyPackage.MyWidget"
	Type string `json:"Type"`
	// Children is a list of widgets that are children of this widget
	Children []Widget `json:"Children"`
	// RequiredArgs is a list of arguments that are required for the widget to run
	RequiredArgs NFData.NFInterface `json:"-"`
	// OptionalArgs is a list of arguments that are optional for the widget to run
	OptionalArgs NFData.NFInterface `json:"-"`
	// Args is a list of arguments that are passed to the widget through the scene
	Args NFData.NFInterface `json:"Args"`
}

func (w Widget) CheckArgs() error {
	ok, miss := w.Args.HasAllKeys(w.RequiredArgs)
	if ok {
		return nil
	}
	return NFError.ErrMissingArgument(w.Type, miss...)
}

type widgetHandler func(window fyne.Window, args NFData.NFInterface, widget Widget) (fyne.CanvasObject, error)

// Widgets is a map of all the widgets that are registered and can be used by the engine
var Widgets = map[string]widgetHandler{}

func (w Widget) Parse(window fyne.Window) (fyne.CanvasObject, error) {
	if handler, ok := Widgets[w.Type]; ok {
		if err := w.CheckArgs(); err != nil {
			return nil, err
		}
		return handler(window, w.Args, w)
	} else {
		return nil, NFError.ErrNotImplemented
	}
}

func (w Widget) GetInfo() (ID, Type string) {
	ID = w.ID
	Type = w.Type
	if ID == "" {
		ID = "Unknown"
	}
	return
}

// Register adds a custom widget to the customWidgets map
func (w Widget) Register(handler widgetHandler) {
	//Check if the name is already registered
	if _, ok := Widgets[w.Type]; ok {
		log.Printf("Widget %s is already registered, if you are using third party widgets yell at the developer of this type to use proper type naming", w.Type)
		return
	}
	Widgets[w.Type] = handler
	if ShouldExport {
		err := w.Export()
		if err != nil {
			log.Println(err)
		}
	}

}

var ShouldExport = false
var ExportPath = "export/widgets"

// Export is a function that is used to export the widget to a json file
// These files are used in the main editor to determine inputs needed to call a widget in a scene
func (w Widget) Export() error {
	//Make the json safe struct
	wBytes := struct {
		RequiredArgs map[string][]string `json:"RequiredArgs"`
		OptionalArgs map[string][]string `json:"OptionalArgs"`
	}{
		RequiredArgs: w.RequiredArgs.Export(),
		OptionalArgs: w.OptionalArgs.Export(),
	}

	//Export the function to a json file
	jsonBytes, err := json.MarshalIndent(wBytes, "", "  ")
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

	//Write the file to the export path
	err = os.WriteFile(ExportPath+"/"+w.Type+".json", jsonBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}
