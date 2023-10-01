package NFEditor

import (
	error2 "NovellaForge/pkg/NFError"
	"NovellaForge/pkg/NFLayout"
	"NovellaForge/pkg/NFScene"
	"NovellaForge/pkg/NFWidget"
	"bytes"
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
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

var (
	ActiveProject        Project
	ActiveLayouts        []NFLayout.Layout
	ActiveWidgets        []NFWidget.Widget
	ActiveSceneGroups    []SceneGroup
	ActiveScenes         []NFScene.Scene // Ungrouped scenes
	ActiveFunctionGroups []FunctionGroup
	ActiveFunctions      []Function // Ungrouped functions
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

func OpenProject(path string, window fyne.Window) error {
	//Check if the path even exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return error2.ErrProjectNotFound
	}

	//Check if the path ends in .NFProject
	if filepath.Ext(path) != ".NFProject" {
		//Check if the path is a directory and if the directory contains a .NFProject file
		if _, err := os.Stat(path + "/" + filepath.Base(path) + ".NFProject"); os.IsNotExist(err) {
			return error2.ErrProjectNotFound
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
	project, err := DeserializeProject(file)
	if err != nil {
		return err
	}

	//Load the project
	err = LoadProject(project, window)
	if err != nil {
		return err
	}

	return nil

}

func DeserializeProject(file []byte) (Project, error) {
	project := Project{}
	err := json.Unmarshal(file, &project)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

// LoadProject takes a deserialized project and loads it into the editor loading the scenes and functions as well
func LoadProject(project Project, window fyne.Window) error {
	return error2.ErrNotImplemented
}

func CreateProject(project Project, window fyne.Window) error {
	//Pop up a dialog with a progress bar and a label that says "Creating Project"
	progressDialog := dialog.NewCustomWithoutButtons("Creating Project", container.NewVBox(widget.NewProgressBarInfinite(), widget.NewLabel("Creating Project: "+project.GameName)), window)
	progressDialog.Show()
	defer progressDialog.Hide()

	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	//Check if the NovellaForge directory exists
	if _, err = os.Stat(configDir + "/NovellaForge"); os.IsNotExist(err) {
		//if it doesn't exist, create it
		err = os.Mkdir(configDir+"/NovellaForge", 0755)
		if err != nil {
			return err
		}
	}

	//Check if the projects directory exists
	projectsDir := configDir + "/NovellaForge/projects"
	if _, err = os.Stat(projectsDir); os.IsNotExist(err) {
		//if it doesn't exist, create it
		err = os.Mkdir(projectsDir, 0755)
		if err != nil {
			return err
		}
	}

	//First check if the project directory already exists
	projectDir := projectsDir + "/" + project.GameName
	if _, err = os.Stat(projectDir); !os.IsNotExist(err) {
		return error2.ErrProjectAlreadyExists
	}

	//Create the project directory
	err = os.Mkdir(projectDir, 0755)
	if err != nil {
		return err
	}

	//Create the project file
	err = os.WriteFile(projectDir+"/"+project.GameName+".NFProject", SerializeProject(project), 0644)
	if err != nil {
		return err
	}

	//Create a default game.go file with an empty main function for now
	err = os.WriteFile(projectDir+"/"+project.GameName+".go", []byte("package main\n\nfunc main() {\n\n}"), 0644)
	if err != nil {
		return err
	}

	//Create all necessary directories
	neededDirectories := []string{"data/scenes", "data/assets/images", "data/assets/sounds", "data/assets/videos", "data/assets/other"}
	//first create the data directory
	err = os.Mkdir(projectDir+"/data", 0755)
	if err != nil {
		return err
	}

	//Then the data/assets directory
	err = os.Mkdir(projectDir+"/data/assets", 0755)
	if err != nil {
		return err
	}

	//then create the subdirectories
	for _, directory := range neededDirectories {
		err = os.Mkdir(projectDir+"/"+directory, 0755)
		if err != nil {
			return err
		}
	}

	//Then create the default go files in the necessary directories
	goDirectories := []string{"layouts", "scenes", "functions", "widgets"}
	for _, directory := range goDirectories {
		//First create the directory
		err = os.Mkdir(projectDir+"/data/"+directory, 0755)
		if err != nil {
			return err
		}

		//Then create the default go file
		err = os.WriteFile(projectDir+"/data/"+directory+"/"+"defaults.go", []byte("package "+directory+"\n\n"), 0644)
		if err != nil {
			return err
		}
	}

	//Initialize the go mod file by running go mod init with os/exec
	log.Printf("Initializing go mod file")

	// Initialize the go mod
	var stderr bytes.Buffer
	cmd := exec.Command("go", "mod", "init", project.GameName)
	cmd.Stderr = &stderr
	cmd.Dir = projectDir
	err = cmd.Run()
	if err != nil {
		log.Printf("Error initializing go mod file: %v", err)
		log.Printf("Stderr: %s", stderr.String())
		return errors.New("NFerror initializing go mod file")
	}

	//TODO: Find a way around downloading fyne every time a project is created in order to not require an internet connection
	log.Printf("Installing fyne")
	cmd = exec.Command("go", "get", "fyne.io/fyne/v2")
	cmd.Stderr = &stderr
	cmd.Dir = projectDir
	err = cmd.Run()
	if err != nil {
		log.Printf("Error installing fyne: %v", err)
		log.Printf("Stderr: %s", stderr.String())
		return errors.New("NFerror installing fyne")
	}

	log.Printf("Initialization successful")
	//Wait for 2 seconds to finish the progress bar and make sure everything is done
	time.Sleep(1 * time.Second)

	return nil

	//TODO: Load the project into the editor and open the project Before that can be done the default files need to be created and templates need to be created for the files
	// Also add the project to the projects.nf file in the cache directory
}

func SerializeProject(project Project) []byte {
	//Marshal the project to JSON
	serializedProject, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return nil
	}
	return serializedProject

}
