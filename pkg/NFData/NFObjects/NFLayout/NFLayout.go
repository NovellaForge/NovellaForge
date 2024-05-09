package NFLayout

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFWidget"
	"log"
	"os"
	"path/filepath"
)

// layoutHandler is a function type that handles the layout logic.
type layoutHandler func(window fyne.Window, args *NFData.NFInterfaceMap, l *Layout) (fyne.CanvasObject, error)

// Layout struct holds all the information about a layout.
//
// Layouts are used to define how widgets are placed in a scene.
//
// They contain a list of widgets that are children of the layout.
//
// They also contain a list of arguments that are required for the layout to run.
type Layout struct {
	Type         string                 `json:"Type"`    // Type of the layout
	Children     []*NFWidget.Widget     `json:"Widgets"` // List of widgets that are children of the layout
	RequiredArgs *NFData.NFInterfaceMap `json:"-"`       // List of arguments that are required for the layout to run
	OptionalArgs *NFData.NFInterfaceMap `json:"-"`       // List of arguments that are optional for the layout to run
	Args         *NFData.NFInterfaceMap `json:"Args"`    // List of arguments that are passed to the layout
}

func (l *Layout) AddChild(object NFObjects.NFObject) {
	//Try to convert the object to a widget
	newWidget, ok := object.(*NFWidget.Widget)
	if !ok {
		log.Println("Cannot add non-widget object to widget")
		return
	}
	l.Children = append(l.Children, newWidget)
}

func (l *Layout) GetType() string {
	return l.Type
}

func (l *Layout) SetType(t string) {
	l.Type = t
}

func (l *Layout) GetArgs() *NFData.NFInterfaceMap {
	return l.Args
}

func (l *Layout) SetArgs(args *NFData.NFInterfaceMap) {
	l.Args = args
}

// NewLayout creates a new layout with the given type.
func NewLayout(layoutType string, children []*NFWidget.Widget, args *NFData.NFInterfaceMap) *Layout {
	return &Layout{
		Type:     layoutType,
		Children: children,
		Args:     args,
	}
}

// NewChildren creates a new list of children widgets.
func NewChildren(children ...*NFWidget.Widget) []*NFWidget.Widget {
	return children
}

// NewLayoutRegister creates a new layout with the given type and registers it.
func NewLayoutRegister(layoutType string, requiredArgs, optionalArgs *NFData.NFInterfaceMap, handler layoutHandler) {
	layout := &Layout{
		Type:         layoutType,
		RequiredArgs: requiredArgs,
		OptionalArgs: optionalArgs,
	}
	layout.Register(handler)
}

// CheckArgs checks if all required arguments are present in Args.
// Returns an error if a required argument is missing.
func (l *Layout) CheckArgs() error {
	info, err := GetLayoutInfo(l.Type)
	if err != nil {
		return err
	}
	l.RequiredArgs = info.RequiredArgs
	ok, miss := l.Args.HasAllKeys(l.RequiredArgs)
	if ok {
		return nil
	}
	var missingErr error
	for _, m := range miss {
		errors.Join(missingErr, NFError.NewErrMissingArgument(l.Type, m))
	}
	return missingErr
}

// layouts is a map that holds all the registered layout handlers.
var layouts = map[string]layoutWithHandler{}

type layoutWithHandler struct {
	Info    Layout
	Handler layoutHandler
}

// GetLayoutInfo returns the layout info for the given layout type
func GetLayoutInfo(layout string) (Layout, error) {
	if l, ok := layouts[layout]; ok {
		return l.Info, nil
	} else {
		return Layout{}, NFError.NewErrNotImplemented("Layout Type: " + layout + " is not implemented")
	}
}

// newWithHandler creates a new layout with the given info and handler.
func newWithHandler(info Layout, handler layoutHandler) layoutWithHandler {
	return layoutWithHandler{
		Info:    info,
		Handler: handler,
	}
}

// Parse checks if the layout type exists in the registered layouts.
// If it does, it uses the assigned handler to parse the layout.
// If it doesn't, it returns an error.
func (l *Layout) Parse(window fyne.Window) (fyne.CanvasObject, error) {
	if ref, ok := layouts[l.Type]; ok {
		if err := l.CheckArgs(); err != nil {
			return nil, err
		}
		return ref.Handler(window, l.Args, l)
	} else {
		return nil, errors.New("layout unable to be parsed")
	}
}

// Register registers a custom layout handler.
// If the name is already registered, it will not be registered again.
func (l *Layout) Register(handler layoutHandler) {
	if _, ok := layouts[l.Type]; ok {
		log.Printf("Layout %s already registered\nIf you are using third party layouts, please yell at the developer to make sure they have properly namespaced their layouts", l.Type)
		return
	}
	layouts[l.Type] = newWithHandler(*l, handler)
}

// ExportPath is the path where the layout will be exported.
var ExportPath = "exports/layouts"

// ExportRegistered exports all registered layouts to json files.
func ExportRegistered() {
	for _, layout := range layouts {
		err := layout.Info.Export()
		if err != nil {
			log.Println(err)
		}
	}
}

// Export exports the layout to a json file.
// These files are used in the main editor to determine inputs needed to create the layout in a scene.
func (l *Layout) Export() error {
	//Check if the export path has a trailing slash and add one if it doesn't
	if ExportPath[len(ExportPath)-1] != '/' {
		ExportPath += "/"
	}

	lBytes := NFData.AssetProperties{
		Type:         l.Type,
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

	err = os.WriteFile(ExportPath+l.Type+".NFLayout", jsonBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func Load(path string) (a NFData.AssetProperties, err error) {
	if filepath.Ext(path) != ".NFLayout" {
		return NFData.AssetProperties{}, errors.New("invalid file type")
	}
	err = a.Load(path)
	return a, err
}
