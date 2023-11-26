package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/internal/NFEditor"
	"github.com/NovellaForge/NovellaForge/pkg/NFLog"
	"log"
	"net/http"
	"os"
	"time"
)

//Profiler stuff
import _ "net/http/pprof"

/*
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
var window fyne.Window
var application fyne.App
var appReady = false

func init() {
	//Start the profiler
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	application = app.NewWithID("com.novellaforge.editor")
	window = application.NewWindow(WindowTitle)
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
	go SplashScreenLoop(window)
}

func main() {
	NFEditor.CreateMainContent(window)
	appReady = true
	application.Run()
}

func SplashScreenLoop(window fyne.Window) {
	drv := fyne.CurrentApp().Driver()
	for {
		//Check if the driver is ready
		if drv != nil {
			//If it is, break out of the loop
			break
		}
	}

	if drv, ok := drv.(desktop.Driver); ok {
		splash := drv.CreateSplashWindow()
		splashBox := container.NewVBox(
			widget.NewLabelWithStyle("NovellaForge", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Version: "+Version, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabelWithStyle("Developed By: "+Author, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabelWithStyle("Powered By: Fyne", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		)
		splash.SetContent(splashBox)
		splash.Show()
		go func() {
			for !appReady {
			}
			time.Sleep(time.Second * 3)
			splash.Close()
			window.Show()
		}()
	}
}
