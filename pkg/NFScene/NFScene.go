package NFScene

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFFS"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SceneMap is a map of scene names to their paths for easy access
var SceneMap = map[string]string{}

// Get gets a scene from the SceneMap loading it from the filesystem
func Get(name string) (*Scene, error) {
	if path, ok := SceneMap[name]; ok {
		file, err := NFFS.Open(path)
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

	//Loop through the children of the layout and set the ID to SceneName.TypeName#Number if it doesn't have one
	counts := map[string]int{}
	for i, child := range scene.Layout.Children {
		if child.ID == "" {
			if count, ok := counts[child.Type]; ok {
				child.ID = scene.Name + "." + child.Type + "#" + strconv.Itoa(count+1)
				scene.Layout.Children[i] = child
				counts[child.Type] = count + 1
			} else {
				counts[child.Type] = 1
				child.ID = scene.Name + "." + child.Type + "#1"
				scene.Layout.Children[i] = child
			}
		}
	}
	//Save the scene back to the file in case the ID's were changed
	err = scene.Save(path)
	if err != nil {
		log.Println("Error saving scene: ", err)
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
	err := NFFS.Walk(path, func(path string, d fs.DirEntry, err error) error {
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

// Scene is the struct that holds all the information about a scene
// TODO Add startup values to the scene to be populated when the scene is loaded
type Scene struct {
	Name   string                 `json:"Name"`
	Layout *NFLayout.Layout       `json:"Layout"`
	Args   *NFData.NFInterfaceMap `json:"Args"`
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
func (scene *Scene) ParseAndLoad(window fyne.Window, overlay ...NFLayout.Layout) (fyne.CanvasObject, error) {
	NFData.ActiveSceneData = NFData.NewSceneData(scene.Name)
	NFData.ActiveSceneData.Layouts.Set("main", *scene.Layout)
	NFData.ActiveSceneData.Variables = scene.Args
	for i, o := range overlay {
		NFData.ActiveSceneData.Layouts.Set("overlay_"+strconv.Itoa(i), o)
	}
	return scene.Parse(window, overlay...)
}

// Parse parses a scene and returns a fyne.CanvasObject that can be added to the window each argument passed should be an overlay, with the first being the bottom most overlay
func (scene *Scene) Parse(window fyne.Window, overlay ...NFLayout.Layout) (fyne.CanvasObject, error) {
	stack := container.NewStack()
	layout, err := scene.Layout.Parse(window)
	if err != nil {
		return nil, err
	}
	stack.Add(layout)
	for _, o := range overlay {
		obj, e := o.Parse(window)
		if e == nil {
			stack.Add(obj)
		}
	}
	return stack, err
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

	//Make sure each of the layouts children have a unique ID by counting the ones of the same type and naming them SceneName.TypeName#Number
	//Create a map of string to int
	counts := map[string]int{}
	//Iterate over the children
	for i, child := range scene.Layout.Children {
		//Check if the child has an ID
		if child.ID == "" {
			//If it doesn't, check the count of the type
			if count, ok := counts[child.Type]; ok {
				//if it does add one to the count and name it SceneName.TypeName#Number
				child.ID = scene.Name + "." + child.Type + "#" + strconv.Itoa(count+1)
				scene.Layout.Children[i] = child
				counts[child.Type] = count + 1
			} else {
				//Set the count to 1 and name it SceneName.TypeName#1
				counts[child.Type] = 1
				child.ID = scene.Name + "." + child.Type + "#1"
				scene.Layout.Children[i] = child
			}
		}
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