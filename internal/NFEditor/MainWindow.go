package NFEditor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/pkg/NFError"
	"github.com/NovellaForge/NovellaForge/pkg/NFLog"
	"log"
	"os"
	"time"
)

func CreateMainMenu(window fyne.Window) {

	openRecentMenu := fyne.NewMenuItem("Open Recent", func() {
		OpenRecentDialog(window)
	})
	projectMenu := fyne.NewMenu("Project")

	projects, err := ReadProjectInfo()
	if err != nil {
		//Show an error dialog
		dialog.ShowError(err, window)
	}
	if len(projects) == 0 {
		openRecentMenu.Disabled = true
	} else {
		openRecentMenu.Disabled = false
		//Create a list of all the projects as menu items of the child menu
		for i := 0; i < len(projects); i++ {
			newMenuItem := fyne.NewMenuItem(projects[i].Name, func() {
				err = OpenFromInfo(projects[i], window)
				if err != nil {
					//Show an NFerror dialog
					dialog.ShowError(err, window)
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
			fyne.NewMenuItem("Preview Game", func() {
				//TODO: Open a preview of the game
			}),
			fyne.NewMenuItem("Terminal", func() {
				NFLog.ShowDialog(window)
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
	newDialog := dialog.NewCustom("Open Recent", "Close", box, window)
	projects, err := ReadProjectInfo()
	if err != nil {
		//Show an error dialog
		dialog.ShowError(err, window)
		return
	}
	if len(projects) == 0 {
		//Show an error dialog
		dialog.ShowError(NFError.ErrNoProjects, window)
		return
	}

	//Sort the projects by the last opened date
	for i := 0; i < len(projects); i++ {
		for j := 0; j < len(projects); j++ {
			if projects[i].OpenDate.After(projects[j].OpenDate) {
				temp := projects[i]
				projects[i] = projects[j]
				projects[j] = temp
			}
		}
	}

	//Create a scrollable list of all the projects
	list := widget.NewList(
		func() int {
			return len(projects) + 1
		},
		func() fyne.CanvasObject {
			nameLabel := widget.NewLabel("Name")
			lastOpenedLabel := widget.NewLabel("Last Opened")
			//Set the style of the labels to bold
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}
			lastOpenedLabel.TextStyle = fyne.TextStyle{Bold: true}
			// Align the labels to the center
			nameLabel.Alignment = fyne.TextAlignCenter
			lastOpenedLabel.Alignment = fyne.TextAlignCenter
			return container.NewHBox(widget.NewLabel("Name"), layout.NewSpacer(), widget.NewLabel("Last Opened"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id == 0 {
				return
			}
			//Set the first label to the project name and the second to the last opened date
			item.(*fyne.Container).Objects[0].(*widget.Label).SetText(projects[id-1].Name)
			item.(*fyne.Container).Objects[2].(*widget.Label).SetText(projects[id-1].OpenDate.Format("01/02/2006 13:04"))
		})
	list.OnSelected = func(id widget.ListItemID) {
		if id == 0 {
			return
		}
		err = OpenFromInfo(projects[id-1], window)
		if err != nil {
			//Show an error dialog
			errDialog := dialog.NewError(err, window)
			errDialog.SetOnClosed(func() {
				newDialog.Hide()
			})
			errDialog.Show()
		} else {
			newDialog.Hide()
		}
	}

	scrollBox := container.NewVScroll(list)
	scrollBox.SetMinSize(fyne.NewSize(400, 300))
	box.Add(scrollBox)
	newDialog.Resize(fyne.NewSize(800, 600))
	newDialog.Show()
}

func OpenProjectDialog(window fyne.Window) {
	// Create a custom file dialog
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			log.Println("Error opening file")
			// Show an error dialog
			dialog.ShowError(err, window)
			return
		}

		//If the reader is nil, then the user canceled the dialog and nothing should happen
		if reader == nil {
			return
		}

		// Read the file
		file, err := os.ReadFile(reader.URI().Path())
		if err != nil {
			log.Println("Error reading file")
			// Show an error dialog
			dialog.ShowError(err, window)
			return
		}
		// Deserialize the project
		project, err := Deserialize(file)
		if err != nil {
			log.Println("Error deserializing project")
			// Show an error dialog
			dialog.ShowError(err, window)
			return
		}
		// Load the project

		info := ProjectInfo{
			Name:     project.GameName,
			Path:     reader.URI().Path(),
			OpenDate: time.Now(),
		}
		err = project.Load(window, info)
		if err != nil {
			// Show an error dialog
			log.Println("Error loading project")
			dialog.ShowError(err, window)
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
	nameEntry := widget.NewEntry()
	nameValidationLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	nameEntry.SetPlaceHolder("Project Name")
	nameConfirmButton := widget.NewButton("Confirm", func() {})
	authorEntry := widget.NewEntry()
	authorEntry.SetPlaceHolder("Author")
	authorConfirmButton := widget.NewButton("Create Project", func() {})
	authorBackButton := widget.NewButton("Back", func() {})
	//Add everything to the box
	box.Add(nameValidationLabel)
	box.Add(nameEntry)
	box.Add(nameConfirmButton)
	box.Add(authorEntry)
	box.Add(authorConfirmButton)
	box.Add(authorBackButton)
	authorConfirmButton.Hide()
	authorBackButton.Hide()
	authorEntry.Hide()

	nameEntry.Validator = func(s string) error {
		_, err = SanitizeProjectName(s)
		if err != nil {
			nameValidationLabel.SetText(err.Error())
		} else {
			nameValidationLabel.SetText("")
		}
		return err
	}

	nameEntry.SetOnValidationChanged(func(e error) {
		if e != nil {
			nameConfirmButton.Disable()
		} else {
			nameConfirmButton.Enable()
		}
	})
	nameConfirmButton.OnTapped = func() {
		_, err = SanitizeProjectName(nameEntry.Text)
		if err != nil {
			dialog.ShowError(err, window)
			nameEntry.SetValidationError(err)
			return
		}
		projectName = nameEntry.Text
		nameEntry.Hide()
		nameConfirmButton.Hide()
		nameValidationLabel.Hide()
		authorEntry.Show()
		authorConfirmButton.Show()
		authorBackButton.Show()
	}
	authorConfirmButton.OnTapped = func() {
		newProject.GameName = projectName
		newProject.Author = authorEntry.Text
		newProject.Version = "0.0.1"
		newProject.Credits = "Created with NovellaForge"
		err = newProject.Create(window)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		log.Printf("Created project %s", newProject.GameName)
		projectDialog.Hide()
		err := newProject.Load(window)
		if err != nil {
			return
		}
	}
	authorBackButton.OnTapped = func() {
		authorConfirmButton.Hide()
		authorBackButton.Hide()
		authorEntry.Hide()
		nameEntry.Show()
		nameConfirmButton.Show()
		nameValidationLabel.Show()
	}
	projectDialog.Resize(fyne.NewSize(400, 300))
	projectDialog.Show()
}
