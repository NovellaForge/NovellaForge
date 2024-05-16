package NFScene

import (
	"cmp"
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFFS"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFFunction"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFLayout"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// SceneMap is a map of scene names to their paths for easy access
var SceneMap = map[string]string{}

// Scene is the struct that holds all the information about a scene
type Scene struct {
	Name      string                 `json:"Name"`
	UUID      uuid.UUID              `json:"UUID"`
	Layout    *NFLayout.Layout       `json:"Layout"`
	Functions []*NFFunction.Function `json:"Functions"` // List of functions that are children of the scene for action based execution
	Args      *NFData.NFInterfaceMap `json:"Args"`
}

func (scene *Scene) FetchAll() (map[uuid.UUID][]NFObjects.NFObject, int) {
	all := make(map[uuid.UUID][]NFObjects.NFObject)
	return all, scene.FetchChildrenAndFunctions(all)
}

func (scene *Scene) FetchAllChildren() (map[uuid.UUID][]NFObjects.NFObject, int) {
	all := make(map[uuid.UUID][]NFObjects.NFObject)
	return all, scene.FetchChildren(all)
}

func (scene *Scene) FetchAllFunctions() (map[uuid.UUID][]NFObjects.NFObject, int) {
	all := make(map[uuid.UUID][]NFObjects.NFObject)
	return all, scene.FetchFunctions(all)
}

func (scene *Scene) MakeId() {
	log.Println("Remaking ID for Scene: ", scene.Name)
	scene.UUID = uuid.New()
	scene.Layout.MakeId()
	for _, function := range scene.Functions {
		function.MakeId()
	}
}

func (scene *Scene) CheckArgs() error {
	log.Println("Scene has no required arguments")
	return nil
}

func (scene *Scene) Validate() error {
	ids := scene.FetchIDs()
	//Check if the scene has an ID
	if scene.UUID == uuid.Nil {
		log.Println("A scene having no ID indicates that UUID generation has not occurred and needs to be re-called")
		return NFError.NewErrCriticalSceneValidation("Scene: " + scene.Name + " has no ID, UUID remake scheduled")
	}
	var valError error
	for _, function := range scene.Functions {
		valError = errors.Join(valError, function.Validate())
	}
	valError = errors.Join(valError, scene.Layout.Validate())
	if valError != nil {
		log.Println("Current Scene Validation Error: ", valError)
		if errors.Is(valError, NFError.ErrCriticalSceneValidation) {
			return NFError.NewErrCriticalSceneValidation("Scene: " + scene.Name + " has critical validation errors and should not be loaded")
		} else if errors.Is(valError, NFError.ErrSceneValidation) {
			return NFError.NewErrSceneValidation("Scene: " + scene.Name + " has validation errors")
		}
	}
	//Check if any ids appear more than once
	idMap := make(map[uuid.UUID]int)
	for _, id := range ids {
		if id == uuid.Nil {
			log.Println("A child object in Scene: ", scene.Name, " has no ID")
			return NFError.NewErrCriticalSceneValidation("Scene: " + scene.Name + " has a child with no ID, UUID remake scheduled")
		}
		idMap[id]++
	}
	for id, count := range idMap {
		if count > 1 {
			log.Println("ID: ", id, " appears ", count, " times")
			return NFError.NewErrCriticalSceneValidation("Scene: " + scene.Name + " has duplicate IDs, UUID remake scheduled")
		}
	}
	return nil
}

func (scene *Scene) GetID() uuid.UUID {
	return scene.UUID
}

func (scene *Scene) FetchIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0)
	ids = append(ids, scene.GetID())
	for _, function := range scene.Functions {
		ids = append(ids, function.GetID())
	}
	ids = append(ids, scene.Layout.FetchIDs()...)
	return ids
}

func (scene *Scene) GetName() string {
	return scene.Name
}

func (scene *Scene) SetName(newName string) {
	newName = strings.TrimSpace(newName)
	if newName != "" {
		scene.Name = newName
	}
}

func (scene *Scene) AddChild(newChild NFObjects.NFRendered) {
	log.Println("Scene cannot directly have children adding child to layout")
	scene.Layout.AddChild(newChild)
}

