package NFProject

import (
	"bytes"
	"embed"
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
	"html/template"
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
	//go:embed Templates/*/*
	Templates            embed.FS
	ActiveProject        Project
	ActiveLayouts        []NFLayout.Layout
	ActiveWidgets        []NFWidget.Widget
	ActiveSceneGroups    map[string]NFScene.Scene
	ActiveFunctions      []NFFunction.Function // Ungrouped functions
	ActiveFunctionGroups map[string][]NFFunction.Function
)

func (p Project) UpdateProjectInfo(info ProjectInfo) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	novellaForgeDir := homeDir + "/Documents/NovellaForge"
	projects, err := ReadProjectInfo()
	if err != nil {
		return err
	}

	_, err = os.Stat(info.Path)
	if os.IsNotExist(err) {
		//File does not exist so remove it from the projects.nf file
		for i, project := range projects {
			if project.Name == info.Name {
				projects = append(projects[:i], projects[i+1:]...)
				serializedProjects, err := json.Marshal(projects)
				if err != nil {
					return err
				}
				err = os.WriteFile(novellaForgeDir+"/projects.nf", serializedProjects, 0777)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}

	projectFound := false
	for i, project := range projects {
		if project.Name == info.Name {
			projectFound = true
			projects[i] = info
			serializedProjects, err := json.Marshal(projects)
			if err != nil {
				return err
			}
			err = os.WriteFile(novellaForgeDir+"/projects.nf", serializedProjects, 0777)
			if err != nil {
				return err
			}
			return nil
		}
	}
	if !projectFound {
		projects = append(projects, info)
		serializedProjects, err := json.Marshal(projects)
		if err != nil {
			return err
		}
		err = os.WriteFile(novellaForgeDir+"/projects.nf", serializedProjects, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadProjectInfo reads the project info from the project file
func ReadProjectInfo() ([]ProjectInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	novellaForgeDir := homeDir + "/Documents/NovellaForge"
	//Check if the NovellaForge directory exists
	if _, err := os.Stat(novellaForgeDir); os.IsNotExist(err) {
		//if it doesn't exist, create it
		err = os.MkdirAll(novellaForgeDir, 0777)
		if err != nil {
			return nil, err
		}
	}

	//Check if the projects.nf file exists
	if _, err := os.Stat(novellaForgeDir + "/projects.nf"); os.IsNotExist(err) {
		//if it doesn't exist, create it
		err = os.WriteFile(novellaForgeDir+"/projects.nf", []byte("[]"), 0777)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	//Read the file
	file, err := os.ReadFile(novellaForgeDir + "/projects.nf")
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

func OpenFromInfo(info ProjectInfo, window fyne.Window) error {
	path := info.Path
	//Check if the path even exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		tmpProject := Project{
			GameName: info.Name,
		}
		err = tmpProject.UpdateProjectInfo(info)
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

	err = project.Load(window, info)
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
func (p Project) Load(window fyne.Window, info ...ProjectInfo) error {
	ActiveProject = p
	if len(info) == 0 {
		return nil
	}
	err := p.UpdateProjectInfo(info[0])
	if err != nil {
		return err
	}
	//Load the scenes
	return nil
}

func (p Project) Create(window fyne.Window) error {
	//Pop up a dialog with a progress bar and a label that says "Creating Project"
	progressDialog := dialog.NewCustomWithoutButtons("Creating Project", container.NewVBox(widget.NewProgressBarInfinite(), widget.NewLabel("Creating Project: "+p.GameName)), window)
	progressDialog.Show()
	defer progressDialog.Hide()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	novellaForgeDir := homeDir + "/Documents/NovellaForge"
	projectsDir := fyne.CurrentApp().Preferences().StringWithFallback("projectDir", novellaForgeDir+"/projects")

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
		"internal/function",
		"internal/layout",
		"internal/widget",
	}

	for _, dir := range neededDirectories {
		err = os.MkdirAll(projectDir+"/"+dir, 0755)
		if err != nil {
			return err
		}
	}

	//Write the main.go file
	t, err := template.ParseFS(Templates, "Templates/MainGame/Template.go")
	if err != nil {
		return err
	}
	mainGameFile, err := os.Create(projectDir + "/cmd/" + p.GameName + "/" + p.GameName + ".go")
	if err != nil {
		return err
	}

	mainGameData := struct {
		LocalConfig string
	}{
		LocalConfig: p.GameName + "/internal/config/Config.go",
	}

	err = t.Execute(mainGameFile, mainGameData)
	if err != nil {
		return err
	}
	mainGameFile.Close()

	//Write the config file
	t, err = template.ParseFS(Templates, "Templates/Config/Template.go")
	if err != nil {
		return err
	}
	configData := struct {
		GameName     string
		GameVersion  string
		GameAuthor   string
		GameCredits  string
		StartUpScene string
		NewGameScene string
	}{
		GameName:     p.GameName,
		GameVersion:  "0.0.1",
		GameAuthor:   p.Author,
		GameCredits:  p.Credits,
		StartUpScene: "MainMenu",
		NewGameScene: "NewGame",
	}
	configFile, err := os.Create(projectDir + "/internal/config/Config.go")
	if err != nil {
		return err
	}
	err = t.Execute(configFile, configData)
	if err != nil {
		return err
	}
	configFile.Close()

	//Write the custom import files
	t, err = template.ParseFS(Templates, "Templates/CustomFunction/Template.go")
	if err != nil {
		return err
	}
	customFunctionFile, err := os.Create(projectDir + "/internal/function/CustomFunctions.go")
	if err != nil {
		return err
	}
	err = t.Execute(customFunctionFile, nil)
	if err != nil {
		return err
	}
	customFunctionFile.Close()

	t, err = template.ParseFS(Templates, "Templates/CustomLayout/Template.go")
	if err != nil {
		return err
	}
	customLayoutFile, err := os.Create(projectDir + "/internal/layout/CustomLayouts.go")
	if err != nil {
		return err
	}
	err = t.Execute(customLayoutFile, nil)
	if err != nil {
		return err
	}
	customLayoutFile.Close()

	t, err = template.ParseFS(Templates, "Templates/CustomWidget/Template.go")
	if err != nil {
		return err
	}
	customWidgetFile, err := os.Create(projectDir + "/internal/widget/CustomWidgets.go")
	if err != nil {
		return err
	}
	err = t.Execute(customWidgetFile, nil)
	if err != nil {
		return err
	}
	customWidgetFile.Close()

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
	cmd = exec.Command("go", "get", "fyne.io/fyne/v2@latest")
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
