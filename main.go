package main

import (
	"NovellaForge/pkg/project"
	"NovellaForge/pkg/utils/editor"
	"bytes"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"os"
	"os/exec"
	"regexp"
)

/*
Todo:
	- Check for the proper fyne setup and install it if it is not there (fyne setup)
	- Check for the fyne dependencies and install them if they are not there (fyne install)
	- Begin work on the open project functionality
		- It should show a list of all scenes in the project in a tree view and allow the user to open them in a scene editor tab
		- The user can only be in one scene tab at a time and can switch between it and the tree view at any time.
		- The left side of the screen will be a slide out menu that will allow the user to switch between any other project or editor settings
		- It should also allow the user to open the project settings
		- It should also allow the user to open the project credits
		- It should also have a build button with checkboxes for windows, mac, and linux (The build button should be disabled if the project has not been saved and each os will build for all architectures)

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
	- When building check if the author/credits etc are empty and if they are prompt the user to fill them in




*/

const EditorVersion = "0.0.1"
const MinGoVersion = "1.16"
const MaxGoVersion = "1.21"

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

	err := createMainWindow(window)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	window.ShowAndRun()
}

// createMainWindow creates the main editor window
func createMainWindow(w fyne.Window) error {
	// Initialize Config here
	var err error
	Config, err = editor.LoadConfig(EditorVersion)
	if err != nil {
		return err
	}
	if Config.LastProject != "" {
		LastProject = Config.LastProject
	}

	projectList, err := project.ScanProjects()
	if err != nil {
		//Show an error dialog
		dialog.ShowError(err, w)
	}
	var projectItems []*fyne.MenuItem
	for _, p := range projectList {
		projectItems = append(projectItems, fyne.NewMenuItem(p, func() {
			// TODO: Implement opening of this project
		}))
	}

	// Main Menu
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			&fyne.MenuItem{
				Label:     "Open Recent...",
				ChildMenu: fyne.NewMenu("Recent Projects", projectItems...),
			},
			fyne.NewMenuItem("Open Project", func() {
				project.ShowOpenProjectDialog(w)
			}),
			fyne.NewMenuItem("Create Project", func() {
				// Check Go version
				isValid, message := checkGoVersion(MinGoVersion, MaxGoVersion)
				if !isValid {
					dialog.ShowInformation("Go Version Check Failed", message, w)
					return
				} else {
					project.CreateProject(w)
				}
			}),
			fyne.NewMenuItem("Editor Settings", func() {
				showSettings(w)
			}),
		),
	)
	w.SetMainMenu(mainMenu)

	// Content
	createProjectButton := widget.NewButton("Create Project", func() {
		// Check Go version
		isValid, message := checkGoVersion(MinGoVersion, MaxGoVersion)
		if !isValid {
			dialog.ShowInformation("Go Version Check Failed", message, w)
			return
		}
		project.CreateProject(w)
	})
	openProjectButton := widget.NewButton("Open Project", func() {
		project.ShowOpenProjectDialog(w)
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

func createProjectOverviewWindow(w fyne.Window) {

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
			err = createMainWindow(w)
			if err != nil {
				dialog.ShowError(err, w)
			}
		}
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