func (scene *Scene) GetChildByID(childID uuid.UUID) NFObjects.NFObject {
	return scene.Layout.GetChildByID(childID)
}

func (scene *Scene) GetFunctionByID(functionID uuid.UUID) NFObjects.NFObject {
	for _, function := range scene.Functions {
		if function.GetID() == functionID {
			return function
		}
	}
	return scene.Layout.GetFunctionByID(functionID)
}

func (scene *Scene) GetByID(ID uuid.UUID) NFObjects.NFObject {
	if scene.GetID() == ID {
		return scene
	}
	for _, function := range scene.Functions {
		if function.GetID() == ID {
			return function
		}
	}
	return scene.Layout.GetByID(ID)
}

func (scene *Scene) AddFunction(newFunction NFObjects.NFObject) {
	if function, ok := newFunction.(*NFFunction.Function); ok {
		scene.Functions = append(scene.Functions, function)
	} else {
		log.Println("Function is not of type *NFFunction.Function")
	}
}

func (scene *Scene) DeleteByID(ID uuid.UUID, search bool) error {
	if scene.GetID() == ID {
		return NFError.NewErrSceneValidation("Cannot delete scene object")
	}
	for i, function := range scene.Functions {
		if function.GetID() == ID {
			scene.Functions = append(scene.Functions[:i], scene.Functions[i+1:]...)
			log.Println("Function with ID: ", ID, " deleted")
			return nil
		}
	}
	if search {
		return scene.Layout.DeleteByID(ID, search)
	} else {
		return NFError.NewErrNotFound("Object with ID: " + ID.String() + " not found")
	}
}

func (scene *Scene) DeleteChild(childID uuid.UUID, search bool) error {
	if search {
		log.Println("Searching for child with ID: ", childID)
		err := scene.Layout.DeleteChild(childID, search)
		if err != nil {
			return err
		}
		log.Println("Child with ID: ", childID, " deleted")
	} else {
		log.Println("Scene cannot directly have children use scene.DeleteChild(childID, true) to search for children matching the id through the layout and its children")
	}
	return nil
}

func (scene *Scene) DeleteFunction(functionID uuid.UUID, search bool) error {
	//Check if the ID is in the functions
	for i, function := range scene.Functions {
		if function.GetID() == functionID {
			scene.Functions = append(scene.Functions[:i], scene.Functions[i+1:]...)
			log.Println("Function with ID: ", functionID, " deleted")
			return nil
		}
	}
	if search {
		return scene.Layout.DeleteFunction(functionID, search)
	} else {
		return NFError.NewErrNotFound("Function with ID: " + functionID.String() + " not found")
	}
}

func (scene *Scene) GetFunctions() []NFObjects.NFObject {
	functions := make([]NFObjects.NFObject, 0)
	for _, function := range scene.Functions {
		functions = append(functions, function)
	}
	return functions
}

func (scene *Scene) FetchChildrenAndFunctions(childrenAndFunctions map[uuid.UUID][]NFObjects.NFObject) int {
	if childrenAndFunctions == nil {
		childrenAndFunctions = make(map[uuid.UUID][]NFObjects.NFObject)
	}
	childrenAndFunctions[scene.GetID()] = append(childrenAndFunctions[scene.GetID()], scene.Layout)
	count := 1
	for _, function := range scene.Functions {
		childrenAndFunctions[scene.GetID()] = append(childrenAndFunctions[scene.GetID()], function)
		count++
	}
	//Add the layout as a child
	count += scene.Layout.FetchChildrenAndFunctions(childrenAndFunctions)
	return count
}

func (scene *Scene) FetchChildren(children map[uuid.UUID][]NFObjects.NFObject) int {
	if children == nil {
		children = make(map[uuid.UUID][]NFObjects.NFObject)
	}
	children[scene.GetID()] = append(children[scene.GetID()], scene.Layout)
	count := 1
	count += scene.Layout.FetchChildren(children)
	return count
}

func (scene *Scene) FetchFunctions(functions map[uuid.UUID][]NFObjects.NFObject) int {
	if functions == nil {
		functions = make(map[uuid.UUID][]NFObjects.NFObject)
	}
	functions[scene.GetID()] = append(functions[scene.GetID()], scene.Layout)
	count := 1
	for _, function := range scene.Functions {
		functions[scene.GetID()] = append(functions[scene.GetID()], function)
		count++
	}
	count += scene.Layout.FetchFunctions(functions)
	return count
}

