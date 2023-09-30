package main

import (
	"NovellaForge/pkg/NFEditor"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"time"
)

/*
TODO New Editor Requirements:
	- Editor should open up and show four main buttons centered in the window, New Project, Open Project, Continue Last Project, and Exit
	- New Project should open a dialog that asks for the project name and the project directory and stores that information and its last open date in some form of persistent editor file
	- Open Project should open a dialog that asks for the project directory and loads the project from the .NFProject file and will store its information in the same persistent editor file
	- Continue Last Project should open the project that was last opened according to the persistent editor file, if it no longer exists a warning dialog should pop up and it will be removed from that file
	- Exit should close the editor
	- The editor should have a menu bar at the top with the following options: File, View, Help
	- File should have the following options: New Project, Open Project, Save Project, Save Project As, Close Project, Settings, and Exit
	- Settings should pop up a dialog that allows the user to change the resolution manually, toggle fullscreen, and more to come as the editor is developed like the ability to change the theme
	- View should have the following options: Full Screen toggle, Preview Game, And more to come as the editor is developed
	- Help should have the following options: Documentation, About, And more to come as the editor is developed
	- The editor window should update its title to the project name when a project is opened for editing and append the scene name when a scene is opened for editing (If this is possible)
	- The editor should have a main window that is 1280x720 and centered on the screen but is fully resizable and remembers its size and position between sessions
	- When a project is created, it will create a projects folder in the directory where the editor application is located and create a folder in that directory with the project name and create a project.NFProject file in that directory
	- It will also create a default <project name>.go file in that directory that will contain the fyne app and window and the main menu set as well as a layouts folder, scene folder, and functions folder in the project directory
	- It will also add a DefaultLayouts.go file to the layouts folder that will contain all the default layouts and a DefaultFunctions.go file to the functions folder that will contain all the default functions
	- And it will add a menu's folder to the scene folder that will contain all the default menus
	- In the main game.go file it will first open a small window with just some settings buttons like volume and a toggle to open that menu in the future and a start button that will open the main menu scene
	- If the toggle is checked it will open the menu scene on game launch instead of the settings window
	- When a project is opened the editor should load the project.NFProject file and deserialize it into a project struct and then load the project into the editor as the ActiveProject
	- It should then scan the scenes folder in the project directory and look for .NFScene files and load them into the editor as ungrouped scenes, any folders within the scenes folder become SceneGroups with the .NFScene files in them as scenes
	- If the scene is not in the scenes folder it should be ignored
	- All scene groups and scenes should be added to the ActiveProject struct
	- The Active Project is then loaded into the Window as a tree view with the project name as the root node and then all the project struct values as children of the root node with the ungrouped scenes and scene groups at the bottom of the tree with scene groups coming before ungrouped scenes
	- When an object is clicked in the tree view it should pop up a basic editor for that object that allows the user to edit the values of the object and save them back to the project struct (Scenes and functions will have a more advanced editor that takes up the right side of the main window)
	- there should also be a search bar at the top of the tree view that allows the user to search for a specific scene, scene group, or function by name but won't filter out the credits, author, etc.
	- When a scene group is clicked the user should have the option to rename it or group it into another scene group or ungroup it
	- When a scene is clicked the user should have the option to rename it or move it to another scene group or ungroup it
	- Right under the search bar should be an add scene button, and each scene group should have an add scene button next to it
	- When a scene is added first a dialog should pop up asking for the scene name and layouts type and then a new scene should be created with the given name and layouts type and added to the project struct and the tree view
	- When saved the scene should be saved to its appropriate .NFScene file according to where it should be in the scenes folder
	- Scenes should have an edit button next to them that opens the scene editor for that scene
	- In the scene editor there should be a tree view on the left side that shows all the objects in the scene and allows the user to select them and edit their values containers should be able to be expanded and collapsed
	- Objects should be displayed as their type and name in the tree view like type: name
	- At the top of the object tree view there should be a search bar that allows the user to search for a specific object by name
	- Underneath the search bar there should be a button that allows the user to add a new object to the scene that will pop up a dialog with a drop down menu that is populated with all the available object types from the default widgets and containers and the custom ones and a text box that allows the user to name the object
	- When the user clicks the add button the object should be added to the scene and the object tree view and the user should be able to edit its values
	- Values are populated into the object editor based on the type of object selected custom objects will pull the types from the json data and populate the editor with the correct type fields
	- When the user clicks an object in the tree view the object editor should be populated with the objects values and the user should be able to edit them
	- When the user clicks the save button the object editor should save the values back to the object and the object tree view should update to reflect the changes
	- Above the search bar in the main project tree, there should be a build game button that will build the game and save it to the project directory in a build folder
	- When the build button is clicked it should give options for the user to build the game for windows, mac, linux, android, and ios


TODO Secondary Editor Requirements:
	- The editor should have a terminal window that can be opened and closed from the view menu (Create this just using a table with the first line being the timestamp, second being the status(Info, warning, NFerror, etc, and third being the message, with a double click on that row copying the message to the clipboard)
	- All editor config data and projects should be stored in the os.UserConfigDir() directory in a NovellaForge folder
	- ALL extra editor files outside of the application itself should be generated or downloaded to the os.UserConfigDir() directory in a NovellaForge folder
	- Game should also have a terminal window that can be opened and closed from the view menu (Create this just using a table with the first line being the timestamp, second being the status(Info, warning, NFerror, etc, and third being the message, with a double click on that row copying the message to the clipboard)

Todo IDEAS:
	- A game window will have a main menu at the top with options to save, load, view credits, set preferences, exit, etc.
	- On game launch there will be a splash screen with the game name and version and a button to start the game or view credits or change settings
	- in game f1 will quick save to the oldest or first empty save slot and f2 will quick load from the newest save slot pressing f5 through f8 will save to those slots and f9 through f12 will load from those slots (Max of 4 quick saves and 4 quick loads)
	- The game will have a settings menu that will allow the user to change the resolution, toggle fullscreen, change the volume, and change the keybindings
	- The game will have a credits menu that will display the credits from the project.NFProject file
	- The game will allow normal saves that will be stored in the game directory in a saves folder there will be no limit to the number of saves the user can make they will be named save1.novella save2.novella etc. The user can name their saves whatever they want (Saves will just be json arrays anyway)
	- The game will have a load menu that will display all saves in the saves folder and allow the user to load them
	- When building check if the author/credits etc are empty and if they are prompt the user to fill them in
	- Check for the proper fyne setup and install it if it is not there (fyne setup)
	- Check for the fyne dependencies and install them if they are not there (fyne install)

*/

// main is the entry point for the application
func main() {
	application := app.New()
	window := application.NewWindow(NFEditor.WindowTitle)
	window.Resize(fyne.NewSize(1280, 720))

	//Convert the PNG icon to a fyne resource
	iconResource, err := fyne.LoadResourceFromPath(NFEditor.Icon)
	if err != nil {
		log.Printf("Failed to load icon: %v", err)
		application.SetIcon(theme.FileApplicationIcon())
	} else {
		application.SetIcon(iconResource)
		window.SetIcon(application.Icon())
	}

	NFEditor.CreateMainContent(window)

	go SplashScreenLoop(window)
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
			widget.NewLabelWithStyle("Version: "+NFEditor.Version, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabelWithStyle("Developed By: "+NFEditor.Author, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabelWithStyle("Powered By: Fyne", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		)
		splash.SetContent(splashBox)
		splash.Show()
		go func() {
			time.Sleep(time.Second * 3)
			splash.Close()
			window.Show()
		}()
	}
}
