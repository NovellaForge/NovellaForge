package editor

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type SceneContent struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Text  string      `json:"text"`
	Value interface{} `json:"value"`
}

type Scene struct {
	ID          string         `json:"id"`
	SceneType   string         `json:"sceneType"`
	ContentList []SceneContent `json:"contentList"`
}

type Project struct {
	GameName    string             `json:"gameName"`
	Version     string             `json:"version"`
	Author      string             `json:"author"`
	Credits     string             `json:"credits"`
	SceneGroups map[string][]Scene `json:"sceneGroups"`
}

// CreateDefaultMainMenu creates a default "Main Menu" scene
func CreateDefaultMainMenu() Scene {
	return Scene{
		ID:        "MainMenu",
		SceneType: "Menu",
		ContentList: []SceneContent{
			{
				Name:  "Title",
				Type:  "Label",
				Text:  "Welcome to the Main Menu",
				Value: nil,
			},
			{
				Name:  "StartButton",
				Type:  "Button",
				Text:  "Start",
				Value: "StartGame()",
			},
			{
				Name:  "ExitButton",
				Type:  "Button",
				Text:  "Exit",
				Value: "ExitGame()",
			},
		},
	}
}

// NewProject initializes a new project with default values
func NewProject(gameName string) Project {
	mainMenuScene := CreateDefaultMainMenu()
	return Project{
		GameName: gameName,
		Version:  "1.0",
		Author:   "Anonymous",
		Credits:  "Made with NovellaForge",
		SceneGroups: map[string][]Scene{
			"Menus": {mainMenuScene},
		},
	}
}

const DefaultGameGO = `package main
	
import "fmt"
	
func main() {
	fmt.Println("This is your game.")
}`

var ActiveProject Project

// serializeProject Serialize Project to JSON
func serializeProject(project Project) (string, error) {
	jsonData, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// deserializeProject Deserialize JSON to Project
func deserializeProject(jsonData string) (Project, error) {
	var project Project
	err := json.Unmarshal([]byte(jsonData), &project)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

// ScanProjects scans the projects directory
// and returns a list of projects sorted by last modified time
func ScanProjects() ([]string, error) {
	var projects []string
	var projectNames []string // For storing the final project names

	// Recursive function to walk through directories
	var walkFn filepath.WalkFunc = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(info.Name()) == ".novel" {
			projects = append(projects, path)
		}
		return nil
	}

	err := filepath.Walk("projects", walkFn)
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, nil
	}

	// Sort by last modified time
	sort.Slice(projects, func(i, j int) bool {
		infoI, _ := os.Stat(projects[i])
		infoJ, _ := os.Stat(projects[j])

		// If the time for I does not exist, return false
		if infoI == nil {
			return false
		}

		// If J does not exist, return true
		if infoJ == nil {
			return true
		}

		return infoI.ModTime().After(infoJ.ModTime())
	})

	// Limit to last 10 edited projects
	limit := 10
	if len(projects) < 10 {
		limit = len(projects)
	}

	// Strip directory and '.novel' extension to get the project names
	for _, fullPath := range projects[:limit] {
		projectNameWithExt := filepath.Base(fullPath)
		projectName := strings.TrimSuffix(projectNameWithExt, ".novel")
		projectNames = append(projectNames, projectName)
	}

	return projectNames, nil
}

