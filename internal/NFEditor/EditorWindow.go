package NFEditor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFConfig"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFLog"
	"log"
	"os"
	"path/filepath"
	"time"
)

// CreateMainContent updates the loading variable as the NovellaForge content is created
func CreateMainContent(window fyne.Window) {
	// Runs "go version" to check if Go is installed
	//loading.SetProgress(10, "Checking Dependencies")
	CheckAndInstallDependencies(window)

	// Creates a main menu to hold the buttons below
	//loading.SetProgress(30, "Creating Main Menu")
	CreateMainMenu(window)

	// Create a grid layout for the four main buttons
	//loading.SetProgress(50, "Creating Main Content")
	grid := CreateMainGrid(window)
	//Fetch the buttons from the grid to allow for disabling them
	openRecentButton := grid.(*fyne.Container).Objects[2].(*widget.Button)
	continueLastButton := grid.(*fyne.Container).Objects[3].(*widget.Button)

	//loading.SetProgress(70, "Initializing Buttons")
	err := InitButtons(window, continueLastButton, openRecentButton)
	if err != nil {
		//Show an error dialog
		dialog.ShowError(err, window)
		return
	}
	//loading.SetProgress(100, "Setting Content")
	window.SetContent(grid)
	//loading.Complete()
}

func InitButtons(window fyne.Window, continueLastButton *widget.Button, openRecentButton *widget.Button) error {
	projects, err := ReadProjectInfo()
	if err != nil {
		//Show an error dialog
		dialog.ShowError(err, window)
		return err
	}
	var project NFInfo
	if len(projects) == 0 {
		continueLastButton.Disable()
		openRecentButton.Disable()
	} else {
		//Get the most recently opened project
		project = projects[0]
		for i := 0; i < len(projects); i++ {
			if projects[i].OpenDate.After(project.OpenDate) {
				project = projects[i]
			}
		}
		//Os stat the project path to see if it still exists
		_, err = os.Stat(project.Path)
		if err != nil {
			//If the project does not exist, disable the continue last button
			continueLastButton.Disable()
		}

	}
	continueLastButton.OnTapped = func() {
		err = OpenFromInfo(project, window)
		if err != nil {
			//Show an error dialog
			dialog.ShowError(err, window)
		}
	}
	return nil
}

// CreateMainGrid creates the main grid layout for the main window
func CreateMainGrid(window fyne.Window) fyne.CanvasObject {
	grid := container.New(layout.NewGridLayout(2))

	// Create a New Project button in the top left
	newProjectButton := widget.NewButton("New Project", func() {
		NewProjectDialog(window)
	})
	// Create an Open Project button in the top right
	openProjectButton := widget.NewButton("Open Project", func() {
		OpenProjectDialog(window)
	})
	// Create an Open Recent button in the bottom left
	var openRecentButton *widget.Button
	openRecentButton = widget.NewButton("Open Recent", func() {
		err := OpenRecentDialog(window)
		if err != nil {
			//If there are no projects, or another issue occurs, disable the button
			openRecentButton.Disable()
		}
	})
	// Create a Continue Last button in the bottom right
	continueLastButton := widget.NewButton("Continue Last", func() {})
	//Add the buttons to the grid
	grid.Add(newProjectButton)
	grid.Add(openProjectButton)
	grid.Add(openRecentButton)
	grid.Add(continueLastButton)
	return grid
}

type OpenPage int // Enum for the different pages of the editor to allow changing the menu bar based on the page
const (
	HomePage OpenPage = iota
	EditorPage
)

func CreateMainMenu(window fyne.Window) {
	openRecentMenu := fyne.NewMenuItem("Open Recent", func() {
		_ = OpenRecentDialog(window)
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
			notFound := false
			_, err := os.Stat(projects[i].Path)
			if err != nil {
				notFound = true
			}
			var newMenuItem *fyne.MenuItem
			if notFound {
				newMenuItem = fyne.NewMenuItem(projects[i].Name+" (Not Found)", func() {
					err = UpdateProjectInfo(projects[i])
					if err != nil {
						//Show an error dialog
						dialog.ShowError(err, window)
					}
					CreateMainMenu(window)
				})
			} else {
				newMenuItem = fyne.NewMenuItem(projects[i].Name, func() {
					err = OpenFromInfo(projects[i], window)
					if err != nil {
						dialog.ShowError(err, window)
					}
				})
			}
			projectMenu.Items = append(projectMenu.Items, newMenuItem)
		}
	}
	openRecentMenu.ChildMenu = projectMenu
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Home", func() {
				grid := CreateMainGrid(window)
				continueButton := grid.(*fyne.Container).Objects[3].(*widget.Button)
				openRecentButton := grid.(*fyne.Container).Objects[2].(*widget.Button)
				err = InitButtons(window, continueButton, openRecentButton)
				if err != nil {
					dialog.ShowError(err, window)
				} else {
					CreateMainMenu(window)
					window.SetContent(grid)
				}
			}),
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

func OpenRecentDialog(window fyne.Window) error {
	box := container.NewVBox()
	newDialog := dialog.NewCustom("Open Recent", "Close", box, window)
	projects, err := ReadProjectInfo()
	if err != nil {
		//Show an error dialog
		dialog.ShowError(err, window)
		return err
	}
	if len(projects) == 0 {
		//Show an error dialog
		dialog.ShowError(NFError.ErrNoProjects, window)
		return err
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
			project := projects[id-1]
			_, err := os.Stat(project.Path)
			if err != nil {
				item.(*fyne.Container).Objects[0].(*widget.Label).SetText(project.Name + " (Not Found, click to hide)")
			} else {
				item.(*fyne.Container).Objects[0].(*widget.Label).SetText(project.Name)
			}
			item.(*fyne.Container).Objects[2].(*widget.Label).SetText(project.OpenDate.Format("01/02/2006 13:04"))
		})
	list.OnSelected = func(id widget.ListItemID) {
		if id == 0 {
			return
		}
		project := projects[id-1]
		_, err := os.Stat(project.Path)
		if err != nil {
			err := UpdateProjectInfo(project)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			projects, err = ReadProjectInfo()
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			list.Refresh()
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
	return nil
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
		err = project.Load(window)
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
	projectInfo := NFInfo{}
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
		config := NFConfig.NewConfig(projectName, "0.0.1", authorEntry.Text, "Created with NovellaForge")
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Println("Error getting home directory")
			dialog.ShowError(err, window)
			return
		}
		novellaForgeDir := homeDir + "/Documents/NovellaForge"
		projectInfo.Name = projectName
		projectsDir := fyne.CurrentApp().Preferences().StringWithFallback("projectDir", novellaForgeDir+"/projects")
		projectPath := projectsDir + "/" + projectName + "/" + projectName + ".NFProject"
		projectPath = filepath.Clean(projectPath)
		projectInfo.Path = projectPath
		projectInfo.OpenDate = time.Now()
		newProject := NFProject{
			Info:   projectInfo,
			Config: config,
		}

		err = newProject.Create(window)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		log.Printf("Created project %s", newProject.Info.Name)
		projectDialog.Hide()

		err = newProject.Info.Load(window)
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
