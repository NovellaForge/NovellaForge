package NFWidget

import (
	"cmp"
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"github.com/google/uuid"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFFunction"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Widget is the struct that holds all the information about a widget
type Widget struct {
	// Name is the name of the widget for display in the editor
	Name string `json:"Name"`
	// ID is the unique ID of the widget for later reference in editing
	UUID uuid.UUID `json:"UUID"`
	// Type is the type of widget that is used to parse the widget this should be Globally Unique, so when making
	//custom ones prefix it with your package name like "MyPackage.MyWidget"
	Type string `json:"Type"`
	// Children is a list of widgets that are children of this widget
	Children []*Widget `json:"Children"`
	// Functions is a list of functions that are children of this widget for action based execution
	Functions []*NFFunction.Function `json:"Functions"`
	// SupportedActions is a list of actions that the widget supports for action based execution
	SupportedActions []string `json:"-"`
	// RequiredArgs is a list of arguments that are required for the widget to run
	RequiredArgs *NFData.NFInterfaceMap `json:"-"`
	// OptionalArgs is a list of arguments that are optional for the widget to run
	OptionalArgs *NFData.NFInterfaceMap `json:"-"`
	// Args is a list of arguments that are passed to the widget through the scene
	Args *NFData.NFInterfaceMap `json:"Args"`
}

func (w *Widget) Validate() error {
	name := w.GetName()
	w.SetName(w.GetName())
	newName := w.GetName()
	var fullErr error
	if name != newName {
		fullErr = errors.Join(fullErr, NFError.NewErrSceneValidation("Name: "+name+" is invalid, changing to: "+newName))
	}
	for _, child := range w.Children {
		if err := child.Validate(); err != nil {
			fullErr = errors.Join(fullErr, err)
		}
	}
	for _, function := range w.Functions {
		if err := function.Validate(); err != nil {
			fullErr = errors.Join(fullErr, err)
		}
	}
	if err := w.CheckArgs(); err != nil {
		fullErr = errors.Join(fullErr, err)
	}
	return fullErr
}

func (w *Widget) MakeId() {
	w.UUID = uuid.New()
	for _, child := range w.Children {
		child.MakeId()
	}
	for _, function := range w.Functions {
		function.MakeId()
	}
}

func (w *Widget) GetID() uuid.UUID {
	return w.UUID
}

func (w *Widget) FetchIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0)
	ids = append(ids, w.UUID)
	for _, child := range w.Children {
		ids = append(ids, child.FetchIDs()...)
	}
	for _, function := range w.Functions {
		ids = append(ids, function.GetID())
	}
	return ids
}

func (w *Widget) GetName() string {
	return w.Name
}

func (w *Widget) SetName(newName string) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		newName = w.GetType()
	}
	w.Name = newName
}

func (w *Widget) AddChild(new NFObjects.NFRendered) {
	if newWidget, ok := new.(*Widget); ok {
		w.Children = append(w.Children, newWidget)
	} else {
		log.Println("Cannot add a non widget object as a child to a widget")
	}
}

func (w *Widget) AddFunction(new NFObjects.NFObject) {
	if newFunction, ok := new.(*NFFunction.Function); ok {
		w.Functions = append(w.Functions, newFunction)
	} else {
		log.Println("Cannot add a non function object as a function to a widget")
	}
}

func (w *Widget) DeleteByID(ID uuid.UUID, search bool) error {
	for i, child := range w.Children {
		if child.GetID() == ID {
			w.Children = append(w.Children[:i], w.Children[i+1:]...)
			return nil
		}

	}
	for i, function := range w.Functions {
		if function.GetID() == ID {
			w.Functions = append(w.Functions[:i], w.Functions[i+1:]...)
			return nil
		}
	}
	if search {
		for _, child := range w.Children {
			if err := child.DeleteByID(ID, true); err == nil {
				return nil
			}
		}
	}
	return NFError.NewErrNotFound("Could not find ID: " + ID.String())
}

func (w *Widget) DeleteChild(childID uuid.UUID, search bool) error {
	for i, child := range w.Children {
		if child.GetID() == childID {
			w.Children = append(w.Children[:i], w.Children[i+1:]...)
			return nil
		}
	}
	if search {
		for _, child := range w.Children {
			if err := child.DeleteChild(childID, true); err == nil {
				return nil
			}
		}
	}
	return NFError.NewErrNotFound("Could not find ID: " + childID.String())
}

func (w *Widget) DeleteFunction(functionID uuid.UUID, search bool) error {
	for i, function := range w.Functions {
		if function.GetID() == functionID {
			w.Functions = append(w.Functions[:i], w.Functions[i+1:]...)
			return nil
		}
	}
	if search {
		for _, child := range w.Children {
			if err := child.DeleteFunction(functionID, true); err == nil {
				return nil
			}
		}
	}
	return NFError.NewErrNotFound("Could not find ID: " + functionID.String())
}

func (w *Widget) GetFunctions() []NFObjects.NFObject {
	functions := make([]NFObjects.NFObject, 0)
	for _, function := range w.Functions {
		functions = append(functions, function)
	}
	return functions
}

