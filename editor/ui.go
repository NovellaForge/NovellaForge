package editor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"os"
)

func CreateMainContent(window fyne.Window) {
	//TODO: Create the main content of the editor
}

func CreateMainMenu(window fyne.Window) {

	openRecentMenu := fyne.NewMenuItem("Open Recent", func() {
		//TODO: Open a dialog that shows a scrollable list of all the projects that have been opened in the past and will open the selected project
		// Also add a button to just open a project from the file system
	})
	openRecentMenu.ChildMenu = fyne.NewMenu("Projects")

	projects, err := ReadProjectInfo()
	if err != nil {
		//Show an error dialog
		dialog.NewError(err, window)
	}
	if len(projects) == 0 {
		//TODO: Disable the open recent menu
	}
	//TODO: Add the projects to the open recent menu sorted by last open date

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New Project", func() {
				//TODO: Open a dialog that asks for the project name and the project directory and stores that information and its last open date in some form of persistent editor file
			}),
			fyne.NewMenuItem("Open Project", func() {
				//TODO: Open a dialog that asks for the project directory and loads the project from the .NFProject file and will store its information in the same persistent editor file
			}),
			openRecentMenu,
			fyne.NewMenuItem("Quit", func() {
				os.Exit(0)
			}),
		),
		fyne.NewMenu("View",
			fyne.NewMenuItem("Full Screen", func() {
				window.SetFullScreen(!window.FullScreen())
			}),
			fyne.NewMenuItem("Preview Game", func() {
				//TODO: Open a second window that shows the game preview
			}),
		),
		fyne.NewMenu("Help",
			fyne.NewMenuItem("Documentation", func() {
				//TODO: Open the documentation in the default browser
			}),
			fyne.NewMenuItem("About", func() {
				//TODO: Open a dialog that shows the about information
			}),
		),
	)

	window.SetMainMenu(mainMenu)
}