func (scene *Scene) RunAllActions(action string, window fyne.Window, newValues *NFData.NFInterfaceMap) (int, []*NFData.NFInterfaceMap, error) {
	var fullErr error
	var returnValues []*NFData.NFInterfaceMap
	runnableFunctions := make([]*NFFunction.Function, 0)
	count := 0
	for _, function := range scene.Functions {
		if function.Action == action {
			runnableFunctions = append(runnableFunctions, function)
		}
	}

	if len(runnableFunctions) == 0 {
		return 0, returnValues, NFError.NewErrNotImplemented("Action: " + action + " for Scene: " + scene.GetName())
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

func (scene *Scene) RunAction(action string, window fyne.Window, newValues *NFData.NFInterfaceMap) (*NFData.NFInterfaceMap, error) {
	runnableFunctions := make([]*NFFunction.Function, 0)
	for _, function := range scene.Functions {
		if function.Action == action {
			runnableFunctions = append(runnableFunctions, function)
		}
	}

	if len(runnableFunctions) == 0 {
		return nil, NFError.NewErrNotImplemented("Action: " + action + " for Scene: " + scene.GetName())
	}

	//Sort the functions by their priority highest to lowest
	slices.SortStableFunc(runnableFunctions, func(i, j *NFFunction.Function) int {
		return cmp.Compare(i.Priority, j.Priority)
	})

	if len(runnableFunctions) > 0 {
		return runnableFunctions[0].Run(window, newValues)
	}
	return nil, NFError.NewErrNotImplemented("Action: " + action + " for Scene: " + scene.GetName())
}

func (scene *Scene) GetType() string {
	return scene.Name
}

func (scene *Scene) SetType(t string) {
	scene.Name = t
}

func (scene *Scene) GetArgs() *NFData.NFInterfaceMap {
	return scene.Args
}

func (scene *Scene) SetArgs(args *NFData.NFInterfaceMap) {
	scene.Args = args
}

//
//END of NFObject interface implementation
//

// NewScene creates a new scene with the given name, layout, and arguments
func NewScene(name string, layout *NFLayout.Layout, args *NFData.NFInterfaceMap) *Scene {
	return &Scene{
		Name:   name,
		Layout: layout,
		Args:   args,
	}
}

// ParseAndLoad runs Parse before loading the scene in to active scene data
func (scene *Scene) ParseAndLoad(window fyne.Window) (*SceneStack, error) {
	sceneObject, err := scene.Parse(window)
	if err != nil {
		return nil, err
	}
	NFData.ActiveSceneData = NFData.NewSceneData(scene.Name)
	NFData.ActiveSceneData.Layouts.Set("main", *scene.Layout)
	NFData.ActiveSceneData.Variables = scene.Args
	return sceneObject, nil
}

type SceneStack struct {
	widget.BaseWidget
	container *fyne.Container
}

type StackRenderer struct {
	stack *SceneStack
}

func (s *StackRenderer) Destroy() {}

func (s *StackRenderer) Layout(size fyne.Size) {
	s.stack.container.Layout.Layout(s.stack.container.Objects, size)
}

func (s *StackRenderer) MinSize() fyne.Size {
	return s.stack.container.MinSize()
}

func (s *StackRenderer) Objects() []fyne.CanvasObject {
	return s.stack.container.Objects
}

func (s *StackRenderer) Refresh() {
	s.stack.container.Refresh()
}

func (s *SceneStack) CreateRenderer() fyne.WidgetRenderer {
	return &StackRenderer{s}
}

func NewSceneStack(window fyne.Window, scene fyne.CanvasObject) *SceneStack {
	stack := &SceneStack{
		container: container.NewStack(scene),
	}
	stack.ExtendBaseWidget(stack)
	stack.RefreshOverlays(window)
	return stack
}

func (s *SceneStack) RefreshOverlays(window fyne.Window) {
	scene := s.container.Objects[0]
	s.container.Objects = nil
	s.Refresh()
	s.container.Objects = append(s.container.Objects, scene)
	for _, overlay := range Overlays {
		if overlay.visible {
			layout, err := overlay.layout.Parse(window)
			if err != nil {
				log.Println(err)
				dialog.ShowError(err, window)
			}
			s.container.Objects = append(s.container.Objects, layout)
		}
	}
	s.Refresh()
}

// Parse parses a scene and returns a fyne.CanvasObject that can be added to the window
func (scene *Scene) Parse(window fyne.Window) (*SceneStack, error) {
	layout, err := scene.Layout.Parse(window)
	if err != nil {
		return nil, err
	}
	return NewSceneStack(window, layout), err
}

func (scene *Scene) Save(path string) error {
	err := scene.Validate()
	if err != nil {
		if errors.Is(err, NFError.ErrCriticalSceneValidation) {
			return err
		}
	}

	//Check if the path is a file
	if filepath.Ext(path) != "" {
		path = filepath.Dir(path)
	}

	//Add the scene name and .NFScene to the path
	path = path + "/" + scene.Name + ".NFScene"

	path = filepath.Clean(path)
	if !fs.ValidPath(path) {
		log.Println("Invalid path")
		return NFError.NewErrSceneSave("Invalid path")
	}
	// Marshal the jsonScene
	jsonBytes, err := json.MarshalIndent(scene, "", "\t")
	if err != nil {
		return NFError.NewErrSceneSave("Error marshalling scene: " + scene.Name)
	}
	//Create the directories if they don't exist
	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return NFError.NewErrSceneSave("Error creating directories for scene: " + scene.Name)
	}
	// Write the file
	err = os.WriteFile(path, jsonBytes, 0755)
	if err != nil {
		return err
	}
	return nil
}

// Get gets a scene from the SceneMap loading it from the filesystem
func Get(name string, config ...NFFS.Configuration) (*Scene, error) {
	//If config[0] is not passed, use the default configuration
	if len(config) == 0 {
		config = append(config, NFFS.NewConfiguration(true))
	}
	if path, ok := SceneMap[name]; ok {
		file, err := NFFS.Open(path, config[0])
		if err != nil {
			return nil, err
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		scene := &Scene{}
		err = json.Unmarshal(data, scene)
		if err != nil {
			return nil, err
		}
		return scene, nil
	}
	return nil, errors.New("scene not registered")
}

func Load(path string) (*Scene, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	scene := &Scene{}
	err = json.Unmarshal(data, scene)
	if err != nil {
		return nil, err
	}

	err = scene.Validate()
	if err != nil {
		if errors.Is(err, NFError.ErrCriticalSceneValidation) {
			//Remake the ID and save the scene
			scene.MakeId()
		}
		if errors.Is(err, NFError.ErrSceneValidation) || errors.Is(err, NFError.ErrCriticalSceneValidation) {
			log.Println("Saving new validation changes to Scene and revalidating: ", scene.Name)
			err = scene.Save(path)
			if err != nil {
				if errors.Is(err, NFError.ErrCriticalSceneValidation) || errors.Is(err, NFError.ErrSceneSave) {
					return nil, err
				}
			}
		}
	}
	return scene, nil
}

// Register registers a scene with the SceneMap
func Register(name, path string) error {
	path = filepath.Clean(path)
	if !fs.ValidPath(path) {
		return errors.New("invalid path")
	}
	//Check if the name is already in the SceneMap
	if _, ok := SceneMap[name]; ok {
		return errors.New("scene already registered")
	}
	SceneMap[name] = path
	return nil
}

// RegisterAll registers all scenes by walking both the embedded and local filesystems
//
// This function is heavy and should only be called once at the start of the program,
// For adding scenes after the program has started use Register instead
//
// Like most NFFS functions this function only functions in the local directory and the embedded filesystems
func RegisterAll(path string) error {
	err := NFFS.Walk(path, NFFS.NewConfiguration(true), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		//Check if the file ends in .NFScene
		if strings.HasSuffix(path, ".NFScene") {
			name := strings.TrimSuffix(d.Name(), ".NFScene")
			if oldPath, ok := SceneMap[name]; ok {
				log.Println("Scene already registered: ", name, " at ", oldPath)
				log.Println("Make sure scenes have unique names for easy management")
			} else {
				SceneMap[name] = path
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
