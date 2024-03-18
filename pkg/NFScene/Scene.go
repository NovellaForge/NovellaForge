package NFScene

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/NovellaForge/NovellaForge/pkg/NFLayout"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Properties map[string]interface{}

type Scene struct {
	Name       string          `json:"Name"`
	Layout     NFLayout.Layout `json:"Layout"`
	Properties Properties      `json:"Properties"`
}

// make sure that the compiler doesn't complain about the unused functions since they are meant for external use
func _() {
	GetAll()
	LoadAll()
	_, _ = Load("An example path")
	_, _ = LoadByName("An example name")
}

// all is a map of string to Scene that contains all scenes in the scenes folder
// for this to be repopulated you must call LoadAll
var all = map[string]Scene{}

func GetAll(path ...string) map[string]Scene {
	log.Println("Getting all Loaded scenes")
	if len(all) == 0 {
		log.Println("No scenes yet loaded, loading all scenes")
		LoadAll(path[0])
		log.Printf("Loaded %d scenes", len(all))
	}
	return all
}

// LoadAll returns all scenes reloading them from the disk. If you don't want to reload them, use GetAll
func LoadAll(path ...string) map[string]Scene {
	if len(path) == 0 {
		log.Println("No path provided, using default path")
		path = append(path, "data/scenes")
	}
	// Start scanning from the base directory
	scenes, err := scanDir(path[0])
	if err != nil {
		return nil
	} else if scenes == nil {
		return nil
	}
	all = scenes
	return scenes

}

func Load(path string) (Scene, error) {
	//Check if the path ends in .NFScene
	if !strings.HasSuffix(path, ".NFScene") {
		return Scene{}, errors.New("path must end in .NFScene")
	}

	//Open the file
	file, err := os.Open(path)
	if err != nil {
		return Scene{}, err
	}

	//Decode the file
	var scene Scene
	err = json.NewDecoder(file).Decode(&scene)
	if err != nil {
		return Scene{}, err
	}

	//Return the scene
	return scene, nil
}

func LoadByName(name string) (Scene, error) {
	//Check if the name ends in .NFScene
	if !strings.HasSuffix(name, ".NFScene") {
		name += ".NFScene"
	}

	//Check if the name contains / or \ and if it does run the load by path function
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return Load(name)
	}

	//Scan the data/scenes directory
	scenes, err := scanDir("data/scenes", name)
	if err != nil {
		return Scene{}, err
	}
	//get the first scene
	for _, scene := range scenes {
		return scene, nil
	}
	return Scene{}, errors.New("scene not found")
}

func scanDir(s string, args ...string) (map[string]Scene, error) {
	//If the args are empty, set findScene to false and then set both args to ""
	findScene := false
	if args[0] != "" {
		findScene = true
	}
	//Create a map of string to Scene
	scenes := map[string]Scene{}
	//Scan the directory
	err := filepath.Walk(s, func(path string, info os.FileInfo, err error) error {
		name := args[0]
		if findScene {
			//Make sure the name ends in .NFScene
			if !strings.HasSuffix(name, ".NFScene") {
				name += ".NFScene"
			}
		}
		//Check if the path ends in .NFScene
		if strings.HasSuffix(path, ".NFScene") && !findScene {
			//Load the scene
			scene, err := Load(path)
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
				scene.Properties["Group"] = path
			}
			//Add the scene to the map
			scenes[scene.Name] = scene
		} else if findScene && strings.HasSuffix(path, name) {
			//Load the scene
			scene, err := Load(path)
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

// Parse parses a scene and returns a fyne.CanvasObject that can be added to the window each argument passed should be an overlay, with the first being the bottom most overlay
func (scene *Scene) Parse(window fyne.Window, overlay ...fyne.CanvasObject) (fyne.CanvasObject, error) {
	stack := container.NewStack()
	layout, err := scene.Layout.Parse(window)
	stack.Add(layout)
	for _, o := range overlay {
		stack.Add(o)
	}
	return stack, err
}

// Export exports the scene to a file
func (scene *Scene) Export() error {
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
		return errors.New("file already exists")
	}
	//Create the jsonBytes for the scene
	jsonBytes, err := json.MarshalIndent(scene, "", "	")
	if err != nil {
		return err
	}
	//Create the directories if they don't exist
	err = os.MkdirAll("exports/scenes", 0755)
	//Write the file
	err = os.WriteFile("exports/scenes/"+scene.Name+".json", jsonBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}
