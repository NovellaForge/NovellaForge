package NFEditor

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFConfig"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFWidget/CalsWidgets"
	"html/template"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type NFInfo struct {
	Name     string    `json:"Name"`
	Path     string    `json:"Path"`
	OpenDate time.Time `json:"Last Opened"`
}

type NFProject struct {
	Info   NFInfo
	Config *NFConfig.NFConfig
}

var (
	//go:embed Templates/*
	Templates     embed.FS
	ActiveProject *NFProject
)

func UpdateProjectInfo(info NFInfo) error {
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
func ReadProjectInfo() ([]NFInfo, error) {
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
	var projects []NFInfo
	err = json.Unmarshal(file, &projects)
	if err != nil {
		return nil, err
	}

	//Sort the projects by the last opened date
	for i := 0; i < len(projects); i++ {
		for j := 0; j < len(projects); j++ {
			if projects[i].OpenDate.After(projects[j].OpenDate) {
				temp := projects[i]
				projects[i] = projects[j]
				projects[j] = temp
			}
		}
	}

	//return the slice of structs
	return projects, nil
}

func OpenFromInfo(info NFInfo, window fyne.Window) error {
	path := info.Path
	//Check if the path even exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		newError := UpdateProjectInfo(info)
		if newError != nil {
			errors.Join(err, newError, NFError.NewErrProjectNotFound(info.Name))
		}
		return err
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

	err = project.Load(window)
	if err != nil {
		return err
	}

	return nil

}

func Deserialize(file []byte) (project NFInfo, err error) {
	err = json.Unmarshal(file, &project)
	if err != nil {
		return NFInfo{}, err
	}
	return project, nil
}

// Load takes a deserialized project and loads it into the editor loading the scenes and functions as well
func (p NFInfo) Load(window fyne.Window) error {
	Project := &NFProject{
		Info: p,
	}
	//Walk the game directory of the project for the .NFConfig file
	//If it doesn't exist, return an error
	gameDir := filepath.Dir(p.Path) + "/data/"
	//Look for the first file ending in .NFConfig
	err := filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".NFConfig" {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			err = Project.Config.Load(file)
			if err != nil {
				return err
			}
		}
		return nil
	})
	err = UpdateProjectInfo(p)
	if err != nil {
		return err
	}
	ActiveProject = Project
	window.SetContent(CreateSceneEditor(window))
	return nil
}

