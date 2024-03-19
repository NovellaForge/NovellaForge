package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/internal/NFEditor"
	"github.com/NovellaForge/NovellaForge/pkg/NFLog"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget/CalsWidgets"

	//Profiler stuff

	_ "net/http/pprof"
)

/*
TODO Priorities:
	- Update to use the new splash screen format
	- Update the templates to use the actual go template functionality
	- Update the project creation to use the new templates
		- Update the project creation to use the built in os perm modes
	- Refactor all the parsing



TODO New Editor Requirements:
	- Parsers:
		- Potentially switch all handlers over to ...interface{} and then use reflection to determine the type of the interface and parse it instead of using a map[string]interface{}
	- Scene Editor:
		- The scene editor should be a window that takes up the right side of the main window
		- The scene editor should have a tree view on the left side that shows all the objects in the scene and allows the user to select them and edit their values containers should be able to be expanded and collapsed
		- Objects should be displayed as their type and name in the tree view like type: name
		- At the top of the object tree view there should be a search bar that allows the user to search for a specific object by name
		- Underneath the search bar there should be a button that allows the user to add a new object to the scene that will pop up a dialog with a drop down menu that is populated with all the available object types from the default widgets and containers and the custom ones and a text box that allows the user to name the object
		- When the user clicks the add button the object should be added to the scene and the object tree view and the user should be able to edit its values
		- Values are populated into the object editor based on the type of object selected custom objects will pull the types from the json data and populate the editor with the correct type fields
		- When the user clicks an object in the tree view the object editor should be populated with the objects values and the user should be able to edit them
		- When the user clicks the save button the object editor should save the values back to the object and the object tree view should update to reflect the changes
		- Above the search bar in the main project tree, there should be a build game button that will build the game and save it to the project directory in a build folder
		- When the build button is clicked it should give options for the user to build the game for windows, mac, linux, android, and ios
	- Main Editor:
		- Should store project info in a file that can be loaded and opened to view the project history and info allowing to continue last project and open recent projects
		- Terminal window that displays the output of the game or editor when it is running with an entry box at the bottom that allows the user to enter commands



Todo IDEAS:
	- in game Shift + f1 will quick save to the oldest or first empty save slot and Shift + f2 will quick load from the newest save slot pressing Shift plus f5 through f8 will save to those slots and f9 through f12 will load from those slots (Max of 4 quick saves and 4 quick loads)
	- Open the terminal with Shift f3 and close it with Shift f4 or escape. Shift f3 will open the saves menu
	- Check for GO and automatically install it if it is not there after prompting the user for an automatic install
	- Check for the fyne dependencies and install them if they are not there (fyne install)
*/

const (
	Version = "0.0.1"
	Icon    = "assets/icons/editor.png"
	Author  = "The Novella Forge Team"
)

var WindowTitle = "Novella Forge" + " " + Version

func main() {
	// Start the profiler (located at localhost:6060/debug/pprof/)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Create a new application and window with the title based on the version
	application := app.NewWithID("com.novellaforge.editor")
	window := application.NewWindow(WindowTitle)

	// Use common 720p resolution for base window size
	window.Resize(fyne.NewSize(1280, 720))
	iconResource, err := fyne.LoadResourceFromPath(Icon)
	if err != nil {
		log.Printf("Failed to load icon: %v", err)
		application.SetIcon(theme.FileApplicationIcon())
	} else {
		application.SetIcon(iconResource)
		window.SetIcon(application.Icon())
	}
	userHome, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	err = NFLog.SetUp(window, application.Preferences().StringWithFallback("logDir", userHome+"/NovellaForge/Logs"))
	if err != nil {
		log.Fatal(err)
	}

	// Create a loading widget shown while the main NovellaForge content is loading
	loadingChannel := make(chan struct{})
	loading := CalsWidgets.NewLoading(loadingChannel, 0*time.Second, 100)
	var splash fyne.Window

	// If the user is on a desktop, show a splash screen while the main content is loading
	if drv, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok {
		splash = drv.CreateSplashWindow()
		splash.SetContent(container.NewVBox(
			widget.NewLabelWithStyle("NovellaForge", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Version: "+Version, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabelWithStyle("Developed By: "+Author, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabelWithStyle("Powered By: Fyne", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			loading.Box,
		))
	}

	//Sets the main window to be the master window so that it can be focused and when it is closed the application will close
	window.SetMaster()

	// Once the loading bar is complete, close the splash screen and show the main window
	// This code runs in a thread, so we can continue to load the main content while the splash screen is shown
	// Note that the splash variable contains the loading bar, and once the loading bar is complete, the splash screen is closed
	// The window variable contains the NovellaForge main content, which is shown after the splash screen is closed
	go func() {
		<-loadingChannel
		if splash != nil {
			splash.Close()
			window.Show()
			window.RequestFocus()
		}
	}()

	// Create the main NovellaForge content in a thread, which will also update the loading bar
	go CreateMainContent(window, loading)

	// Show the splash screen if it was created, otherwise show the main window
	if splash != nil {
		splash.Show()
		application.Run()
	} else {
		window.ShowAndRun()
	}
}

// CreateMainContent updates the loading variable as the NovellaForge content is created
func CreateMainContent(window fyne.Window, loading *CalsWidgets.Loading) {
	// Runs "go version" to check if Go is installed
	loading.SetProgress(0, "Checking Dependencies")
	NFEditor.CheckAndInstallDependencies(window)

	// Creates a main menu to hold the buttons below
	loading.SetProgress(10, "Creating Main Menu")
	NFEditor.CreateMainMenu(window)

	// Create a grid layout for the four main buttons
	loading.SetProgress(20, "Creating Main Content")
	grid := container.New(layout.NewGridLayout(2))

	// Create a New Project button in the top left
	newProjectButton := widget.NewButton("New Project", func() {
		NFEditor.NewProjectDialog(window)
	})
	// Create a Open Project button in the top right
	openProjectButton := widget.NewButton("Open Project", func() {
		NFEditor.OpenProjectDialog(window)
	})
	// Create an Open Recent button in the bottom left
	openRecentButton := widget.NewButton("Open Recent", func() {
		NFEditor.OpenRecentDialog(window)
	})
	// Create a Continue Last button in the bottom right
	continueLastButton := widget.NewButton("Continue Last", func() {})
	loading.SetProgress(50, "Checking for Recent Projects")
	projects, err := NFEditor.ReadProjectInfo()
	if err != nil {
		//Show an error dialog
		dialog.ShowError(err, window)
		return
	}
	var project NFEditor.ProjectInfo
	if len(projects) == 0 {
		continueLastButton.Disable()
	} else {
		//Get the most recently opened project
		project = projects[0]
		for i := 0; i < len(projects); i++ {
			if projects[i].OpenDate.After(project.OpenDate) {
				project = projects[i]
			}
		}
	}
	continueLastButton.OnTapped = func() {
		err = NFEditor.OpenFromInfo(project, window)
		if err != nil {
			//Show an error dialog
			dialog.ShowError(err, window)
		}
	}
	loading.SetProgress(95, "Adding Content to Grid")
	//Add the buttons to the grid
	grid.Add(newProjectButton)
	grid.Add(openProjectButton)
	grid.Add(openRecentButton)
	grid.Add(continueLastButton)
	loading.SetProgress(100, "Setting Content")
	window.SetContent(grid)
	time.Sleep(1 * time.Second)
	loading.Complete()
}
