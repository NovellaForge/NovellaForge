package NFScene

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/NovellaForge/NovellaForge/pkg/NFLayout"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Properties map[string]interface{}

type Scene struct {
	Name       string          `json:"Name"`
	Layout     NFLayout.Layout `json:"Layout"`
	Properties Properties      `json:"Properties"`
}

// All is a map of string to Scene that contains all scenes in the scenes folder for this to be populated you must call GetAll at least once
var All = map[string]Scene{}

// GetAll returns all scenes reloading them from the disk, if you don't want to reload them, just use All
func GetAll() map[string]Scene {
	scenes := make(map[string]Scene)
	loadScene := func(path string) {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Error reading file: %v", err)
			return
		}

		var scene Scene
		err = json.Unmarshal(data, &scene)
		if err != nil {
			log.Printf("Error unmarshalling JSON: %v", err)
			return
		}

		scenes[scene.Name] = scene
	}

	// Recursive function to scan directories
	var scanDir func(string)
	scanDir = func(dir string) {
		files, err := os.ReadDir(dir)
		if err != nil {
			log.Printf("Error reading directory: %v", err)
			return
		}

		for _, f := range files {
			fullPath := filepath.Join(dir, f.Name())
			if f.IsDir() {
				scanDir(fullPath)
			} else if strings.HasSuffix(f.Name(), ".NFScene") {
				loadScene(fullPath)
			}
		}
	}

	// Start scanning from the base directory
	scanDir("data/scenes")
	All = scenes
	return scenes
}

// Parse parses a scene and returns a fyne.CanvasObject that can be added to the window
func (scene *Scene) Parse(window fyne.Window) (fyne.CanvasObject, error) {
	//TODO Add an overlay option that allows certain overlay layouts to be parsed on top of the scene
	layout, err := scene.Layout.Parse(window)
	return container.NewStack(layout), err
}