func (w *Widget) GetChildByID(childID uuid.UUID) NFObjects.NFObject {
	for _, child := range w.Children {
		if child.GetID() == childID {
			return child
		}
	}
	for _, child := range w.Children {
		if c := child.GetChildByID(childID); c != nil {
			return c
		}
	}
	return nil
}

func (w *Widget) GetFunctionByID(functionID uuid.UUID) NFObjects.NFObject {
	for _, function := range w.Functions {
		if function.GetID() == functionID {
			return function
		}
	}
	for _, child := range w.Children {
		if c := child.GetFunctionByID(functionID); c != nil {
			return c
		}
	}
	return nil
}

func (w *Widget) GetByID(ID uuid.UUID) NFObjects.NFObject {
	if w.GetID() == ID {
		return w
	}
	for _, child := range w.Children {
		if child.GetID() == ID {
			return child
		}
	}
	for _, function := range w.Functions {
		if function.GetID() == ID {
			return function
		}
	}
	for _, child := range w.Children {
		if c := child.GetByID(ID); c != nil {
			return c
		}
	}
	return nil
}

func (w *Widget) FetchChildrenAndFunctions(childrenAndFunctions map[uuid.UUID][]NFObjects.NFObject) int {
	if childrenAndFunctions == nil {
		childrenAndFunctions = make(map[uuid.UUID][]NFObjects.NFObject)
	}
	count := 0
	for _, function := range w.Functions {
		childrenAndFunctions[w.UUID] = append(childrenAndFunctions[w.UUID], function)
		count++
	}
	for _, child := range w.Children {
		childrenAndFunctions[w.UUID] = append(childrenAndFunctions[w.UUID], child)
		count++
		count += child.FetchChildrenAndFunctions(childrenAndFunctions)
	}

	return count
}

func (w *Widget) FetchChildren(children map[uuid.UUID][]NFObjects.NFObject) int {
	if children == nil {
		children = make(map[uuid.UUID][]NFObjects.NFObject)
	}
	count := 0
	for _, child := range w.Children {
		children[w.UUID] = append(children[w.UUID], child)
		count++
		count += child.FetchChildren(children)
	}
	return count
}

func (w *Widget) FetchFunctions(functions map[uuid.UUID][]NFObjects.NFObject) int {
	if functions == nil {
		functions = make(map[uuid.UUID][]NFObjects.NFObject)
	}
	count := 0
	for _, function := range w.Functions {
		functions[w.UUID] = append(functions[w.UUID], function)
		count++
	}
	for _, child := range w.Children {
		count += child.FetchFunctions(functions)
	}
	return count
}

func (w *Widget) RunAllActions(action string, window fyne.Window, newValues *NFData.NFInterfaceMap) (int, []*NFData.NFInterfaceMap, error) {
	var fullErr error
	var returnValues []*NFData.NFInterfaceMap
	runnableFunctions := make([]*NFFunction.Function, 0)
	count := 0
	for _, function := range w.Functions {
		if function.Action == action {
			runnableFunctions = append(runnableFunctions, function)
		}
	}

	if len(runnableFunctions) == 0 {
		return 0, returnValues, NFError.NewErrNotImplemented("Action: " + action + " for Widget: " + w.GetName() + ":" + w.GetID().String())
	}

	//Sort the functions by their priority highest to lowest
	slices.SortStableFunc(runnableFunctions, func(i, j *NFFunction.Function) int {
		return cmp.Compare(i.Priority, j.Priority)
	})

	for _, function := range runnableFunctions {
		newReturn, err := function.Run(window, newValues)
		if err != nil {
			fullErr = errors.Join(fullErr, err)
		}
		returnValues = append(returnValues, newReturn)
		count++
	}

	return count, returnValues, fullErr
}

func (w *Widget) RunAction(action string, window fyne.Window, newValues *NFData.NFInterfaceMap) (*NFData.NFInterfaceMap, error) {
	runnableFunctions := make([]*NFFunction.Function, 0)
	for _, function := range w.Functions {
		if function.Action == action {
			runnableFunctions = append(runnableFunctions, function)
		}
	}

	if len(runnableFunctions) == 0 {
		return nil, NFError.NewErrNotImplemented("Action: " + action + " for Widget: " + w.GetName() + ":" + w.GetID().String())
	}

	//Sort the functions by their priority highest to lowest
	slices.SortStableFunc(runnableFunctions, func(i, j *NFFunction.Function) int {
		return cmp.Compare(i.Priority, j.Priority)
	})

	if len(runnableFunctions) > 0 {
		return runnableFunctions[0].Run(window, newValues)
	}
	return nil, NFError.NewErrNotImplemented("Action: " + action + " for Widget: " + w.GetName() + ":" + w.GetID().String())
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

func NewChildren(children ...*Widget) []*Widget {
	return children
}

// New creates a new widget with the given type
func New(widgetType string, children []*Widget, args *NFData.NFInterfaceMap) *Widget {
	return &Widget{
		Name:     widgetType,
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
		return nil, NFError.NewErrNotImplemented(w.Type + ":" + w.GetID().String())
	}
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
