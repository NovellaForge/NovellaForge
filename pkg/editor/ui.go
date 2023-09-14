package editor

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"os"
)

func CreateMainContent(window fyne.Window) {
	//TODO: Create the main content of the editor
}

func CreateMainMenu(window fyne.Window) {

	openRecentMenu := fyne.NewMenuItem("Open Recent", func() {
		OpenRecentDialog(window)
	})
	projectMenu := fyne.NewMenu("Project")

	projects, err := ReadProjectInfo()
	if err != nil {
		//Show an error dialog
		dialog.NewError(err, window)
	}
	if len(projects) == 0 {
		openRecentMenu.Disabled = true
	} else {
		openRecentMenu.Disabled = false
		//Create a list of all the projects as menu items of the child menu
		for i := 0; i < len(projects); i++ {
			newMenuItem := fyne.NewMenuItem(projects[i].Name, func() {
				err = OpenRecentProject(projects[i], window)
				if err != nil {
					//Show an error dialog
					dialog.NewError(err, window)
				}
			})
			projectMenu.Items = append(projectMenu.Items, newMenuItem)
		}
	}
	openRecentMenu.ChildMenu = projectMenu
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New Project", func() {
				NewProjectDialog(window)
			}),
			fyne.NewMenuItem("Open Project", func() {
				OpenProjectDialog(window)
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
		),
		fyne.NewMenu("Help",
			fyne.NewMenuItem("Documentation", func() {
				//TODO: Open the documentation page
			}),
			fyne.NewMenuItem("About", func() {
				//TODO: Open a dialog that shows the about information
			}),
		),
	)

	window.SetMainMenu(mainMenu)
}

func OpenRecentDialog(window fyne.Window) {
	box := container.NewVBox()
	newDialog := dialog.NewCustom("Open Recent", "Open", box, window)
	projects, err := ReadProjectInfo()
	if err != nil {
		//Show an error dialog
		dialog.NewError(err, window)
		return
	}
	if len(projects) == 0 {
		//Show an error dialog
		dialog.NewError(ErrNoProjects, window)
		return
	}

	//Create a scrollable list of all the projects
	list := widget.NewList(
		func() int {
			return len(projects)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(projects[id].Name)
		})
	list.OnSelected = func(id widget.ListItemID) {
		err = OpenRecentProject(projects[id], window)
		if err != nil {
			//Show an error dialog
			dialog.NewError(err, window)
		}
	}
	scrollBox := container.NewVScroll(list)

	openProjectButton := widget.NewButton("Open Project", func() {
		OpenProjectDialog(window)
		newDialog.Hide()
	})
	box.Add(scrollBox)
	box.Add(openProjectButton)
	newDialog.Resize(fyne.NewSize(800, 600))
	newDialog.Show()
}

func OpenProjectDialog(window fyne.Window) {
	// Create a custom file dialog
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			// Show an error dialog
			dialog.NewError(err, window)
			return
		}

		//If the reader is nil, then the user canceled the dialog and nothing should happen
		if reader == nil {
			return
		}

		// Read the file
		file, err := os.ReadFile(reader.URI().Path())
		if err != nil {
			// Show an error dialog
			dialog.NewError(err, window)
			return
		}
		// Deserialize the project
		project, err := DeserializeProject(file)
		if err != nil {
			// Show an error dialog
			dialog.NewError(err, window)
			return
		}
		// Load the project
		err = LoadProject(project, window)
		if err != nil {
			// Show an error dialog
			dialog.NewError(err, window)
			return
		}
	}, window)

	// Set filter for .NFProject files
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".NFProject"}))

	// Set the starting directory to the Projects directory
	projectURI, err := storage.ListerForURI(storage.NewFileURI("projects"))
	if err == nil {
		fileDialog.SetLocation(projectURI)
	}

	//Set the size of the dialog to 800x600
	fileDialog.Resize(fyne.NewSize(800, 600))

	// Show the dialog
	fileDialog.Show()
}

func NewProjectDialog(window fyne.Window) {
	//Create a dialog that allows the user to create a new project
	var err error
	box := container.NewVBox()
	projectDialog := dialog.NewCustom("New Project", "Cancel", container.NewVBox(layout.NewSpacer(), box, layout.NewSpacer()), window)
	newProject := Project{}
	projectName := ""
	nameValidationLabel := widget.NewLabel("")
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Project Name")
	nameEntry.Validator = func(s string) error {
		_, err = sanitizeProjectName(s)
		if err != nil {
			nameValidationLabel.SetText(err.Error())
		} else {
			nameValidationLabel.SetText("")
		}
		return err
	}
	nameConfirmButton := widget.NewButton("Confirm", func() {})
	nameEntry.SetOnValidationChanged(func(e error) {
		if e != nil {
			nameConfirmButton.Disable()
		} else {
			nameConfirmButton.Enable()
		}
	})
	authorEntry := widget.NewEntry()
	authorEntry.SetPlaceHolder("Author")
	authorConfirmButton := widget.NewButton("Confirm", func() {})
	authorBackButton := widget.NewButton("Back", func() {})
	nameConfirmButton.OnTapped = func() {
		_, err = sanitizeProjectName(nameEntry.Text)
		if err != nil {
			dialog.NewError(err, window)
			nameEntry.SetValidationError(err)
			return
		}
		projectName = nameEntry.Text
		box.Remove(nameEntry)
		box.Remove(nameConfirmButton)
		box.Remove(nameValidationLabel)
		box.Add(authorEntry)
		box.Add(authorConfirmButton)
		box.Add(authorBackButton)
	}
	authorConfirmButton.OnTapped = func() {
		newProject.GameName = projectName
		newProject.Author = authorEntry.Text
		err = CreateProject(newProject, window)
		if err != nil {
			dialog.NewError(err, window)
			return
		}
		projectDialog.Hide()
	}
	authorBackButton.OnTapped = func() {
		box.Remove(authorEntry)
		box.Remove(authorConfirmButton)
		box.Remove(authorBackButton)
		box.Add(nameValidationLabel)
		box.Add(nameEntry)
		box.Add(nameConfirmButton)
	}
	box.Add(nameValidationLabel)
	box.Add(nameEntry)
	box.Add(nameConfirmButton)
	projectDialog.Resize(fyne.NewSize(400, 300))
	projectDialog.Show()
}

func CreateProject(project Project, window fyne.Window) error {
	//First check if the project directory already exists
	if _, err := os.Stat("projects/" + project.GameName); !os.IsNotExist(err) {
		return ErrProjectAlreadyExists
	} else {
		//Create the project directory
		err = os.Mkdir("projects/"+project.GameName, 0755)
		if err != nil {
			return err
		}
		//Create the project file
		err = os.WriteFile("projects/"+project.GameName+"/"+project.GameName+".NFProject", SerializeProject(project), 0644)
		if err != nil {
			return err
		}
		//TODO Create the rest of the project after we have built it in the editor

	}
	return nil
}

func SerializeProject(project Project) []byte {
	//Marshal the project to JSON
	serializedProject, err := json.Marshal(project)
	if err != nil {
		return nil
	}
	return serializedProject

}