func (p NFProject) Create(window fyne.Window) error {
	shouldDelete := false
	deletePath := ""
	defer func() {
		if shouldDelete && deletePath != "" {
			deletePath = filepath.Clean(deletePath)
			if fs.ValidPath(deletePath) {
				vbox := container.NewVBox(
					widget.NewLabel("An error occurred while creating the project."),
					widget.NewLabel("Would you like to delete the broken project?"),
					widget.NewLabel("Delete will remove: "+deletePath),
					widget.NewLabel("This action cannot be undone."),
				)
				dialog.ShowCustomConfirm("Delete Broken Project?", "Yes", "No", vbox, func(b bool) {
					if b {
						err := os.RemoveAll(deletePath)
						if err != nil {
							dialog.ShowError(err, window)
						}
					}
				}, window)
			}
		}
	}()
	loadingChannel := make(chan struct{})
	loading := CalsWidgets.NewLoading(loadingChannel, 100*time.Millisecond, 100)
	loading.SetProgress(0, "Creating Project: "+p.Config.Name)
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

	projectDir := projectsDir + "/" + p.Config.Name
	projectDir = filepath.Clean(projectDir)
	if !fs.ValidPath(projectDir) {
		return errors.New("invalid path")
	}

	loading.SetProgress(10, "Checking if project already exists")
	_, err = os.Stat(projectDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(projectDir, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		return NFError.NewErrProjectAlreadyExists(p.Config.Name)
	}

	//Create the project directory
	loading.SetProgress(20, "Creating Project Directory")
	err = os.MkdirAll(projectDir, os.ModePerm)
	deletePath = projectDir
	if err != nil {
		shouldDelete = true
		return err
	}

	//Create the project file
	loading.SetProgress(30, "Creating Project File")
	projectFilePath := projectDir + "/" + p.Config.Name + ".NFProject"
	projectFilePath = filepath.Clean(projectFilePath)
	if !fs.ValidPath(projectFilePath) {
		shouldDelete = true
		return errors.New("invalid path")
	}
	err = os.WriteFile(projectFilePath, p.SerializeInfo(), os.ModePerm)
	if err != nil {
		shouldDelete = true
		return err
	}

	neededDirectories := []string{
		"cmd/" + p.Config.Name,
		"data/assets/image",
		"data/assets/audio",
		"data/assets/video",
		"data/assets/other",
		"data/scenes",
		"internal/functions",
		"internal/layouts",
		"internal/widgets",
	}

	percentPerDir := 10 / len(neededDirectories)
	for _, dir := range neededDirectories {
		dirPath := projectDir + "/" + dir
		dirPath = filepath.Clean(dirPath)
		if !fs.ValidPath(dirPath) {
			shouldDelete = true
			return errors.New("invalid path")
		}
		loading.SetProgress(loading.GetProgress()+float64(percentPerDir), "Creating "+dir)
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			shouldDelete = true
			return err
		}
	}

	loading.SetProgress(50, "Saving Local config")
	err = p.Config.Save(projectDir + "/data/Game.NFConfig")
	if err != nil {
		shouldDelete = true
		return err
	}

	// templateCombo is a struct that holds the template path, destination path, and data to be used in the template
	type templateCombo struct {
		templatePath    string
		destinationPath string
		data            interface{}
	}

	// neededFiles is a slice of templateCombos that holds all the files that need to be created
	var neededFiles []templateCombo

	//cmd/ProjectName/ProjectName.go
	mainGameData := struct {
		LocalFileSystem string
		LocalFunctions  string
		LocalLayouts    string
		LocalWidgets    string
	}{
		LocalFileSystem: p.Config.Name + "/data",
		LocalFunctions:  p.Config.Name + "/internal/functions",
		LocalLayouts:    p.Config.Name + "/internal/layouts",
		LocalWidgets:    p.Config.Name + "/internal/widgets",
	}
	neededFiles = append(neededFiles, templateCombo{
		templatePath:    "Templates/MainGame/MainGame.got",
		destinationPath: projectDir + "/cmd/" + p.Config.Name + "/" + p.Config.Name + ".go",
		data:            mainGameData,
	})

	//data/FileSystem.go
	neededFiles = append(neededFiles, templateCombo{
		templatePath:    "Templates/FileLoader/FileLoader.got",
		destinationPath: projectDir + "/data/FileSystem.go",
		data:            nil,
	})

	//internal/functions/CustomFunctions.go
	neededFiles = append(neededFiles, templateCombo{
		templatePath:    "Templates/CustomFunction/CustomFunction.got",
		destinationPath: projectDir + "/internal/functions/CustomFunction.go",
		data:            nil,
	})

	//internal/layouts/CustomLayouts.go
	neededFiles = append(neededFiles, templateCombo{
		templatePath:    "Templates/CustomLayout/CustomLayout.got",
		destinationPath: projectDir + "/internal/layouts/CustomLayout.go",
		data:            nil,
	})

	//internal/widgets/CustomWidgets.go
	neededFiles = append(neededFiles, templateCombo{
		templatePath:    "Templates/CustomWidget/CustomWidget.got",
		destinationPath: projectDir + "/internal/widgets/CustomWidget.go",
		data:            nil,
	})

	//data/scenes/MainMenu.NFScene
	neededFiles = append(neededFiles, templateCombo{
		templatePath:    "Templates/Scenes/MainMenu.NFScene",
		destinationPath: projectDir + "/data/scenes/MainMenu.NFScene",
		data:            nil,
	})

	//data/scenes/NewGame.NFScene
	neededFiles = append(neededFiles, templateCombo{
		templatePath:    "Templates/Scenes/NewGame.NFScene",
		destinationPath: projectDir + "/data/scenes/NewGame.NFScene",
		data:            nil,
	})

	percentPerFile := 30 / len(neededFiles)
	for _, file := range neededFiles {
		pathWithoutProject := strings.TrimPrefix(file.destinationPath, projectDir)
		loading.SetProgress(loading.GetProgress()+float64(percentPerFile), "Creating "+pathWithoutProject)
		err = MakeFromTemplate(file.templatePath, file.destinationPath, file.data)
		if err != nil {
			shouldDelete = true
			return err
		}
	}

	//Initialize the go mod file by running go mod init with os/exec
	loading.SetProgress(80, "Initializing go mod file")

	// Initialize the go mod
	var stderr bytes.Buffer
	cmd := exec.Command("go", "mod", "init", p.Config.Name)
	cmd.Stderr = &stderr
	cmd.Dir = projectDir
	err = cmd.Run()
	if err != nil {
		shouldDelete = true
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

func (p NFProject) SerializeInfo() []byte {
	//Marshal the project to JSON
	serializedProject, err := json.MarshalIndent(p.Info, "", "  ")
	if err != nil {
		return nil
	}
	return serializedProject
}

func MakeFromTemplate(templatePath, destinationPath string, data interface{}) error {
	destinationPath = filepath.Clean(destinationPath)
	if !fs.ValidPath(destinationPath) {
		return errors.New("invalid path")
	}
	_, err := fs.Stat(Templates, templatePath)
	if err != nil {
		return err
	}

	destinationFile, err := os.Create(destinationPath)
	defer destinationFile.Close()
	if err != nil {
		return err
	}
	t, err := template.ParseFS(Templates, templatePath)
	if err != nil {
		return err
	}
	err = t.Execute(destinationFile, data)
	if err != nil {
		return err
	}
	return nil
}