// sanitizeProjectName sanitizes the project name to ensure it's a valid Go identifier
func sanitizeProjectName(name string) (string, error) {
	// Replace spaces with underscores
	sanitizedName := strings.ReplaceAll(name, " ", "_")

	// Use regex to match valid Go identifiers
	re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`) //TODO we may want to add other languages to this regex if they are supported by go
	if !re.MatchString(sanitizedName) {
		return "", fmt.Errorf("invalid project name. The project name should be a valid Go identifier (Meaning no spaces or special characters except underscores)")
	}

	return sanitizedName, nil
}

func CreateProject(w fyne.Window) {

	// Prompt user for project name
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter Project Name")

	dialog.ShowCustomConfirm("New Project", "Create", "Cancel", nameEntry, func(b bool) {
		if !b {
			return
		}
		projectName := nameEntry.Text
		if projectName == "" {
			dialog.ShowInformation("Invalid Name", "Project name cannot be empty", w)
			return
		}

		// Sanitize and validate project name
		sanitizedName, err := sanitizeProjectName(projectName)
		if err != nil {
			dialog.ShowInformation("Invalid Name", err.Error(), w)
			return
		}

		// Create project files
		err, _ = checkAndCreateProjectFiles(sanitizedName, false)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		// Open the project
		err = OpenProject(sanitizedName, w, false)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

	}, w)
}

// ShowOpenProjectDialog opens up a small popup window for opening a project
func ShowOpenProjectDialog(w fyne.Window) {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {

		//Get the full path to the project file
		projectPath := reader.URI().Path()

		// Load the selected project
		err = OpenProject(projectPath, w, true)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

	}, w)

	// Set the location to the local projects folder
	listableURI, err := storage.ListerForURI(storage.NewFileURI("projects"))
	if err == nil {
		fd.SetLocation(listableURI)
	}
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".novel"}))
	fd.Show()
}

// OpenProject opens a project and makes sure all the required files and directories exist
func OpenProject(projectName string, w fyne.Window, fromPath bool) error {

	// Check and prepare project files
	err, projectPath := checkAndCreateProjectFiles(projectName, fromPath)
	if err != nil {
		return err
	}

	var jsonData []byte
	if fromPath {
		// Get the JSON data from the file
		jsonData, err = os.ReadFile(projectName)
		if err != nil {
			return err
		}
	} else {
		// Get the JSON data from the file by making the path to the project file
		path := filepath.Join("projects", projectName, projectName+".novel")
		jsonData, err = os.ReadFile(path)
		if err != nil {
			return err
		}
	}
	// Deserialize the JSON data to a Project struct
	project, err := deserializeProject(string(jsonData))
	if err != nil {
		return err
	}
	// Set the active project
	ActiveProject = project

	// Populate the tree view for this project
	tree := populateProjectTree(w)
	if tree == nil {
		return fmt.Errorf("failed to create project tree")
	}

	tree.Refresh()

	// Create a split container with the sidebar and main content area
	mainHSplit := container.NewHSplit(tree, container.NewVBox(
		// TODO: Place other UI components here
		//Add a label with the project name
		widget.NewLabel(ActiveProject.GameName),
	))
	mainHSplit.Offset = 0.25 // set the default size of the sidebar to 1/4 of the window

	w.SetContent(mainHSplit)

	//If the project path is not empty, then we need to set the last project to the project path
	if projectPath != "" {
		Config.LastProject = projectPath
		err = SaveConfig(Config)
		if err != nil {
			dialog.ShowError(err, w)
		}
	}

	return nil
}

// checkAndCreateProjectFiles checks the existence of essential project files and directories,
// and creates them if they are missing.
func checkAndCreateProjectFiles(projectName string, fromPath bool) (error, string) {
	// Define the list of required directories and files
	projectDir := ""
	if fromPath {
		//If the project is being opened from a path, then we need to get the true project name and project directory
		projectDir = filepath.Dir(projectName)
		projectName = filepath.Base(projectName)
		//Remove the file extension from the project name
		projectName = strings.TrimSuffix(projectName, filepath.Ext(projectName))
		//Check if the project directory exists
		if _, err := os.Stat(projectDir); os.IsNotExist(err) {
			// Directory does not exist, create it
			err = os.Mkdir(projectDir, 0755)
			if err != nil {
				return err, ""
			}
		}
	} else {
		projectDir = filepath.Join("projects", projectName)
		// Check if the project directory exists
		if _, err := os.Stat(projectDir); os.IsNotExist(err) {
			// Directory does not exist, create it
			err = os.Mkdir(projectDir, 0755)
			if err != nil {
				return err, ""
			}
		}
	}

	requiredDirs := []string{
		"assets",
		"scenes",
		"characters",
	}
	requiredFiles := []string{
		fmt.Sprintf("%s.novel", projectName),
		fmt.Sprintf("%s.go", projectName),
	}

	// Check and create required directories
	for _, dir := range requiredDirs {
		dirPath := filepath.Join(projectDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			// Directory does not exist, create it
			err = os.Mkdir(dirPath, 0755)
			if err != nil {
				return err, ""
			}
		}
	}

	// Check and create required files
	for _, file := range requiredFiles {
		filePath := filepath.Join(projectDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// File does not exist, create it
			//Switch on the file extension to determine what to write to the file
			switch filepath.Ext(file) {
			case ".novel":
				//Serialize the project struct to json
				jsonData, err := serializeProject(NewProject(projectName))
				if err != nil {
					return err, ""
				}
				//Write the json to the file
				err = os.WriteFile(filePath, []byte(jsonData), 0644)
				if err != nil {
					return err, ""
				}
			case ".go":
				err = os.WriteFile(filePath, []byte(DefaultGameGO), 0644)
				if err != nil {
					return err, ""
				}
			}

		}
	}

	// Check for go.mod and run `go mod init` if not found
	goModPath := filepath.Join(projectDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		// go.mod does not exist, initialize Go module
		cmd := exec.Command("go", "mod", "init", projectName)
		cmd.Dir = projectDir
		err = cmd.Run()
		if err != nil {
			log.Printf("Could not initialize Go module: %v", err)
			return err, ""
		}
	}

	//Get the filepath to the .novel file
	projectPath := filepath.Join(projectDir, projectName+".novel")

	return nil, projectPath
}

// GetScenes reads the .novel file and returns a list of scenes for a given group
func _(projectName string, groupName string) ([]Scene, error) {
	// Path to the .novel file
	projectPath := filepath.Join("projects", projectName, projectName+".novel")

	// Read the .novel file
	jsonData, err := os.ReadFile(projectPath)
	if err != nil {
		return nil, err
	}

	// Deserialize the JSON data to a Project struct
	project, err := deserializeProject(string(jsonData))
	if err != nil {
		return nil, err
	}

	scenes, ok := project.SceneGroups[groupName]
	if !ok {
		return nil, fmt.Errorf("scene group '%s' not found in project '%s'", groupName, projectName)
	}

	return scenes, nil
}

func populateProjectTree(_ fyne.Window) *widget.Tree {
	tree := widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			// Handle the root node
			if id == "" {
				return []widget.TreeNodeID{"GameName", "Version", "Author", "Credits", "SceneGroups"}
			}

			// Handle SceneGroups
			if id == "SceneGroups" {
				// Convert map keys to a slice
				var groupNames []widget.TreeNodeID
				for key := range ActiveProject.SceneGroups {
					groupNames = append(groupNames, widget.TreeNodeID(key))
				}
				return groupNames
			}

			// Handle individual SceneGroups
			if scenes, exists := ActiveProject.SceneGroups[string(id)]; exists {
				var sceneIDs []widget.TreeNodeID
				for _, scene := range scenes {
					// Create a unique ID by concatenating the group name and the scene ID
					uniqueID := widget.TreeNodeID(string(id) + "_" + scene.ID)
					sceneIDs = append(sceneIDs, uniqueID)
				}
				return sceneIDs
			}

			return []widget.TreeNodeID{}
		},
		func(id widget.TreeNodeID) bool {
			// The root, SceneGroups, and individual SceneGroups are branches
			if id == "" || id == "SceneGroups" {
				return true
			}
			if _, exists := ActiveProject.SceneGroups[string(id)]; exists {
				return true
			}
			return false
		},
		func(branch bool) fyne.CanvasObject {
			if branch {
				return widget.NewLabel("Branch template")
			}
			return widget.NewLabel("Leaf template")
		},
		func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {

			// Extract the display text from the id
			displayText := id
			if strings.Contains(string(id), "_") {
				parts := strings.Split(string(id), "_")
				if len(parts) > 1 {
					displayText = widget.TreeNodeID(parts[1])
				}
			}
			o.(*widget.Label).SetText(displayText)
		},
	)

	return tree
}
