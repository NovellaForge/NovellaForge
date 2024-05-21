package NFLayout

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
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFWidget"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
	Name             string                 `json:"Name"`      // Name of the layout
	Type             string                 `json:"Type"`      // Type of the layout
	UUID             uuid.UUID              `json:"UUID"`      // uuid.UUID of the layout
	Children         []*NFWidget.Widget     `json:"Widgets"`   // List of widgets that are children of the layout
	Functions        []*NFFunction.Function `json:"Functions"` // List of functions that are children of the layout for action based execution
	SupportedActions []string               `json:"-"`
	RequiredArgs     *NFData.NFInterfaceMap `json:"-"`    // List of arguments that are required for the layout to run
	OptionalArgs     *NFData.NFInterfaceMap `json:"-"`    // List of arguments that are optional for the layout to run
	Args             *NFData.NFInterfaceMap `json:"Args"` // List of arguments that are passed to the layout
}

func (l *Layout) Validate() error {
	name := l.GetName()
	l.SetName(l.GetName())
	newName := l.GetName()
	var fullErr error
	if name != newName {
		fullErr = errors.Join(fullErr, NFError.NewErrSceneValidation("Name: "+name+" is invalid, setting to: "+newName))
	}
	for _, child := range l.Children {
		err := child.Validate()
		if err != nil {
			fullErr = errors.Join(fullErr, err)
		}
	}
	for _, function := range l.Functions {
		err := function.Validate()
		if err != nil {
			fullErr = errors.Join(fullErr, err)
		}
	}
	return fullErr
}

func (l *Layout) MakeId() {
	l.UUID = uuid.New()
	for _, child := range l.Children {
		child.MakeId()
	}
	for _, function := range l.Functions {
		function.MakeId()
	}
}

func (l *Layout) FetchIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0)
	ids = append(ids, l.UUID)
	for _, child := range l.Children {
		ids = append(ids, child.FetchIDs()...)
	}
	for _, function := range l.Functions {
		ids = append(ids, function.GetID())
	}
	return ids
}

func (l *Layout) GetName() string {
	return l.Name
}

func (l *Layout) SetName(newName string) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		newName = "Layout"
	}
	l.Name = newName
}

func (l *Layout) AddChild(new NFObjects.NFRendered) {
	if newWidget, ok := new.(*NFWidget.Widget); ok {
		l.Children = append(l.Children, newWidget)
	} else {
		log.Println("Cannot add non widget object to layout as child")
	}
}

func (l *Layout) AddFunction(new NFObjects.NFObject) {
	if newFunction, ok := new.(*NFFunction.Function); ok {
		l.Functions = append(l.Functions, newFunction)
	} else {
		log.Println("Cannot add non function object to layout as function")
	}
}

func (l *Layout) DeleteByID(ID uuid.UUID, search bool) error {
	for i, child := range l.Children {
		if child.GetID() == ID {
			l.Children = append(l.Children[:i], l.Children[i+1:]...)
			log.Println("Deleted child with ID: " + ID.String() + " from layout: " + l.GetID().String())
			return nil
		}
	}
	for i, function := range l.Functions {
		if function.GetID() == ID {
			l.Functions = append(l.Functions[:i], l.Functions[i+1:]...)
			log.Println("Deleted function with ID: " + ID.String() + " from layout: " + l.GetID().String())
			return nil
		}
	}
	if search {
		for _, child := range l.Children {
			err := child.DeleteByID(ID, search)
			if err == nil {
				return nil
			}
		}
	}
	return NFError.NewErrNotFound("Object with ID: " + ID.String() + " not found in layout: " + l.GetID().String())
}

func (l *Layout) DeleteChild(childID uuid.UUID, search bool) error {
	for i, child := range l.Children {
		if child.GetID() == childID {
			l.Children = append(l.Children[:i], l.Children[i+1:]...)
			log.Println("Deleted child with ID: " + childID.String() + " from layout: " + l.GetID().String())
			return nil
		}
	}
	if search {
		for _, child := range l.Children {
			err := child.DeleteChild(childID, search)
			if err == nil {
				return nil
			}
		}

	}
	return NFError.NewErrNotFound("Child with ID: " + childID.String() + " not found in layout: " + l.GetID().String())
}

func (l *Layout) DeleteFunction(functionID uuid.UUID, search bool) error {
	for i, function := range l.Functions {
		if function.GetID() == functionID {
			l.Functions = append(l.Functions[:i], l.Functions[i+1:]...)
			log.Println("Deleted function with ID: " + functionID.String() + " from layout: " + l.GetID().String())
			return nil
		}
	}
	if search {
		for _, child := range l.Children {
			err := child.DeleteFunction(functionID, search)
			if err == nil {
				return nil
			}
		}
	}
	return NFError.NewErrNotFound("Function with ID: " + functionID.String() + " not found in layout: " + l.GetID().String())
}

func (l *Layout) GetFunctions() []NFObjects.NFObject {
	functions := make([]NFObjects.NFObject, 0)
	for _, function := range l.Functions {
		functions = append(functions, function)
	}
	return functions
}

