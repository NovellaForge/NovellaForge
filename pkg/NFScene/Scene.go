package NFScene

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Properties map[string]interface{}

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

// NewAndAddScene creates a new scene with the given name, layout, and arguments and adds it to the all map
func NewAndAddScene(name string, layout *NFLayout.Layout, args *NFData.NFInterfaceMap) *Scene {
	scene := NewScene(name, layout, args)
	all[name] = scene
	return scene
}

// all is a map of string to Scene that contains all scenes in the scenes folder
// for this to be repopulated you must call LoadAll
var all = map[string]*Scene{}

// GetAll returns all scenes that have been loaded. If no scenes have been loaded, it will load all scenes from the disk
func GetAll(path ...string) map[string]*Scene {
	log.Println("Getting all Loaded scenes")
	if len(all) == 0 {
		log.Println("No scenes yet loaded, loading all scenes")
		if len(path) == 0 {
			LoadAll()
		} else {
			LoadAll(path[0])
		}
		log.Printf("Loaded %d scenes", len(all))
	}
	return all
}

// LoadAll returns all scenes reloading them from the disk. If you don't want to reload them, use GetAll
func LoadAll(p ...string) map[string]*Scene {
	path := ""
	if len(p) == 0 {
		log.Println("No path provided, using default path")
		path = "./data/scenes"
	} else {
		path = p[0]
	}
	log.Println("Loading all scenes from", path)
	// Start scanning from the base directory
	scenes, err := scanDir(path)
	if err != nil {
		return nil
	} else if scenes == nil {
		return nil
	}
	all = scenes
	return scenes
}

// scanDir scans a directory for scenes
func scanDir(s string, args ...string) (map[string]*Scene, error) {
	//If the args are empty, set findScene to false and then set both args to ""
	findScene := false
	if len(args) > 0 && args[0] != "" {
		findScene = true
	}

	//Create a map of string to Scene
	scenes := map[string]*Scene{}
	//Scan the directory
	log.Println("Scanning directory", s)
	err := filepath.Walk(s, func(path string, info os.FileInfo, err error) error {
		name := ""
		if findScene {
			name = args[0]
			//Make sure the name ends in .NFScene
			if !strings.HasSuffix(name, ".NFScene") {
				name += ".NFScene"
			}
		}
		//Check if the path ends in .NFScene
		if strings.HasSuffix(path, ".NFScene") && !findScene {
			//Load the scene
			scene, err := Import(path)
			if err != nil {
				return err
			}

			//Get the path to the scene starting after the scenes directory
			path = path[len("data/scenes/"):]
			//If the path contains a /, then it is in a group
			splitPath := strings.Split(path, "/")
			if len(splitPath) > 1 {
				//Remove the last element of the path as it is the scene name
				splitPath = splitPath[:len(splitPath)-1]
				//Join the path back together
				path = strings.Join(splitPath, "/")
				//Add the scene to the group
				_ = scene.Args.Add("SceneGroup", path)
			}
			//Add the scene to the map
			scenes[scene.Name] = scene
		} else if findScene && strings.HasSuffix(path, name) {
			//Load the scene
			scene, err := Import(path)
			if err != nil {
				return err
			}

			//Add the scene to the map
			scenes[scene.Name] = scene

		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	//Return the map
	return scenes, nil

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

// Export exports the scene to a file
func (scene *Scene) Export(overwrite bool) error {
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
	//check if the file already exists before writing to it
	if _, err := os.Stat("exports/scenes/" + scene.Name + ".json"); err == nil {
		if !overwrite {
			return errors.New("file already exists")
		}
	}

	// Marshal the jsonScene
	jsonBytes, err := json.MarshalIndent(scene, "", "\t")
	if err != nil {
		return err
	}
	//Create the directories if they don't exist
	err = os.MkdirAll("exports/scenes", 0755)
	//Write the file
	err = os.WriteFile("exports/scenes/"+scene.Name+".json", jsonBytes, 0755)
	if err != nil {
		return err
	}
	return nil
}

// Import imports a scene from a file
func Import(path string) (*Scene, error) {

	//Check if the path ends in .NFScene
	if !strings.HasSuffix(path, ".NFScene") {
		return nil, errors.New("path must end in .NFScene")
	}

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the file into a jsonScene
	var scene Scene
	err = json.NewDecoder(file).Decode(&scene)
	if err != nil {
		log.Println("Error decoding scene")
		return nil, err
	}

	return &scene, nil
}
