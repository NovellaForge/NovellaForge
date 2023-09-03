package main

import (
	"NovellaForge/pkg/editor"
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"os"
)

/*
Todo:
	- Begin the Game Compiler
		- It should also have a build button with checkboxes for windows, mac, and linux (The build button should be disabled if the project has not been saved and each os will build for all architectures)
		- The build button should also have a checkbox for whether or not to zip the build
		- The build button should also have a checkbox for whether or not to include the source code
	- Determine Character Format and how to store them and make them send messages to the engine (Currently I am thinking of having all scenes that are not menus have a hidden dialog/narration box that is always there and the character will send messages to that box)
	- Add scene Component and Character creation to the editor
	- Fix the file saving for the scene editor for adding and removing components
	- Make sure that all scene components and scene types and things that players should not be able to change are encrypted in the project.novel file on build with a toggle in the editor build settings to encrypt the full file or just the things that should be encrypted



Todo IDEAS:
	- Update Regex for project creation to allow only the valid characters that are a union of the valid characters for all operating systems we want to support
	- Engine should eventually support mac os and linux as well as windows
	- The scenes will be opened using a populateScene function that takes in a Scene and args and just sets the game field contents to the scene contents
	- A game window will have a main menu at the top with options to save, load, view credits, set preferences, exit, etc.
	- On game launch there will be a splash screen with the game name and version and a button to start the game or view credits or change settings
	- in game f1 will quick save to the oldest or first empty save slot and f2 will quick load from the newest save slot pressing f5 through f8 will save to those slots and f9 through f12 will load from those slots (Max of 4 quick saves and 4 quick loads)
	- The game will have a settings menu that will allow the user to change the resolution, toggle fullscreen, change the volume, and change the keybindings
	- The game will have a credits menu that will display the credits from the project.novel file
	- The game will allow normal saves that will be stored in the game directory in a saves folder there will be no limit to the number of saves the user can make they will be named save1.novella save2.novella etc. The user can name their saves whatever they want (Saves will just be json arrays anyway)
	- The game will have a load menu that will display all saves in the saves folder and allow the user to load them
	- When building check if the author/credits etc are empty and if they are prompt the user to fill them in
	- Check for the proper fyne setup and install it if it is not there (fyne setup)
	- Check for the fyne dependencies and install them if they are not there (fyne install)

*/

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
	window := application.NewWindow(fmt.Sprintf("NovellaForge Editor %s", editor.Version))

	err := editor.CreateMainWindow(window)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	window.ShowAndRun()
}