func (l *Layout) GetChildByID(childID uuid.UUID) NFObjects.NFObject {
	for _, child := range l.Children {
		if child.GetID() == childID {
			return child
		}
	}
	//Check the children
	for _, child := range l.Children {
		if found := child.GetChildByID(childID); found != nil {
			return found
		}
	}
	return nil
}

func (l *Layout) GetFunctionByID(functionID uuid.UUID) NFObjects.NFObject {
	for _, function := range l.Functions {
		if function.GetID() == functionID {
			return function
		}
	}
	//Check the children
	for _, child := range l.Children {
		if found := child.GetFunctionByID(functionID); found != nil {
			return found
		}
	}
	return nil
}

func (l *Layout) GetByID(ID uuid.UUID) NFObjects.NFObject {
	if l.GetID() == ID {
		return l
	}
	for _, child := range l.Children {
		if found := child.GetByID(ID); found != nil {
			return found
		}
	}
	for _, function := range l.Functions {
		if function.GetID() == ID {
			return function
		}
	}
	return nil
}

func (l *Layout) FetchChildrenAndFunctions(childrenAndFunctions map[uuid.UUID][]NFObjects.NFObject) int {
	if childrenAndFunctions == nil {
		childrenAndFunctions = make(map[uuid.UUID][]NFObjects.NFObject)
	}
	count := 0
	for _, child := range l.Children {
		childrenAndFunctions[l.GetID()] = append(childrenAndFunctions[l.GetID()], child)
		count++
		count += child.FetchChildrenAndFunctions(childrenAndFunctions)
	}
	for _, function := range l.Functions {
		childrenAndFunctions[l.GetID()] = append(childrenAndFunctions[l.GetID()], function)
		count++
	}
	return count
}

func (l *Layout) FetchChildren(children map[uuid.UUID][]NFObjects.NFObject) int {
	if children == nil {
		children = make(map[uuid.UUID][]NFObjects.NFObject)
	}
	count := 0
	for _, child := range l.Children {
		children[l.GetID()] = append(children[l.GetID()], child)
		count++
		count += child.FetchChildren(children)
	}
	return count
}

func (l *Layout) FetchFunctions(functions map[uuid.UUID][]NFObjects.NFObject) int {
	if functions == nil {
		functions = make(map[uuid.UUID][]NFObjects.NFObject)
	}
	count := 0
	for _, function := range l.Functions {
		functions[l.GetID()] = append(functions[l.GetID()], function)
		count++
	}
	for _, child := range l.Children {
		count += child.FetchFunctions(functions)
	}
	return count
}

func (l *Layout) GetID() uuid.UUID {
	return l.UUID
}

func (l *Layout) RunAllActions(action string, window fyne.Window, newValues *NFData.NFInterfaceMap) (int, []*NFData.NFInterfaceMap, error) {
	var fullErr error
	var returnValues []*NFData.NFInterfaceMap
	runnableFunctions := make([]*NFFunction.Function, 0)
	count := 0
	for _, function := range l.Functions {
		if function.Action == action {
			runnableFunctions = append(runnableFunctions, function)
		}
	}

	if len(runnableFunctions) == 0 {
		return 0, returnValues, NFError.NewErrNotImplemented("Action: " + action + " for layout: " + l.GetName() + " with ID: " + l.GetID().String())
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

func (l *Layout) RunAction(action string, window fyne.Window, newValues *NFData.NFInterfaceMap) (*NFData.NFInterfaceMap, error) {
	runnableFunctions := make([]*NFFunction.Function, 0)
	for _, function := range l.Functions {
		if function.Action == action {
			runnableFunctions = append(runnableFunctions, function)
		}
	}

	if len(runnableFunctions) == 0 {
		return nil, NFError.NewErrNotImplemented("Action: " + action + " for Layout: " + l.GetName() + " with ID: " + l.GetID().String())
	}

	//Sort the functions by their priority highest to lowest
	slices.SortStableFunc(runnableFunctions, func(i, j *NFFunction.Function) int {
		return cmp.Compare(i.Priority, j.Priority)
	})

	if len(runnableFunctions) > 0 {
		return runnableFunctions[0].Run(window, newValues)
	}
	return nil, NFError.NewErrNotImplemented("Action: " + action + " for Layout: " + l.GetName() + " with ID: " + l.GetID().String())
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

// New creates a new layout with the given type and generates a new UUID.
func New(layoutType string, children []*NFWidget.Widget, args *NFData.NFInterfaceMap) *Layout {
	return &Layout{
		Type:     layoutType,
		Children: children,
		Args:     args,
		UUID:     uuid.New(),
	}
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

	lBytes := NFObjects.AssetProperties{
		Type:             l.Type,
		RequiredArgs:     l.RequiredArgs.Export(),
		SupportedActions: l.SupportedActions,
		OptionalArgs:     l.OptionalArgs.Export(),
	}

	jsonBytes, err := json.MarshalIndent(lBytes, "", "  ")
	if err != nil {
		return err
	}

	_, err = os.Stat(ExportPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(ExportPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(ExportPath+l.Type+".NFLayout", jsonBytes, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func Load(path string) (a NFObjects.AssetProperties, err error) {
	if filepath.Ext(path) != ".NFLayout" {
		return NFObjects.AssetProperties{}, errors.New("invalid file type")
	}
	err = a.Load(path)
	return a, err
}
