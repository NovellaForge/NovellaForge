package editor

import (
	"bytes"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

const Version = "0.0.1"
const MinGoVersion = "1.16"
const MaxGoVersion = "1.21"

// CreateMainWindow creates the main editor window
func CreateMainWindow(w fyne.Window) error {
	// Initialize Config here
	var err error
	Config, err = LoadConfig(Version)
	if err != nil {
		return err
	}

	projectList, err := ScanProjects()
	if err != nil {
		// Show an error dialog
		dialog.ShowError(err, w)
	}
	var projectItems []*fyne.MenuItem
	var openRecentMenuItems *fyne.MenuItem
	for _, p := range projectList {
		projectItems = append(projectItems, fyne.NewMenuItem(p, func() {
			err = OpenProject(p, w, false)
			if err != nil {
				return
			}
		}))
	}

	openRecentMenuItems = &fyne.MenuItem{
		Label:     "Open Recent...",
		ChildMenu: fyne.NewMenu("Recent Projects", projectItems...),
	}

	if len(projectItems) == 0 {
		openRecentMenuItems.Disabled = true
	} else {
		openRecentMenuItems.Disabled = false
	}

	// Main Menu
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			openRecentMenuItems,
			fyne.NewMenuItem("Open Project", func() {
				ShowOpenProjectDialog(w)
			}),
			fyne.NewMenuItem("Create Project", func() {
				// Check Go version
				isValid, message := checkGoVersion(MinGoVersion, MaxGoVersion)
				if !isValid {
					dialog.ShowInformation("Go Version Check Failed", message, w)
					return
				} else {
					CreateProject(w)
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
		CreateProject(w)
	})
	openProjectButton := widget.NewButton("Open Project", func() {
		ShowOpenProjectDialog(w)
	})
	continueProjectButton := widget.NewButton("Continue Project", func() {
		//Open the project assuming the last project is a valid path
		err = OpenProject(Config.LastProject, w, true)
		if err != nil {
			//Set the last project to empty string so that the continue button is disabled
			Config.LastProject = ""
			err = SaveConfig(Config)
			if err != nil {
				dialog.ShowError(err, w)
			}
			dialog.ShowError(err, w)
		}
	})

	if Config.LastProject == "" {
		continueProjectButton.Disable()
	} else {
		//Check if the file at the LastProject path exists
		if _, err := os.Stat(Config.LastProject); os.IsNotExist(err) {
			//Set the last project to empty string so that the continue button is disabled
			Config.LastProject = ""
			err = SaveConfig(Config)
			if err != nil {
				dialog.ShowError(err, w)
			}
			continueProjectButton.Disable()
		}
	}

	// Initialize the EditorContent with buttons and welcome message
	Content := fyne.CanvasObject(container.NewVBox(
		widget.NewLabel("Welcome to NovellaForge!"),
		widget.NewLabel("Please select an option below to get started."),
		createProjectButton,
		openProjectButton,
		continueProjectButton,
	))

	if Config.Fullscreen {
		w.SetFullScreen(true)
	} else {
		w.SetFullScreen(false)
		// Validate Width and Height before setting
		if Config.Resolution["width"] < 640 {
			Config.Resolution["width"] = 640
		}
		if Config.Resolution["height"] < 480 {
			Config.Resolution["height"] = 480
		}

		w.Resize(fyne.NewSize(Config.Resolution["width"], Config.Resolution["height"]))
	}

	w.SetContent(Content)

	return nil
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

			err := SaveConfig(Config)
			if err != nil {
				dialog.ShowError(err, w)
			}

			// Recreate editor window
			err = CreateMainWindow(w)
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
	//Truncate the version to 2 decimal places
	version = version[:4]
	minFloat, err := strconv.ParseFloat(minVersion, 32)
	if err != nil {
		return false, "Could not parse minimum Go version."
	}
	maxFloat, err := strconv.ParseFloat(maxVersion, 32)
	if err != nil {
		return false, "Could not parse maximum Go version."
	}
	cur, err := strconv.ParseFloat(version, 32)
	if err != nil {
		return false, "Could not parse current Go version."
	}

	if cur < minFloat {
		//Format a message showing the version and the minimum supported version
		message := "Go version: " + version + " is older than the minimum supported version " + minVersion + ". Please update Go to proceed."
		return false, message
	}
	if cur > maxFloat {
		return true, "Go version is newer than the latest supported version " + maxVersion + " It is untested and may not work."
	}

	return true, "Go version: " + version
}
