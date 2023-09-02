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
	GameName          string             `json:"gameName"`
	Version           string             `json:"version"`
	Author            string             `json:"author"`
	Credits           string             `json:"credits"`
	SceneGroups       map[string][]Scene `json:"sceneGroups"`
	SceneGroupIndices map[string]int     `json:"sceneGroupIndices"`
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
		SceneGroupIndices: map[string]int{
			"Menus": 0,
		},
	}
}

const DefaultGameGO = `package main
	
import "fmt"
	
func main() {
	fmt.Println("This is your game.")
}`

var ActiveProject Project
var ActiveProjectPath string

// serializeProject Serialize Project to JSON
func serializeProject(project Project) (string, error) {

	validatedProject := ValidateProject(project)

	jsonData, err := json.MarshalIndent(validatedProject, "", "  ")
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

	validatedProject := ValidateProject(project)
	return validatedProject, nil
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
	ActiveProjectPath = projectPath

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

	//After all that save the project just in case
	err = SaveProject()
	if err != nil {
		dialog.ShowError(err, w)
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

func populateProjectTree(w fyne.Window) *widget.Tree {
	tree := widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			// Handle the root node
			if id == "" {
				return []widget.TreeNodeID{"GameName", "Version", "Author", "Credits", "SceneGroups"}
			}

			// Handle SceneGroups
			if id == "SceneGroups" {
				// Convert map keys to a slice
				var groupNames []string
				for key := range ActiveProject.SceneGroups {
					groupNames = append(groupNames, key)
				}

				// Sort groupNames based on indices
				sort.Slice(groupNames, func(i, j int) bool {
					return ActiveProject.SceneGroupIndices[groupNames[i]] < ActiveProject.SceneGroupIndices[groupNames[j]]
				})

				// Convert sorted group names to TreeNodeID slice
				var sortedGroupNames []widget.TreeNodeID
				for _, name := range groupNames {
					sortedGroupNames = append(sortedGroupNames, name)
				}

				return sortedGroupNames
			}

			// Handle individual SceneGroups
			if scenes, exists := ActiveProject.SceneGroups[id]; exists {
				var sceneIDs []widget.TreeNodeID
				for _, scene := range scenes {
					// Create a unique ID by concatenating the group name and the scene ID
					uniqueID := string(id) + "_" + scene.ID
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
			if _, exists := ActiveProject.SceneGroups[id]; exists {
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
			if strings.Contains(id, "_") {
				parts := strings.Split(id, "_")
				if len(parts) > 1 {
					displayText = parts[1]
				}
			}

			switch id {
			case "GameName":
				displayText = "Game Name: " + ActiveProject.GameName
			case "Version":
				displayText = "Version: " + ActiveProject.Version
			case "Author":
				displayText = "Author: " + ActiveProject.Author
			}

			o.(*widget.Label).SetText(displayText)
		},
	)

	tree.OnSelected = func(id string) {
		if id == "SceneGroups" {
			nameEntry := widget.NewEntry()
			nameEntry.SetPlaceHolder("Name")
			dialog.ShowCustomConfirm("New Scene Group", "Add", "Cancel", nameEntry, func(b bool) {
				if !b {
					return
				}
				groupName := strings.ReplaceAll(nameEntry.Text, "_", " ")
				if groupName == "" {
					dialog.ShowInformation("Invalid Name", "Scene Group name cannot be empty", w)
					return
				}

				//Make sure the scene group name does not already exist
				if _, exists := ActiveProject.SceneGroups[groupName]; exists {
					dialog.ShowInformation("Invalid Name", "Scene Group already exists", w)
					return
				}

				//Append the new scene group to the project
				ActiveProject.SceneGroups[groupName] = []Scene{}

				err := SaveProject()
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				tree.Refresh()

			}, w)
		} else if _, exists := ActiveProject.SceneGroups[id]; exists {
			// This is a Scene Group; open dialog for editing or deleting
			// TODO: Open dialog to either delete the Scene Group or add a new Scene
		} else if strings.Contains(id, "_") {
			// This is a Scene; open dialog for editing, renaming or deleting
			// TODO: Open dialog to edit, rename, or delete the Scene
		} else if id == "GameName" || id == "Version" || id == "Author" {
			// Open the dialog
			valueEntry := widget.NewEntry()
			valueEntry.SetPlaceHolder(fmt.Sprintf("Edit %s", id))

			dialog.ShowCustomConfirm(fmt.Sprintf("Edit %s", id), "Update", "Cancel", valueEntry, func(b bool) {
				if !b {
					return
				}
				newValue := valueEntry.Text
				if newValue == "" {
					dialog.ShowInformation("Invalid Value", fmt.Sprintf("%s cannot be empty", id), w)
					return
				}

				switch id {
				case "GameName":
					ActiveProject.GameName = newValue
				case "Version":
					ActiveProject.Version = newValue
				case "Author":
					ActiveProject.Author = newValue
				}

				err := SaveProject()
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				tree.Refresh()
			}, w)
		}
		tree.UnselectAll()
	}

	return tree
}

// SaveProject saves the active project to the .novel file
func SaveProject() error {
	// Serialize the project struct to JSON
	jsonData, err := serializeProject(ActiveProject)
	if err != nil {
		return err
	}

	// Write the JSON data to the file
	err = os.WriteFile(ActiveProjectPath, []byte(jsonData), 0644)
	if err != nil {
		return err
	}

	return nil
}

// ValidateProject ensures that there are no duplicate scene groups and no duplicate scenes within a specific scene group.
// It takes in a project as a parameter and returns the modified project.
func ValidateProject(project Project) Project {
	// A map to keep track of scene group names and their counts
	sceneGroupCount := make(map[string]int)
	// A new map to store the validated scene groups
	validatedSceneGroups := make(map[string][]Scene)
	for group, scenes := range project.SceneGroups {
		// Remove underscores from scene group names
		cleanGroupName := strings.ReplaceAll(group, "_", " ")

		// Resolve duplicate scene group names
		if count, exists := sceneGroupCount[cleanGroupName]; exists {
			cleanGroupName = fmt.Sprintf("%s%d", cleanGroupName, count+1)
			sceneGroupCount[cleanGroupName] = count + 1
		} else {
			sceneGroupCount[cleanGroupName] = 1
		}

		// A map to keep track of scene IDs and their counts within this group
		sceneIDCount := make(map[string]int)
		// A new slice to store the validated scenes for this group
		var validatedScenes []Scene

		for _, scene := range scenes {
			// Remove underscores from scene IDs
			cleanSceneID := strings.ReplaceAll(scene.ID, "_", "")

			// Resolve duplicate scene IDs within this group
			if count, exists := sceneIDCount[cleanSceneID]; exists {
				cleanSceneID = fmt.Sprintf("%s%d", cleanSceneID, count+1)
				sceneIDCount[cleanSceneID] = count + 1
			} else {
				sceneIDCount[cleanSceneID] = 1
			}

			// Update the scene ID and add to the validated scenes slice
			scene.ID = cleanSceneID
			validatedScenes = append(validatedScenes, scene)
		}

		// Add the validated scenes to the validated scene groups map
		validatedSceneGroups[cleanGroupName] = validatedScenes
	}
	// Update the project with the validated scene groups
	project.SceneGroups = validatedSceneGroups
	updatedProject := UpdateSceneGroupIndices(project)
	return updatedProject
}

func UpdateSceneGroupIndices(project Project) Project {
	//Make sure the scene group indices map exists
	if project.SceneGroupIndices == nil {
		project.SceneGroupIndices = make(map[string]int)
	}
	// Step 2: Remove duplicate indices and maintain order
	usedIndexes := make(map[int]bool)
	for {
		duplicateFound := false
		for _, index := range project.SceneGroupIndices {
			if usedIndexes[index] {
				duplicateFound = true
				// Increment this index and all subsequent indices by 1
				for g, i := range project.SceneGroupIndices {
					if i >= index {
						project.SceneGroupIndices[g] = i + 1
					}
				}
				break
			}
			usedIndexes[index] = true
		}
		if !duplicateFound {
			break
		}
		usedIndexes = make(map[int]bool) // Clear usedIndexes for the next iteration
	}

	highestIndex := 0
	//If the scene group indices map is empty, then we need to set the highest index to 0 and skip the next step
	if len(project.SceneGroupIndices) == 0 {
	} else {
		for _, index := range project.SceneGroupIndices {
			if index > highestIndex {
				highestIndex = index
			}
		}
	}
	// Step 3: Close any gaps in the indices
	expectedIndex := 1 // Start from 1
	for {
		if expectedIndex > highestIndex {
			break
		}
		// Check if expectedIndex is not used
		if !usedIndexes[expectedIndex] {
			// Decrement all indices that are greater than the expected index
			for group, index := range project.SceneGroupIndices {
				if index > expectedIndex {
					project.SceneGroupIndices[group] = index - 1
				}
			}
			// Update usedIndexes map to reflect the new state
			usedIndexes = make(map[int]bool)
			for _, index := range project.SceneGroupIndices {
				usedIndexes[index] = true
			}
		}
		// Increment expectedIndex for the next iteration
		expectedIndex++
	}

	// Step 4: Assign indices to scene groups that do not have one
	for group := range project.SceneGroups {
		if _, exists := project.SceneGroupIndices[group]; !exists {
			highestIndex++
			project.SceneGroupIndices[group] = highestIndex
		}
	}

	return project
}
