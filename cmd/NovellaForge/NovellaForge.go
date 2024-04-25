package main

import (
	"go.novellaforge.dev/novellaforge/assets"
	"go.novellaforge.dev/novellaforge/pkg/NFFunction"
	"go.novellaforge.dev/novellaforge/pkg/NFFunction/DefaultFunctions"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout/DefaultLayouts"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget/DefaultWidgets"
	"golang.org/x/sys/windows"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/internal/NFEditor"
	"go.novellaforge.dev/novellaforge/pkg/NFLog"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget/CalsWidgets"

	//Profiler stuff

	_ "net/http/pprof"
)

/*


TODO 0.0.1: [ ] = Required, ( ) = Optional, X = Done, - = Alternative implementation
 [X] Finish the parsing refactor to use the new interface system.
 [ ] Finish documentation/comments
 [ ] Video And Audio players need to be added to the game templates
 	[ ] Make sure ffmpeg binaries are added to the game folder and are unpacked into the config folder or wherever the game creator specifies
 	[ ] This includes allowing embedding of the binaries in the game via a build option and making sure the path to the binaries is set properly
 [ ] Finish the scene editor
	[X] Add in the ability to add and remove scenes
	[X] Add in the ability to group scenes
	(-) Add in the ability to change the order of scenes (I sorted them alphabetically for now and added the ability to group)
 	[X] Property manager needs to be able to fully edit all widget and container properties
	[ ] Add in the ability to add and remove widgets and containers and change all their relevant properties
	[X] Add in the ability to change the name of scenes
	( ) Possibly add in the ability to group widgets and layouts
	( ) Possibly Add in the ability to change the order of widgets and layouts
 [ ] Refactor NFSave to use the new interface system and integrate it with default game templates
 [ ] Add in the build manager
 	[ ] Include embedding toggle to embed assets in the binary
TODO 0.1.0:
 [ ] Add in a way to run the game from the editor for testing
 [ ] Scene Editor Preview Mode
	[ ] This should add clickable elements to all elements that select them for editing in the properties
    [ ] This should override buttons and other interactive elements to not be clickable
TODO 1.0.0:
 [ ] Add in the debug run mode
   [ ] This should run the game with a debug flag enabled, that enables editing in certain widget that support it
 [ ] Add in the ability to open the project in the default IDE
 [ ] Add in the ability to open the project in the default file manager
 [ ] Add in the ability to open the project in the default terminal
 [ ] Add in the ability to change what IDE is used to open the project
 [ ] Add in the ability to generate a keystore for android builds
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
	DefaultWidgets.Import()
	DefaultLayouts.Import()
	DefaultFunctions.Import()
	// Set the path the imported functions to be exported
	NFFunction.ExportPath = "assets/functions"
	NFWidget.ExportPath = "assets/widgets"
	NFLayout.ExportPath = "assets/layouts"

	// Export the registered functions, widgets, and layouts
	/*
		NFLayout.ExportRegistered()
		NFWidget.ExportRegistered()
		NFFunction.ExportRegistered()
	*/

}

func main() {
	// Start the profiler (located at localhost:6060/debug/pprof/)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Create a new application and window with the title based on the version
	application := app.NewWithID("com.novellaforge.editor")

	//Load the embedded icons/EditorPng as bytes
	iconBytes, err := assets.EditorPng.ReadFile("editor.png")
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
		go NFEditor.CreateMainContent(window, loading)
		application.Run()
	} else {
		go NFEditor.CreateMainContent(window, loading)
		window.ShowAndRun()
	}

	switch runtime.GOOS {
	case "windows":
		//Make sure to reset the TimePeriod for the system timer
		err = windows.TimeEndPeriod(1)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}
}
