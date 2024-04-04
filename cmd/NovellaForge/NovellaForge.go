package main

import (
	"go.novellaforge.dev/novellaforge/assets/icons"
	"go.novellaforge.dev/novellaforge/pkg/NFFunction"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
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
	"go.novellaforge.dev/novellaforge/internal/NFEditor"
	"go.novellaforge.dev/novellaforge/pkg/NFLog"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget/CalsWidgets"

	//Profiler stuff

	_ "net/http/pprof"
)

/*
//TODO Fix up scene loading possibly using a jsonSafe struct


TODO:
 [ ] Finish the parsing refactor to use the new interface system. This should also include a new remaster of the export system so that it just lists the types of the required and optional args
 [ ] Scenes should include an argument interface that gets passed to all sub widgets and containers as a reference, thus allowing variables to be shared between widgets and containers
 [ ] Finish the global variable refactor to use the new interface system and allow for better save control
 [ ] Finish documentation/comments
 [ ] Finish the scene editor
 	[ ] Property manager needs to be able to fully edit all widget and container properties
	[ ] Add in the ability to add and remove widgets and containers and change all their relevant properties
	[ ] Add in the ability to add and remove scenes
	[ ] Add in the ability to change the order of scenes
	[ ] Add in the ability to change the name of scenes
	[ ] Add in the ability to group widgets and containers in scenes and group scenes in a project
	[ ] Add in the ability to change the order of widgets and containers
 [ ] Add in a way to run the game from the editor for testing
 [ ] Add in the preview run mode (with the ability to edit text fields in the preview)
 [ ] Add in the ability to open the project in the default IDE
 [ ] Add in the ability to open the project in the default file manager
 [ ] Add in the ability to open the project in the default terminal
 [ ] Add in the ability to change what IDE is used to open the project
 [ ] Add in the build manager (for building the game on all platforms)
 [ ] Add in the ability to generate a keystore for android builds
---- above this line will be version 0.1.0 and will be the stability requirement for 1.0.0
---- this means that if all features exist and can work in a basic fashion we are in 0.1.0
---- if all features exist in a fully fleshed out and stable fashion we are in 1.0.0
---- THIS LIST IS STILL HEAVILY IN FLUX AND WILL CHANGE AS PROJECT GOALS EVOLVE
TODO Future:
 [ ] Add in drag and drop widget support
 [ ] Add in IOS and Mac support for game builds and Mac support for editor functionality
*/

const (
	Version = "0.0.1"
	Author  = "The Novella Forge Team"
)

var WindowTitle = "Novella Forge" + " " + Version

func init() {
	// Set the imported functions to be exported
	NFFunction.ShouldExport = true
	NFFunction.ExportPath = "assets/functions"
	NFWidget.ShouldExport = true
	NFWidget.ExportPath = "assets/widgets"
	NFLayout.ShouldExport = true
	NFLayout.ExportPath = "assets/layouts"
}

func main() {
	// Start the profiler (located at localhost:6060/debug/pprof/)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Create a new application and window with the title based on the version
	application := app.NewWithID("com.novellaforge.editor")

	//Load the embedded icons/EditorPng as bytes
	iconBytes, err := icons.EditorPng.ReadFile("editor.png")
	if err != nil {
		//If the icon fails to load, log the error and set the icon to the default application icon
		log.Printf("Failed to load icon: %v", err)
		application.SetIcon(theme.FileApplicationIcon())
	} else {
		//If the icon loads successfully, set the icon to the loaded icon after converting it to a StaticResource
		iconResource := fyne.NewStaticResource("editor.png", iconBytes)
		application.SetIcon(iconResource)
	}

	window := application.NewWindow(WindowTitle)

	// Use common 720p resolution for base window size
	window.Resize(fyne.NewSize(1280, 720))

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
	splashContent := container.NewVBox(
		widget.NewLabelWithStyle("NovellaForge", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Version: "+Version, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		widget.NewLabelWithStyle("Developed By: "+Author, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		widget.NewLabelWithStyle("Powered By: Fyne", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		loading.Box,
	)

	// If the user is on a desktop, show a splash screen while the main content is loading
	if drv, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok {
		splash = drv.CreateSplashWindow()
		splash.SetContent(splashContent)
	} else {
		// If the user is not on a desktop, show the splash screen in the main window
		window.SetContent(splashContent)
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

	// Show the splash screen if it was created, otherwise show the main window
	if splash != nil {
		splash.Show()
		// Create the main NovellaForge content in a thread, which will also update the loading bar
		go CreateMainContent(window, loading)
		application.Run()
	} else {
		go CreateMainContent(window, loading)
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
	// Create an Open Project button in the top right
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
	loading.Complete()
}
