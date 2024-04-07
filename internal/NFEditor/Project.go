package NFEditor

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget/CalsWidgets"
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

type ProjectData struct {
	ProjectInfo ProjectInfo
	Project     Project
}

var (
	//go:embed Templates/*/*
	Templates  embed.FS
	ActiveGame ProjectData
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
		return NFError.NewErrProjectNotFound(info.Name)
	}

	//Check if the path ends in .NFProject
	if filepath.Ext(path) != ".NFProject" {
		//Check if the path is a directory and if the directory contains a .NFProject file
		if _, err := os.Stat(path + "/" + filepath.Base(path) + ".NFProject"); os.IsNotExist(err) {
			return NFError.NewErrProjectNotFound(info.Name)
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
func (p Project) Load(window fyne.Window, info ProjectInfo) error {
	ActiveGame.Project = p
	ActiveGame.ProjectInfo = info
	err := p.UpdateProjectInfo(info)
	if err != nil {
		return err
	}
	window.SetContent(CreateSceneEditor(window))
	return nil
}

func (p Project) Create(window fyne.Window) error {
	loadingChannel := make(chan struct{})
	loading := CalsWidgets.NewLoading(loadingChannel, 100*time.Millisecond, 100)
	loading.SetProgress(0, "Creating Project: "+p.GameName)
	//Pop up a dialog with a progress bar and a label that says "Creating Project"
	progressDialog := dialog.NewCustomWithoutButtons("Creating Project", loading.Box, window)
	progressDialog.Show()
	defer progressDialog.Hide()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	novellaForgeDir := homeDir + "/Documents/NovellaForge"
	projectsDir := fyne.CurrentApp().Preferences().StringWithFallback("projectDir", novellaForgeDir+"/projects")

	//First check if the project directory already exists

	loading.SetProgress(10, "Checking if project already exists")
	_, err = os.Stat(projectsDir + "/" + p.GameName)
	if os.IsNotExist(err) {
		err = os.MkdirAll(projectsDir+"/"+p.GameName, 0755)
		if err != nil {
			return err
		}
	} else {
		return NFError.NewErrProjectAlreadyExists(p.GameName)
	}

	//Create the project directory
	loading.SetProgress(20, "Creating Project Directory")
	projectDir := projectsDir + "/" + p.GameName
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		return err
	}

	//Create the project file
	loading.SetProgress(30, "Creating Project File")
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

	percentPerDir := 10 / len(neededDirectories)
	for _, dir := range neededDirectories {
		loading.SetProgress(loading.GetProgress()+float64(percentPerDir), "Creating "+dir)
		err = os.MkdirAll(projectDir+"/"+dir, 0755)
		if err != nil {
			return err
		}
	}

	//Write the main.go file
	loading.SetProgress(50, "Creating main.go file")
	t, err := template.ParseFS(Templates, "Templates/MainGame/Template.go")
	if err != nil {
		return err
	}
	mainGameFile, err := os.Create(projectDir + "/cmd/" + p.GameName + "/" + p.GameName + ".go")
	if err != nil {
		return err
	}

	mainGameData := struct {
		LocalConfig    string
		LocalFunctions string
		LocalLayouts   string
		LocalWidgets   string
	}{
		LocalConfig:    p.GameName + "/internal/config",
		LocalFunctions: p.GameName + "/internal/function",
		LocalLayouts:   p.GameName + "/internal/layout",
		LocalWidgets:   p.GameName + "/internal/widget",
	}

	err = t.Execute(mainGameFile, mainGameData)
	if err != nil {
		return err
	}
	mainGameFile.Close()

	//Write the config file
	loading.SetProgress(60, "Creating Config file")
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
	loading.SetProgress(70, "Creating Custom Import Files")
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

	//Put the scene templates in the data/scenes directory
	loading.SetProgress(75, "Creating Scene Templates")
	//Parse the Scenes/MainMenu.json template
	t, err = template.ParseFS(Templates, "Templates/Scenes/MainMenu.json")
	if err != nil {
		return err
	}
	MainMenuSceneFile, err := os.Create(projectDir + "/data/scenes/MainMenu.NFScene")
	if err != nil {
		return err
	}
	err = t.Execute(MainMenuSceneFile, nil)
	if err != nil {
		return err
	}
	MainMenuSceneFile.Close()

	//Parse the Scenes/NewGame.json template
	t, err = template.ParseFS(Templates, "Templates/Scenes/NewGame.json")
	if err != nil {
		return err
	}
	NewGameSceneFile, err := os.Create(projectDir + "/data/scenes/NewGame.NFScene")
	if err != nil {
		return err
	}
	err = t.Execute(NewGameSceneFile, nil)
	if err != nil {
		return err
	}
	NewGameSceneFile.Close()

	//Initialize the go mod file by running go mod init with os/exec
	loading.SetProgress(80, "Initializing go mod file")

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

	loading.SetProgress(90, "Installing fyne")
	cmd = exec.Command("go", "get", "fyne.io/fyne/v2@latest")
	cmd.Stderr = &stderr
	cmd.Dir = projectDir
	err = cmd.Run()
	if err != nil {
		log.Printf("Error installing fyne: %v", err)
		log.Printf("Stderr: %s", stderr.String())
		return errors.New("error installing fyne")
	}

	loading.SetProgress(95, "Installing NovellaForge")
	cmd = exec.Command("go", "get", "go.novellaforge.dev/novellaforge@latest")
	cmd.Stderr = &stderr
	cmd.Dir = projectDir
	err = cmd.Run()
	if err != nil {
		log.Printf("Error installing NovellaForge: %v", err)
		log.Printf("Stderr: %s", stderr.String())
		return errors.New("error installing NovellaForge")
	}

	//Run go mod tidy
	loading.SetProgress(97, "Running go mod tidy")
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Stderr = &stderr
	cmd.Dir = projectDir
	err = cmd.Run()
	if err != nil {
		log.Printf("Error running go mod tidy: %v", err)
		log.Printf("Stderr: %s", stderr.String())
		return errors.New("error running go mod tidy")

	}

	loading.SetProgress(100, "Project Created")
	time.Sleep(1 * time.Second)
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
