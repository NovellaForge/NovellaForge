package NFProject

import (
	"bytes"
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/pkg/NFError"
	"github.com/NovellaForge/NovellaForge/pkg/NFFunction"
	"github.com/NovellaForge/NovellaForge/pkg/NFLayout"
	"github.com/NovellaForge/NovellaForge/pkg/NFScene"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type ProjectInfo struct {
	Name     string    `json:"Name"`
	Path     string    `json:"Path"`
	OpenDate time.Time `json:"Last Opened"`
}

type Project struct {
	GameName string `json:"Game Name"`
	Version  string `json:"Version"`
	Author   string `json:"Author"`
	Credits  string `json:"Credits"`
}

var (
	ActiveProject        Project
	ActiveLayouts        []NFLayout.Layout
	ActiveWidgets        []NFWidget.Widget
	ActiveSceneGroups    map[string]NFScene.Scene
	ActiveFunctions      []NFFunction.Function // Ungrouped functions
	ActiveFunctionGroups map[string][]NFFunction.Function
)

// ReadProjectInfo reads the project info from the project file
func ReadProjectInfo() ([]ProjectInfo, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	//Check if the NovellaForge directory exists
	if _, err := os.Stat(cacheDir + "/NovellaForge"); os.IsNotExist(err) {
		//if it doesn't exist, create it
		err = os.Mkdir(cacheDir+"/NovellaForge", 0755)
		if err != nil {
			return nil, err
		}
	}

	//Check if the projects.nf file exists
	if _, err := os.Stat(cacheDir + "/NovellaForge/projects.nf"); os.IsNotExist(err) {
		//if it doesn't exist, create it
		err = os.WriteFile(cacheDir+"/NovellaForge/projects.nf", []byte("[]"), 0644)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	//Read the file
	file, err := os.ReadFile(cacheDir + "/NovellaForge/projects.nf")
	if err != nil {
		return nil, err
	}

	//unmarshal the json into a slice of structs
	var projects []ProjectInfo
	err = json.Unmarshal(file, &projects)
	if err != nil {
		return nil, err
	}

	//return the slice of structs
	return projects, nil
}

func OpenName(name string, window fyne.Window) error {
	//Search for the project in the projects.nf file
	projects, err := ReadProjectInfo()
	if err != nil {
		return err
	}

	//Loop through the projects and find the one with the matching name
	for _, project := range projects {
		if project.Name == name {
			//Open the project
			err = OpenPath(project.Path, window)
			if err != nil {
				return err
			}
			return nil
		}
	}

	//Check the projects directory for a project with the matching name
	configDir, err := os.UserConfigDir()
	projectsDir := fyne.CurrentApp().Preferences().StringWithFallback("projectDir", configDir+"/NovellaForge/projects")

	//Walk the projects directory
	err = filepath.Walk(projectsDir, func(path string, info os.FileInfo, err error) error {
		//Check if the path is a directory
		if info.IsDir() {
			//Check if the directory name matches the name
			if info.Name() == name {
				//Open the project
				err = OpenPath(path, window)
				if err != nil {
					return err
				}
				return nil
			}
		}
		return nil
	})
	return NFError.ErrProjectNotFound
}

func OpenPath(path string, window fyne.Window) error {
	//Check if the path even exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NFError.ErrProjectNotFound
	}

	//Check if the path ends in .NFProject
	if filepath.Ext(path) != ".NFProject" {
		//Check if the path is a directory and if the directory contains a .NFProject file
		if _, err := os.Stat(path + "/" + filepath.Base(path) + ".NFProject"); os.IsNotExist(err) {
			return NFError.ErrProjectNotFound
		} else {
			//If it does, set the path to the directory
			path = path + "/" + filepath.Base(path) + ".NFProject"
		}
	}

	//Read the project file
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	//Deserialize the project
	project, err := Deserialize(file)
	if err != nil {
		return err
	}

	//Load the project
	err = project.Load(window)
	if err != nil {
		return err
	}

	return nil

}

func Deserialize(file []byte) (Project, error) {
	project := Project{}
	err := json.Unmarshal(file, &project)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

// Load takes a deserialized project and loads it into the editor loading the scenes and functions as well
func (p Project) Load(window fyne.Window) error {
	ActiveProject = p
	//Load the scenes
	return NFError.ErrNotImplemented
}

func (p Project) Create(window fyne.Window) error {
	//Pop up a dialog with a progress bar and a label that says "Creating Project"
	progressDialog := dialog.NewCustomWithoutButtons("Creating Project", container.NewVBox(widget.NewProgressBarInfinite(), widget.NewLabel("Creating Project: "+p.GameName)), window)
	progressDialog.Show()
	defer progressDialog.Hide()

	configDir, err := os.UserConfigDir()
	projectsDir := fyne.CurrentApp().Preferences().StringWithFallback("projectDir", configDir+"/NovellaForge/projects")

	_, err = os.Stat(projectsDir + "/" + p.GameName)
	if os.IsNotExist(err) {
		err = os.MkdirAll(projectsDir+"/"+p.GameName, 0755)
		if err != nil {
			return err
		}
	} else {
		return NFError.ErrProjectAlreadyExists
	}

	//First check if the project directory already exists
	projectDir := projectsDir + "/" + p.GameName
	//Create the project directory
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		return err
	}

	//Create the project file
	err = os.WriteFile(projectDir+"/"+p.GameName+".NFProject", p.Serialize(), 0644)
	if err != nil {
		return err
	}

	neededDirectories := []string{
		"cmd/" + p.GameName,
		"data/assets/image",
		"data/assets/audio",
		"data/assets/video",
		"data/assets/other",
		"data/scenes",
		"internal/config",
		"internal/function/handlers",
		"internal/layout/handlers",
		"internal/widget/handlers",
	}

	for _, dir := range neededDirectories {
		err = os.MkdirAll(projectDir+"/"+dir, 0755)
		if err != nil {
			return err
		}
	}

	//Create a default game.go file with an empty main function for now
	err = os.WriteFile(projectDir+"/cmd/"+p.GameName+"/"+p.GameName+".go", []byte(
		`package main`+"\n"+
			`import . "`+p.GameName+`/internal/config"`+"\n"+
			MainGameTemplate), 0666)
	if err != nil {
		return err
	}

	err = os.WriteFile(projectDir+"/internal/config/Config.go", []byte(
		`package config`+"\n"+
			`const (`+"\n"+
			`GameName = "`+p.GameName+`"`+"\n"+
			`GameVersion = "0.0.1"`+"\n"+
			`GameAuthor = "`+p.Author+`"`+"\n"+
			`GameCredits = "`+p.Credits+`"`+"\n"+
			`StartupScene = "MainMenu""`+"\n"+
			`NewGameScene = "NewScene"`+"\n"), 0666)
	if err != nil {
		return err
	}

	err = os.WriteFile(projectDir+"/internal/function/CustomFunctions.go", []byte(
		`package Functions`+"\n"+
			`import . "`+p.GameName+`/internal/function/handlers"`+"\n"+
			CustomFunctionTemplate), 0666)
	if err != nil {
		return err
	}

	err = os.WriteFile(projectDir+"/internal/layout/CustomLayouts.go", []byte(
		`package Layouts`+"\n"+
			`import . "`+p.GameName+`/internal/layout/handlers"`+"\n"+
			CustomLayoutTemplate), 0666)
	if err != nil {
		return err
	}

	err = os.WriteFile(projectDir+"/internal/widget/CustomWidgets.go", []byte(
		`package Widgets`+"\n"+
			`import . "`+p.GameName+`/internal/widget/handlers"`+"\n"+
			CustomWidgetTemplate), 0666)
	if err != nil {
		return err
	}

	//Write the example for each of the files
	err = os.WriteFile(projectDir+"/internal/function/handlers/ExampleFunction.go", []byte(ExampleFunctionTemplate), 0666)
	if err != nil {
		return err
	}
	err = os.WriteFile(projectDir+"/internal/layout/handlers/ExampleLayout.go", []byte(ExampleLayoutTemplate), 0666)
	if err != nil {
		return err
	}
	err = os.WriteFile(projectDir+"/internal/widget/handlers/ExampleWidget.go", []byte(
		`package handlers`+"\n"+
			`import (`+"\n"+p.GameName+`/internal/function/handlers"`+"\n"+
			ExampleWidgetTemplate), 0666)
	if err != nil {
		return err
	}

	//Initialize the go mod file by running go mod init with os/exec
	log.Printf("Initializing go mod file")

	// Initialize the go mod
	var stderr bytes.Buffer
	cmd := exec.Command("go", "mod", "init", p.GameName)
	cmd.Stderr = &stderr
	cmd.Dir = projectDir
	err = cmd.Run()
	if err != nil {
		log.Printf("Error initializing go mod file: %v", err)
		log.Printf("Stderr: %s", stderr.String())
		return errors.New("error initializing go mod file")
	}

	log.Printf("Installing fyne")
	cmd = exec.Command("go", "get", "fyne.io/fyne/v2")
	cmd.Stderr = &stderr
	cmd.Dir = projectDir
	err = cmd.Run()
	if err != nil {
		log.Printf("Error installing fyne: %v", err)
		log.Printf("Stderr: %s", stderr.String())
		return errors.New("error installing fyne")
	}

	log.Printf("Installing NovellaForge")
	cmd = exec.Command("go", "get", "github.com/NovellaForge/NovellaForge")
	cmd.Stderr = &stderr
	cmd.Dir = projectDir
	err = cmd.Run()
	if err != nil {
		log.Printf("Error installing NovellaForge: %v", err)
		log.Printf("Stderr: %s", stderr.String())
		return errors.New("error installing NovellaForge")
	}

	log.Printf("Initialization successful")
	//Wait for 2 seconds to finish the progress bar and make sure everything is done
	time.Sleep(1 * time.Second)

	//Open the project
	err = OpenPath(projectDir+"/"+p.GameName+".NFProject", window)
	if err != nil {
		return err
	}

	return nil
}

func (p Project) Serialize() []byte {
	//Marshal the project to JSON
	serializedProject, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return nil
	}
	return serializedProject

}
