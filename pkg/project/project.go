package project

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
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

type Content struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type Scene struct {
	ID          string    `json:"id"`
	SceneType   string    `json:"sceneType"`
	ContentList []Content `json:"contentList"`
}

type Project struct {
	GameName string  `json:"gameName"`
	Version  string  `json:"version"`
	Author   string  `json:"author"`
	Credits  string  `json:"credits"`
	Scenes   []Scene `json:"scenes"`
}

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
		return nil, fmt.Errorf("no projects found")
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

// createProjectFiles creates the project files and directories
func createProjectFiles(projectName string) error {
	// Create project directory
	projectDir := filepath.Join("projects", projectName)
	err := os.MkdirAll(projectDir, 0755)
	if err != nil {
		return err
	}

	// Initialize Go module
	cmd := exec.Command("go", "mod", "init", projectName)
	cmd.Dir = projectDir
	err = cmd.Run()
	if err != nil {
		log.Printf("Could not initialize Go module: %v", err)
		return err
	}

	// Create project.novel file using the Project struct
	projectData := Project{
		GameName: projectName,
		Version:  "0.1",
		Author:   "",
		Credits:  "",
		Scenes:   []Scene{},
	}

	jsonData, err := serializeProject(projectData)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(projectDir, projectName+".novel"), []byte(jsonData), 0644)
	if err != nil {
		return err
	}

	// Create example <ProjectName>.go file
	fileName := fmt.Sprintf("%s.go", projectName)
	gameGoContent :=

		`package main
	
	import "fmt"
	
	func main() {
		fmt.Println("This is your game.")
	}`

	err = os.WriteFile(filepath.Join(projectDir, fileName), []byte(gameGoContent), 0644)
	if err != nil {
		return err
	}

	// Create assets, scenes, and characters directories
	for _, dir := range []string{"assets", "scenes", "characters"} {
		err = os.Mkdir(filepath.Join(projectDir, dir), 0755)
		if err != nil {
			return err
		}
	}

	return nil
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
		err = createProjectFiles(sanitizedName)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		// TODO: Update UI to show new project
	}, w)
}

// ShowOpenProjectDialog opens up a small popup window for opening a project
func ShowOpenProjectDialog(w fyne.Window) {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		//Get the name of the project from the URI cutt off the .novel extension
		projectName := strings.TrimSuffix(reader.URI().Name(), ".novel")
		//Open the project
		err = OpenProject(projectName, w)
	}, w)
	//Set the location to the local projects folder
	listableURI, err := storage.ListerForURI(storage.NewFileURI("projects"))
	if err == nil {
		fd.SetLocation(listableURI)
	}
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".novel"}))
	fd.Show()
}

func OpenProject(projectName string, w fyne.Window) error {
	err := checkAndCreateProjectFiles(projectName)
	if err != nil {
		return err
	}

	return nil
}

// checkAndCreateProjectFiles checks the existence of essential project files and directories,
// and creates them if they are missing.
func checkAndCreateProjectFiles(projectName string) error {
	// Define the list of required directories and files
	requiredDirs := []string{
		"assets",
		"scenes",
		"characters",
	}
	requiredFiles := []string{
		fmt.Sprintf("%s.novel", projectName),
		fmt.Sprintf("%s.go", projectName),
	}

	// Base directory for the project
	projectDir := filepath.Join("projects", projectName)

	// Check and create required directories
	for _, dir := range requiredDirs {
		dirPath := filepath.Join(projectDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			// Directory does not exist, create it
			err = os.Mkdir(dirPath, 0755)
			if err != nil {
				return err
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
				//Create a new project struct and serialize it to json
				projectData := Project{
					GameName: projectName,
					Version:  "0.1",
					Author:   "",
					Credits:  "",
					Scenes:   []Scene{},
				}
				//Serialize the project struct to json
				jsonData, err := serializeProject(projectData)
				if err != nil {
					return err
				}
				//Write the json to the file
				err = os.WriteFile(filePath, []byte(jsonData), 0644)
				if err != nil {
					return err
				}
			case ".go":
				//Create a new go file with a basic main function
				gameGoContent :=
					`package main
	
					import "fmt"
	
					func main() {
						fmt.Println("This is your game.")
					}`

				err = os.WriteFile(filePath, []byte(gameGoContent), 0644)
				if err != nil {
					return err
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
			return err
		}
	}

	return nil
}

// GetScenes reads the .novel file and returns a list of scenes
func GetScenes(projectName string) ([]Scene, error) {
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

	return project.Scenes, nil
}

// populateProjectTree populates the tree view with projects and their scenes
func populateProjectTree(projectName string, w fyne.Window) *widget.Tree {
	// Assuming GetScenes gets the scenes for a project
	scenes, err := GetScenes(projectName)
	if err != nil {
		dialog.ShowError(err, w)
		return nil
	}

	tree := widget.NewTree(
		// ChildUIDs callback
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				return []widget.TreeNodeID{projectName}
			} else if id == projectName {
				var sceneIDs []widget.TreeNodeID
				for _, scene := range scenes {
					sceneIDs = append(sceneIDs, scene.ID)
				}
				return sceneIDs
			}
			return []widget.TreeNodeID{}
		},
		// IsBranch callback
		func(id widget.TreeNodeID) bool {
			return id == "" || id == projectName
		},
		// CreateNode callback
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Placeholder")
		},
		// UpdateNode callback
		func(id widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id == "" {
				label.SetText("Projects")
			} else if id == projectName {
				label.SetText(projectName)
			} else {
				for _, scene := range scenes {
					if scene.ID == id {
						label.SetText(scene.ID) // Or any other property of the scene
					}
				}
			}
		},
	)

	return tree
}
