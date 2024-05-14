package NFScene

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFFS"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFLayout"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// SceneMap is a map of scene names to their paths for easy access
var SceneMap = map[string]string{}

// Scene is the struct that holds all the information about a scene
// TODO Add startup values to the scene to be populated when the scene is loaded
type Scene struct {
	Name   string                 `json:"Name"`
	Layout *NFLayout.Layout       `json:"Layout"`
	Args   *NFData.NFInterfaceMap `json:"Args"`
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
	//Check if the path is a file
	if filepath.Ext(path) != "" {
		path = filepath.Dir(path)
	}

	//Add the scene name and .NFScene to the path
	path = path + "/" + scene.Name + ".NFScene"

	path = filepath.Clean(path)
	if !fs.ValidPath(path) {
		log.Println("Invalid path")
		return errors.New("invalid path")
	}
	// Marshal the jsonScene
	jsonBytes, err := json.MarshalIndent(scene, "", "\t")
	if err != nil {
		return err
	}
	//Create the directories if they don't exist
	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
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

func (scene *Scene) Validate() (bool, error) {
	if scene.Name == "" {
		return false, NFError.NewErrSceneValidation("scene name is empty")
	}
	if scene.Layout == nil {
		//Create a new simple layout if the layout is nil
		scene.Layout = NFLayout.NewLayout("VBox", nil, nil)
	}
	return scene.Layout.Validate()
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

	changed, err := scene.Validate()
	if changed {
		if errors.Is(err, NFError.ErrSceneValidation) {
			return nil, err
		}
		//Save the scene if it was changed
		err = scene.Save(path)
		if err != nil {
			return nil, err
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
