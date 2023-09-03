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

type SceneComponent struct {
	Name  string `json:"Name"`
	Type  string `json:"Type"`
	Text  string `json:"Text"`
	Value string `json:"Value"`
}

type Scene struct {
	Name        string           `json:"Name"`
	SceneType   string           `json:"Scene Type"`
	ContentList []SceneComponent `json:"Component List"`
}

type SceneGroup struct {
	Name   string  `json:"Name"`
	Scenes []Scene `json:"Scenes"`
}

type Project struct {
	GameName     string       `json:"Game Name"`
	Version      string       `json:"Version"`
	Author       string       `json:"Author"`
	Credits      string       `json:"Credits"`
	SceneTypes   []string     `json:"Scene Types"`
	ContentTypes []string     `json:"Content Types"`
	SceneGroups  []SceneGroup `json:"Scene Groups"`
}

// CreateDefaultMainMenu creates a default "Main Menu" scene
func CreateDefaultMainMenu() Scene {
	return Scene{
		Name:      "MainMenu",
		SceneType: "Menu",
		ContentList: []SceneComponent{
			{
				Name:  "Title",
				Type:  "Label",
				Text:  "Welcome to the Main Menu",
				Value: "",
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
		SceneGroups: []SceneGroup{
			{
				Name:   "Default",
				Scenes: []Scene{mainMenuScene},
			},
		},
		SceneTypes: []string{
			"Menu",
			"Default",
		},
		ContentTypes: []string{
			"Label",
			"Button",
			"Image",
			"Video",
			"Audio",
			"Text",
		},
	}
}

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

	// Create the main content area
	mainContent := container.NewMax()

	// Populate the tree view for this project
	tree := populateProjectTree(w, mainContent)
	if tree == nil {
		return fmt.Errorf("failed to create project tree")
	}

	tree.Refresh()
	tree.OpenAllBranches()

	// Create a split container with the sidebar and main content area
	mainHSplit := container.NewHSplit(tree, mainContent)
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
		"cmd",
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

func populateProjectTree(w fyne.Window, mainContent *fyne.Container) *widget.Tree {
	var tree *widget.Tree
	tree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				return []widget.TreeNodeID{"GameName", "Version", "Author", "Credits", "SceneTypes", "SceneGroups"}
			}

			if id == "SceneGroups" {
				var groupNames []string
				for _, group := range ActiveProject.SceneGroups {
					groupNames = append(groupNames, group.Name)
				}
				return groupNames
			}

			for _, group := range ActiveProject.SceneGroups {
				if group.Name == id {
					var sceneIDs []string
					for _, scene := range group.Scenes {
						uniqueID := group.Name + "_" + scene.Name // Concatenating the group name and the scene name
						sceneIDs = append(sceneIDs, uniqueID)
					}
					return sceneIDs
				}
			}

			return []widget.TreeNodeID{}
		},
		func(id widget.TreeNodeID) bool {
			if id == "" || id == "SceneGroups" {
				return true
			}
			for _, group := range ActiveProject.SceneGroups {
				if group.Name == id {
					return true
				}
			}
			return false
		},
		func(branch bool) fyne.CanvasObject {
			box := container.NewHBox()
			if branch {
				box.Add(widget.NewLabel("Branch template"))
			} else {
				box.Add(widget.NewLabel("Leaf template"))
			}
			return box
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
			o.(*fyne.Container).Objects[0].(*widget.Label).SetText(displayText)

			//Check if the id is a scene
			for _, group := range ActiveProject.SceneGroups {
				for _, scene := range group.Scenes {
					if id == group.Name+"_"+scene.Name {
						//Check the length of the object's objects
						if len(o.(*fyne.Container).Objects) == 1 {
							//Add an edit button
							editButton := widget.NewButton("Edit", func() {
								//Open the scene editor tab for this scene
								mainContent.Objects = nil
								mainContent.Add(SceneEditor(w, id))

								mainContent.Refresh()
							})
							o.(*fyne.Container).Add(editButton)
						}
					}
				}
			}
		},
	)

	tree.OnSelected = func(id string) {
		if id == "SceneGroups" {
			nameEntry := widget.NewEntry()
			nameEntry.SetPlaceHolder(id)
			sceneGroupNames := []string{"None"}
			for _, group := range ActiveProject.SceneGroups {
				sceneGroupNames = append(sceneGroupNames, group.Name)
			}
			deleteGroupDropDown := widget.NewSelect(sceneGroupNames, func(s string) {})
			deleteGroupDropDown.SetSelected(sceneGroupNames[0])

			//Create the form for the dialog
			form := &widget.Form{
				Items: []*widget.FormItem{
					{Text: "New Scene Group:", Widget: nameEntry},
					{Text: "Delete Scene Group:", Widget: deleteGroupDropDown},
				},
			}

			dialog.ShowCustomConfirm("Edit Groups", "Apply", "Cancel", form, func(b bool) {
				tree.UnselectAll()
				if !b {
					return
				}
				groupName := strings.ReplaceAll(nameEntry.Text, "_", " ")
				if groupName == "" && deleteGroupDropDown.Selected == "None" {
					return
				} else if groupName == "" {

					//Pop up a confirmation dialog to make sure the user wants to delete the selected scene group
					dialog.ShowConfirm("Delete Scene Group", "Are you sure you want to delete this scene group?\nDoing so will also delete all scenes within it", func(b bool) {
						if !b {
							return
						}
						//Delete the selected scene group
						for i, group := range ActiveProject.SceneGroups {
							if group.Name == deleteGroupDropDown.Selected {
								ActiveProject.SceneGroups = append(ActiveProject.SceneGroups[:i], ActiveProject.SceneGroups[i+1:]...)
							}
						}
						err := SaveProject()
						if err != nil {
							dialog.ShowError(err, w)
							return
						}
						tree.Refresh()
					}, w)
				} else {
					//Make sure the scene group name does not already exist using the new name field
					for _, group := range ActiveProject.SceneGroups {
						if group.Name == groupName {
							dialog.ShowInformation("Invalid Name", "Scene Group already exists", w)
							return
						}
					}
					//Append the new scene group to the project
					newSceneGroup := SceneGroup{
						Name:   groupName,
						Scenes: []Scene{},
					}
					ActiveProject.SceneGroups = append(ActiveProject.SceneGroups, newSceneGroup)
					err := SaveProject()
					if err != nil {
						dialog.ShowError(err, w)
						return
					}
					tree.Refresh()
				}
			}, w)
		}
		for _, group := range ActiveProject.SceneGroups {
			if group.Name == id {
				// This is a Scene Group; open dialog for editing or deleting
				var newSceneButton *widget.Button
				var upButton *widget.Button
				var downButton *widget.Button
				//Get the current index of the scene group
				var index int
				for i, g := range ActiveProject.SceneGroups {
					if g.Name == id {
						index = i
					}
				}

				//Create the up button
				upButton = widget.NewButton("Up", func() {
					//Shift the current scene group up in the slice
					for i, g := range ActiveProject.SceneGroups {
						if g.Name == id {
							if i == 0 {
								return
							}
							ActiveProject.SceneGroups[i], ActiveProject.SceneGroups[i-1] = ActiveProject.SceneGroups[i-1], ActiveProject.SceneGroups[i]
							err := SaveProject()
							if err != nil {
								dialog.ShowError(err, w)
								return
							}
							tree.Refresh()
							return
						}
					}
				})

				//Create the down button
				downButton = widget.NewButton("Down", func() {
					//Shift the current scene group down in the slice
					for i, g := range ActiveProject.SceneGroups {
						if g.Name == id {
							if i == len(ActiveProject.SceneGroups)-1 {
								return
							}
							ActiveProject.SceneGroups[i], ActiveProject.SceneGroups[i+1] = ActiveProject.SceneGroups[i+1], ActiveProject.SceneGroups[i]
							err := SaveProject()
							if err != nil {
								dialog.ShowError(err, w)
								return
							}
							tree.Refresh()
							return
						}
					}
				})

				//Create the new scene button
				newSceneButton = widget.NewButton("New Scene", func() {
					//Pops up a new dialog to create a new scene
					nameEntry := widget.NewEntry()
					nameEntry.SetPlaceHolder("Scene Name")
					typeDropDown := widget.NewSelect(ActiveProject.SceneTypes, func(s string) {})
					typeDropDown.SetSelected(ActiveProject.SceneTypes[0])
					dialog.ShowCustomConfirm("New Scene", "Add", "Cancel", container.NewVBox(nameEntry, typeDropDown), func(b bool) {
						tree.UnselectAll()
						if !b {
							return
						}
						sceneName := strings.ReplaceAll(nameEntry.Text, "_", " ")
						if sceneName == "" {
							dialog.ShowInformation("Invalid Name", "Scene name cannot be empty", w)
							return
						}

						//Make sure the scene name does not already exist
						for _, scene := range group.Scenes {
							if scene.Name == sceneName {
								dialog.ShowInformation("Invalid Name", "Scene already exists", w)
								return
							}
						}

						//Add the new scene to the scene group
						newScene := Scene{
							Name:      sceneName,
							SceneType: typeDropDown.Selected,
						}

						//Add the scene to the active project
						group.Scenes = append(group.Scenes, newScene)
						//replace the scene group in the active project
						for i, g := range ActiveProject.SceneGroups {
							if g.Name == group.Name {
								ActiveProject.SceneGroups[i] = group
							}
						}

						err := SaveProject()
						if err != nil {
							dialog.ShowError(err, w)
							return
						}

						tree.Refresh()

					}, w)

				})

				// Create an entry field for renaming the Scene Group
				renameEntry := widget.NewEntry()
				renameEntry.SetPlaceHolder("New Name")

				// Create a form for the dialog
				form := &widget.Form{
					Items: []*widget.FormItem{
						{Text: "Rename:", Widget: renameEntry},
						{Text: "", Widget: newSceneButton},
						{Text: "Shift:", Widget: upButton},
						{Text: "", Widget: downButton},
					},
				}

				// Show the dialog with the custom form
				dialog.ShowCustomConfirm("Edit Scene Group", "Apply Changes", "Cancel", form, func(b bool) {
					tree.UnselectAll()
					if !b {
						//If the index is not the same as the original index, then we need to shift the scene group back to its original index
						var newIndex int
						for i, g := range ActiveProject.SceneGroups {
							if g.Name == id {
								newIndex = i
							}
						}
						if newIndex != index {
							//Check if it is above or below the original index
							if newIndex < index {
								//Shift the scene group down in the slice one index at a time until it is back at its original index
								for i := newIndex; i < index; i++ {
									ActiveProject.SceneGroups[i], ActiveProject.SceneGroups[i+1] = ActiveProject.SceneGroups[i+1], ActiveProject.SceneGroups[i]
								}
							} else {
								//Shift the scene group up in the slice one index at a time until it is back at its original index
								for i := newIndex; i > index; i-- {
									ActiveProject.SceneGroups[i], ActiveProject.SceneGroups[i-1] = ActiveProject.SceneGroups[i-1], ActiveProject.SceneGroups[i]
								}
							}
						}
						//Save and refresh the tree
						err := SaveProject()
						if err != nil {
							dialog.ShowError(err, w)
							return
						}
						tree.Refresh()
						return
					}

					// Rename the Scene Group
					newName := renameEntry.Text
					if newName != "" {
						//Make sure the scene group name does not already exist
						for _, g := range ActiveProject.SceneGroups {
							if g.Name == newName {
								dialog.ShowInformation("Invalid Name", "Scene Group already exists", w)
								return
							}
						}
						//Rename the scene group
						group.Name = newName
						//replace the scene group in the active project
						for i, g := range ActiveProject.SceneGroups {
							if g.Name == id {
								ActiveProject.SceneGroups[i] = group
							}
						}
					}

					// Save the project
					err := SaveProject()
					if err != nil {
						dialog.ShowError(err, w)
						return
					}

					// Refresh the tree
					tree.Refresh()

				}, w)
			}
		}

		if strings.Contains(id, "_") {
			// This is a Scene; open dialog for editing, renaming or deleting
			nameEntry := widget.NewEntry()
			nameEntry.SetPlaceHolder("Name")
			typeDropDown := widget.NewSelect(ActiveProject.SceneTypes, func(s string) {})
			currentType := ActiveProject.SceneTypes[0]
			for _, group := range ActiveProject.SceneGroups {
				for _, scene := range group.Scenes {
					if group.Name+"_"+scene.Name == id {
						currentType = scene.SceneType
					}
				}
			}
			typeDropDown.SetSelected(currentType)

			//Create the form for the dialog
			form := &widget.Form{
				Items: []*widget.FormItem{
					{Text: "Rename:", Widget: nameEntry},
					{Text: "Type:", Widget: typeDropDown},
				},
			}

			dialog.ShowCustomConfirm("Edit Scene", "Apply", "Cancel", form, func(b bool) {
				if !b {
					return
				}
				sceneName := strings.ReplaceAll(nameEntry.Text, "_", " ")
				//if the scene name is blank and the type is the same as the current type, then we don't need to do anything
				if sceneName == "" && typeDropDown.Selected == currentType {
					return
				} else if sceneName == "" {
					//Update the scene type asking for confirmation first
					dialog.ShowConfirm("Update Scene Type", "Are you sure you want to update the scene type?\nDoing so may break content in the scene", func(b bool) {
						if !b {
							return
						}
						//Update the scene type
						for _, group := range ActiveProject.SceneGroups {
							for i, scene := range group.Scenes {
								if group.Name+"_"+scene.Name == id {
									group.Scenes[i].SceneType = typeDropDown.Selected
								}
							}
						}
					}, w)
				} else {
					//Make sure the scene name does not already exist
					for _, group := range ActiveProject.SceneGroups {
						for _, scene := range group.Scenes {
							if group.Name+"_"+scene.Name == id {
								continue
							}
							if scene.Name == sceneName {
								dialog.ShowInformation("Invalid Name", "Scene already exists", w)
								return
							}
						}
					}
					//Update the scene name
					for _, group := range ActiveProject.SceneGroups {
						for i, scene := range group.Scenes {
							if group.Name+"_"+scene.Name == id {
								group.Scenes[i].Name = sceneName
							}
						}
					}

					//Check if the scene type is the same as the current type
					if typeDropDown.Selected != currentType {
						//Update the scene type asking for confirmation first
						dialog.ShowConfirm("Update Scene Type", "Are you sure you want to update the scene type?\nDoing so may break content in the scene", func(b bool) {
							if !b {
								return
							}
							//Update the scene type
							for _, group := range ActiveProject.SceneGroups {
								for i, scene := range group.Scenes {
									if group.Name+"_"+scene.Name == id {
										group.Scenes[i].SceneType = typeDropDown.Selected
									}
								}
							}
						}, w)
					}

				}
				err := SaveProject()
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				tree.Refresh()
			}, w)
		}
		if id == "GameName" || id == "Version" || id == "Author" {
			// Open the dialog
			valueEntry := widget.NewEntry()
			valueEntry.SetPlaceHolder(fmt.Sprintf("Edit %s", id))

			dialog.ShowCustomConfirm(fmt.Sprintf("Edit %s", id), "Update", "Cancel", valueEntry, func(b bool) {
				tree.UnselectAll()
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
	// Loop through all group names making sure there are no duplicates
	var validatedSceneGroups []SceneGroup
	for _, group := range project.SceneGroups {
		// Make sure the group name is not empty
		if group.Name == "" {
			group.Name = "Default"
		}
		// Make sure the group name is unique
		for _, validatedGroup := range validatedSceneGroups {
			if group.Name == validatedGroup.Name {
				// The group name is not unique, so we need to rename it
				group.Name = group.Name + " (Duplicate)"
			}
		}

		// Loop through all scene names making sure there are no duplicates
		var validatedScenes []Scene
		for _, scene := range group.Scenes {
			// Make sure the scene name is not empty
			if scene.Name == "" {
				scene.Name = "Default"
			}
			// Make sure the scene name is unique
			for _, validatedScene := range validatedScenes {
				if scene.Name == validatedScene.Name {
					// The scene name is not unique, so we need to rename it
					scene.Name = scene.Name + " (Duplicate)"
				}
			}
			validatedScenes = append(validatedScenes, scene)
		}
		group.Scenes = validatedScenes
		validatedSceneGroups = append(validatedSceneGroups, group)
	}
	// Update the project with the validated scene groups
	project.SceneGroups = validatedSceneGroups
	return project
}

// SceneEditor creates a new Scene Editor interface
func SceneEditor(w fyne.Window, SceneID string) fyne.CanvasObject {
	//get the selected scene from the active project
	var selectedScene Scene
	for _, group := range ActiveProject.SceneGroups {
		for _, scene := range group.Scenes {
			if group.Name+"_"+scene.Name == SceneID {
				selectedScene = scene
			}
		}
	}
	//If the selected scene is empty, then we need to return an empty container
	if selectedScene.Name == "" {
		return container.NewVBox(widget.NewLabel("Invalid Scene"))
	}

	// Placeholder for the list of elements
	elementList := widget.NewList(
		func() int {
			return len(selectedScene.ContentList)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("element")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(selectedScene.ContentList[i].Name)
		},
	)

	// Placeholder for properties editor
	nameEntry := widget.NewEntry()
	//TypeEntry Dropdown Menu
	typeEntry := widget.NewSelect(ActiveProject.ContentTypes, func(s string) {})

	//If the selected type is not in the content types, then we need to pop up a warning dialog and set the type to the first type in the list
	var typeFound bool
	for _, t := range ActiveProject.ContentTypes {
		if t == selectedScene.ContentList[0].Type {
			typeFound = true
		}
	}
	if !typeFound {
		//Say that the type for this element is invalid and that it is being reset to the first type in the list
		dialog.ShowInformation("Invalid Type", "The type for this element is invalid and is being reset to the first type in the list, but it will not be saved until you click save", w)
		typeEntry.SetSelected(ActiveProject.ContentTypes[0])
	}

	textEntry := widget.NewEntry()
	valueEntry := widget.NewEntry()
	valueEntry.MultiLine = true

	nameEntry.SetPlaceHolder("Name")
	typeEntry.SetSelected(ActiveProject.ContentTypes[0])
	textEntry.SetPlaceHolder("Text")
	valueEntry.SetPlaceHolder("Value")

	// Update property editor when an element is selected
	var selectedItem widget.ListItemID
	//Default everything to the first element
	nameEntry.SetText(selectedScene.ContentList[0].Name)
	typeEntry.SetSelected(selectedScene.ContentList[0].Type)
	textEntry.SetText(selectedScene.ContentList[0].Text)
	valueEntry.SetText(selectedScene.ContentList[0].Value)
	selectedItem = 0

	// To keep track of the selected item
	addButton := widget.NewButton("Add", func() {
		// Create a new element with default values
		newElement := SceneComponent{
			Name: "New Element",
			Type: "Label",
			Text: "Default Text",
			// TODO: Add default values for other fields
		}

		// Append the new element to selectedScene.ContentList
		selectedScene.ContentList = append(selectedScene.ContentList, newElement)

		// Refresh the list to display the new element
		elementList.Refresh()
	})

	removeButton := widget.NewButton("Remove", func() {
		if selectedItem < 0 || selectedItem >= len(selectedScene.ContentList) {
			// Invalid selection
			return
		}

		// Remove the selected element from selectedScene.ContentList
		selectedScene.ContentList = append(selectedScene.ContentList[:selectedItem], selectedScene.ContentList[selectedItem+1:]...)

		// Refresh the list to update the display
		elementList.Refresh()
	})

	// Update property editor when an element is selected
	elementList.OnSelected = func(id widget.ListItemID) {
		selectedItem = id // Update the selected item
		nameEntry.SetText(selectedScene.ContentList[id].Name)
		typeEntry.SetSelected(selectedScene.ContentList[id].Type)
		textEntry.SetText(selectedScene.ContentList[id].Text)
		valueEntry.SetText(selectedScene.ContentList[id].Value)
	}

	saveButton := widget.NewButton("Save", func() {
		// Update the selected element with the new values
		selectedScene.ContentList[selectedItem].Name = nameEntry.Text
		selectedScene.ContentList[selectedItem].Type = typeEntry.Selected
		selectedScene.ContentList[selectedItem].Text = textEntry.Text
		selectedScene.ContentList[selectedItem].Value = valueEntry.Text
		// Refresh the list to update the display
		elementList.Refresh()

		//Save the project
		err := SaveProject()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
	})

	// Combine everything into a layout
	box :=
		container.NewVBox(
			container.NewHBox(addButton, removeButton, saveButton),
			container.NewGridWithColumns(2,
				widget.NewLabel("Name:"), nameEntry,
				widget.NewLabel("Type:"), typeEntry,
				widget.NewLabel("Text:"), textEntry,
			),
			valueEntry,
		)

	editorLayout := container.NewHSplit(
		box,
		elementList)

	//make the split 25 75
	editorLayout.Offset = 0.75

	return editorLayout
}
