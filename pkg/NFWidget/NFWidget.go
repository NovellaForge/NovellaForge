package NFWidget

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
	"log"
	"os"
	"path/filepath"
)

// Widget is the struct that holds all the information about a widget
type Widget struct {
	// ID is the unique ID of the widget for later reference in editing
	ID string `json:"ID"`
	// Type is the type of widget that is used to parse the widget this should be Globally Unique, so when making
	//custom ones prefix it with your package name like "MyPackage.MyWidget"
	Type string `json:"Type"`
	// Children is a list of widgets that are children of this widget
	Children []*Widget `json:"Children"`
	// RequiredArgs is a list of arguments that are required for the widget to run
	RequiredArgs *NFData.NFInterfaceMap `json:"-"`
	// OptionalArgs is a list of arguments that are optional for the widget to run
	OptionalArgs *NFData.NFInterfaceMap `json:"-"`
	// Args is a list of arguments that are passed to the widget through the scene
	Args *NFData.NFInterfaceMap `json:"Args"`
}

func (w *Widget) GetType() string {
	return w.Type
}

func (w *Widget) SetType(t string) {
	w.Type = t
}

func (w *Widget) GetArgs() *NFData.NFInterfaceMap {
	return w.Args
}

func (w *Widget) SetArgs(args *NFData.NFInterfaceMap) {
	w.Args = args
}

// NewWithID creates a new widget with the given ID and type
func NewWithID(id, widgetType string, children []*Widget, args *NFData.NFInterfaceMap) *Widget {
	return &Widget{
		ID:       id,
		Type:     widgetType,
		Children: children,
		Args:     args,
	}
}

func NewChildren(children ...*Widget) []*Widget {
	return children
}

// New creates a new widget with the given type
func New(widgetType string, children []*Widget, args *NFData.NFInterfaceMap) *Widget {
	return &Widget{
		Type:     widgetType,
		Children: children,
		Args:     args,
	}
}

// NewAndRegister creates a new widget with the given type and registers it
func NewAndRegister(widgetType string, requiredArgs, optionalArgs *NFData.NFInterfaceMap, handler widgetHandler) {
	widget := &Widget{
		Type:         widgetType,
		RequiredArgs: requiredArgs,
		OptionalArgs: optionalArgs,
	}
	widget.Register(handler)
}

func (w *Widget) CheckArgs() error {
	info, err := GetWidgetInfo(w.Type)
	if err != nil {
		return err
	}
	w.RequiredArgs = info.RequiredArgs
	ok, miss := w.Args.HasAllKeys(w.RequiredArgs)
	if ok {
		return nil
	}
	var missingErr error
	for _, m := range miss {
		errors.Join(missingErr, NFError.NewErrMissingArgument(w.Type, m))
	}
	return missingErr
}

type widgetHandler func(window fyne.Window, args *NFData.NFInterfaceMap, widget *Widget) (fyne.CanvasObject, error)

// Widgets is a map of all the widgets that are registered and can be used by the engine
var Widgets = map[string]widgetWithHandler{}

type widgetWithHandler struct {
	Info    Widget
	Handler widgetHandler
}

// NewWithHandler creates a new widget with the given info and handler
func newWithHandler(info Widget, handler widgetHandler) widgetWithHandler {
	return widgetWithHandler{
		Info:    info,
		Handler: handler,
	}
}

// GetWidgetInfo gets the info of a widget
func GetWidgetInfo(widget string) (Widget, error) {
	if w, ok := Widgets[widget]; ok {
		return w.Info, nil
	} else {
		return Widget{}, NFError.NewErrNotImplemented(widget)
	}
}

func (w *Widget) Parse(window fyne.Window) (fyne.CanvasObject, error) {
	if ref, ok := Widgets[w.Type]; ok {
		if err := w.CheckArgs(); err != nil {
			return nil, err
		}
		return ref.Handler(window, w.Args, w)
	} else {
		return nil, NFError.NewErrNotImplemented(w.Type + ":" + w.ID)
	}
}

func (w *Widget) GetInfo() (ID, Type string) {
	ID = w.ID
	Type = w.Type
	if ID == "" {
		ID = "Unknown"
	}
	return
}

// Register adds a custom widget to the customWidgets map
func (w *Widget) Register(handler widgetHandler) {
	//Check if the name is already registered
	if _, ok := Widgets[w.Type]; ok {
		log.Printf("Widget %s is already registered, if you are using third party widgets yell at the developer of this type to use proper type naming", w.Type)
		return
	}
	Widgets[w.Type] = newWithHandler(*w, handler)
}

// ExportPath is the path where the widget will be exported
var ExportPath = "exports/widgets"

// ExportRegistered exports all registered widgets to json files
func ExportRegistered() {
	for _, widget := range Widgets {
		err := widget.Info.Export()
		if err != nil {
			log.Println(err)
		}
	}
}

// Export is a function that is used to export the widget to a json file
// These files are used in the main editor to determine inputs needed to call a widget in a scene
func (w *Widget) Export() error {
	//Check if the export path has a trailing slash and add one if it doesn't
	if ExportPath[len(ExportPath)-1] != '/' {
		ExportPath += "/"
	}
	//Make the json safe struct
	wBytes := struct {
		Type         string              `json:"Type"`
		RequiredArgs map[string][]string `json:"RequiredArgs"`
		OptionalArgs map[string][]string `json:"OptionalArgs"`
	}{
		Type:         w.Type,
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
	err = os.WriteFile(ExportPath+w.Type+".NFWidget", jsonBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func Load(path string) (a NFData.AssetProperties, err error) {
	if filepath.Ext(path) != ".NFWidget" {
		return NFData.AssetProperties{}, errors.New("invalid file type")
	}
	err = a.Load(path)
	return a, err
}
