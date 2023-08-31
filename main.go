package main

import (
	"NovellaForge/pkg/utils/editor"
	"bytes"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
)

/*
Todo:
	- Implement Create Project logic
	  	- Should check if the user has GO installed and tell them to install it if they don't (Will need to add max and min version checks)
		- Should prompt user for project name
	  	- Should create a project subdirectory in the projects directory based on the name they give it
	  		- Should create a project.go file with a default project they can begin working on as well as the predefined go mod file (Named according to the name they give the project)
	  		- Will also create a project.novel file with the project metadata including scenes Game name, game version, author, credits, etc. (Also named according to the name they give the project)
			- Will also create an assets folder a scenes folder and a characters folder (Need to define basic character structure and scene structure) inside the project subdirectory
Todo IDEAS:
	- Engine should eventually support mac os and linux as well as windows
	- Scenes should have an enum to switch between layout styles, and then a list of content objects that will be displayed in the scene according to the slots available in that layout style
	- The scenes will be opened using a populateScene function that takes in a Scene and args and just sets the game field contents to the scene contents
	- A game window will have a main menu at the top with options to save, load, view credits, set preferences, exit, etc.
	- On game launch there will be a splash screen with the game name and version and a button to start the game or load a save as well as a button to view credits
	- in game f1 will quick save to the oldest or first empty save slot and f2 will quick load from the newest save slot pressing f5 through f8 will save to those slots and f9 through f12 will load from those slots (Max of 4 quick saves and 4 quick loads)
	- The game will have a settings menu that will allow the user to change the resolution, toggle fullscreen, change the volume, and change the keybinds
	- The game will have a credits menu that will display the credits from the project.novel file
	- The game will allow normal saves that will be stored in the game directory in a saves folder there will be no limit to the number of saves the user can make they will be named save1.novella save2.novella etc. The user can name their saves whatever they want (Saves will just be json arrays anyway)
	- The game will have a load menu that will display all saves in the saves folder and allow the user to load them




*/

const EditorVersion = "0.0.1"

var Config editor.Config
var LastProject string

// main is the entry point for the application
func main() {

	// Check if 'projects' directory exists
	if _, err := os.Stat("projects"); os.IsNotExist(err) {
		// Create directory
		err = os.Mkdir("projects", 0755)
		if err != nil {
			return
		}
	}

	application := app.New()
	window := application.NewWindow(fmt.Sprintf("NovellaForge Editor %s", EditorVersion))

	err := createEditorWindow(window)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	window.ShowAndRun()
}

// createEditorWindow creates the main editor window
func createEditorWindow(w fyne.Window) error {
	// Initialize Config here
	var err error
	Config, err = editor.LoadConfig(EditorVersion)
	if err != nil {
		return err
	}
	if Config.LastProject != "" {
		LastProject = Config.LastProject
	}

	projectList := scanProjects()
	var projectItems []*fyne.MenuItem
	for _, project := range projectList {
		projectItems = append(projectItems, fyne.NewMenuItem(project, func() {
			// TODO: Implement opening of this project
		}))
	}

	// Main Menu
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			&fyne.MenuItem{
				Label:     "Open Recent...",
				ChildMenu: fyne.NewMenu("Recent Projects", projectItems...),
				Action: func() {
					projectList = scanProjects()
				},
			},
			fyne.NewMenuItem("Open Project", func() {
				showOpenProjectDialog(w)
			}),
			fyne.NewMenuItem("Create Project", func() {
				createProject(w)
			}),
			fyne.NewMenuItem("Editor Settings", func() {
				showSettings(w)
			}),
		),
	)
	w.SetMainMenu(mainMenu)

	// Content
	createProjectButton := widget.NewButton("Create Project", func() {
		createProject(w)
	})
	openProjectButton := widget.NewButton("Open Project", func() {
		showOpenProjectDialog(w)
	})
	continueProjectButton := widget.NewButton("Continue Project", func() {
		// TODO: Implement Continue Project logic
	})
	if LastProject == "" {
		continueProjectButton.Disable()
	}

	w.SetContent(container.NewVBox(
		createProjectButton,
		openProjectButton,
		continueProjectButton,
	))

	if Config.Fullscreen {
		w.SetFullScreen(true)
	} else {
		w.SetFullScreen(false)
		//Validate Width and Height before setting
		if Config.Resolution["width"] < 640 {
			Config.Resolution["width"] = 640
		}
		if Config.Resolution["height"] < 480 {
			Config.Resolution["height"] = 480
		}

		w.Resize(fyne.NewSize(Config.Resolution["width"], Config.Resolution["height"]))
	}

	return nil
}

// showOpenProjectDialog opens up a small popup window for opening a project
func showOpenProjectDialog(w fyne.Window) {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		//TODO Implement opening of project
	}, w)
	//Set the location to the local projects folder
	listableURI, err := storage.ListerForURI(storage.NewFileURI("projects"))
	if err == nil {
		fd.SetLocation(listableURI)
	}
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".novel"}))
	fd.Show()
}

// showSettings opens up a small popup window for editor settings
func showSettings(w fyne.Window) {
	var shouldUpdateSize bool
	updateSizeCheck := widget.NewCheck("Make the Window this size by default", func(b bool) {
		shouldUpdateSize = b
	})

	fullScreenCheck := widget.NewCheck("Fullscreen", func(b bool) {
		Config.Fullscreen = b
	})
	fullScreenCheck.SetChecked(Config.Fullscreen)

	// Convert form items to fyne.Container
	formContainer := container.NewVBox(
		container.NewHBox(widget.NewLabel(""), updateSizeCheck),
		container.NewHBox(widget.NewLabel(""), fullScreenCheck),
	)

	// Use dialog.ShowCustomConfirm with a callback
	dialog.ShowCustomConfirm("Editor Settings", "Save & Apply", "Cancel", formContainer, func(b bool) {
		if b {
			if shouldUpdateSize {
				currentSize := w.Canvas().Size()
				Config.Resolution["width"] = currentSize.Width
				Config.Resolution["height"] = currentSize.Height
			}

			err := editor.SaveConfig(Config)
			if err != nil {
				dialog.ShowError(err, w)
			}

			// Recreate editor window
			err = createEditorWindow(w)
			if err != nil {
				dialog.ShowError(err, w)
			}
		}
	}, w)
}

// scanProjects returns a list of the last 10 edited projects
func scanProjects() []string {
	var projects []string
	files, err := os.ReadDir("projects")
	if err != nil {
		return projects
	}

	// Filter and sort by modification time
	var fileInfo []fs.DirEntry
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".novel" {
			fileInfo = append(fileInfo, file)
		}
	}
	sort.Slice(fileInfo, func(i, j int) bool {
		infoI, _ := fileInfo[i].Info()
		infoJ, _ := fileInfo[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	// Limit to last 10 edited projects
	limit := 10
	if len(fileInfo) < 10 {
		limit = len(fileInfo)
	}

	for _, file := range fileInfo[:limit] {
		projects = append(projects, file.Name())
	}

	return projects
}

func createProject(w fyne.Window) {
	// Check Go version
	isValid, message := checkGoVersion("1.16", "1.18")
	if !isValid {
		dialog.ShowInformation("Go Version Check Failed", message, w)
		return
	}

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
		// TODO: Create project logic here
	}, w)
}

func checkGoVersion(minVersion string, maxVersion string) (bool, string) {
	cmd := exec.Command("go", "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		return false, "Go is not installed. Please install it to proceed."
	}

	// Sample go version output: "go version go1.16.3 darwin/amd64"
	// We'll use regular expression to extract the version number
	re := regexp.MustCompile(`go([0-9]+(\.[0-9]+)+)`)
	matches := re.FindStringSubmatch(out.String())
	if len(matches) < 2 {
		return false, "Could not determine Go version."
	}

	version := matches[1]
	// TODO: Compare version against minVersion and maxVersion

	return true, "Go version: " + version
}
