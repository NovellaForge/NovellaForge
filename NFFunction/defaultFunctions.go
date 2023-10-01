package NFFunction

import (
	"NovellaForge/NFSave"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Quit(window fyne.Window, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error) {
	dialog.ShowConfirm("Are you sure you want to quit?", "Are you sure you want to quit?", func(b bool) {
		if b {
			window.Close()
		}
	}, window)
	return nil, nil, nil
}

// CustomError creates a dialog for when an error occurs in the game
func CustomError(window fyne.Window, args map[string]interface{}) (results map[string]interface{}, widgets map[string]fyne.CanvasObject, err error) {
	isErrored := true
	message := args["message"].(string)
	var errDialog *dialog.CustomDialog
	errTextLabel := widget.NewLabel(message)
	errBox := container.NewVBox(
		errTextLabel,
		container.NewHBox(
			widget.NewButton("Attempt to Continue", func() {
				errDialog.Hide()
				isErrored = false
			}),
			widget.NewButton("Close Game", func() {
				window.Close()
			}),
		),
	)
	errDialog = dialog.NewCustomWithoutButtons("Error", errBox, window)
	errDialog.SetOnClosed(func() {
		isErrored = false
	})
	errDialog.Show()
	for isErrored {
		//Sleep the thread for 0.1 seconds
		time.Sleep(100 * time.Millisecond)
	}
	return nil, nil, nil
}

// NewGame creates a new game save file and starts the game
func NewGame(window fyne.Window, args map[string]interface{}) (results map[string]interface{}, widgets map[string]fyne.CanvasObject, err error) {
	//Check if the NewGameScene in the args exists in the args
	if _, ok := args["NewGameScene"]; !ok {
		return nil, nil, errors.New("no NewGameScene in args")
	}
	//Check if the NewGameScene in the args is not nil or empty
	if args["NewGameScene"] == nil || args["NewGameScene"] == "" {
		return nil, nil, errors.New("NewGameScene is nil or empty")
	}
	//Get the NewGameScene from the args
	newGameScene := args["NewGameScene"].(string)

	if NFSave.Active != nil {
		saveAsButton := func() {
			_, _, _ = SaveAs(window, args)
		}
		dialog.ShowCustomConfirm("Do you want to save your current game first?", "Save Game", "Don't Save", widget.NewButton("Save Game As", saveAsButton), func(b bool) {
			if b {
				//Save the game
				err = NFSave.Active.Save()
				name := "save"
				if errors.Is(err, NFSave.ErrSaveNameNotSet) {
					//Walk the directory making sure the save name is unique and if it is not, add a count and check if the count is unique incrementing the count until it is
					_ = filepath.WalkDir(NFSave.Directory, func(path string, d os.DirEntry, err error) error {
						//If the file is a saves.saveExtension file, check if the name matches the save name
						if filepath.Base(path) == name+NFSave.Extension {
							//If the save name matches, increment the count and return
							name += "1"
						}
						return nil
					})
					//Set the save name to the new name
					NFSave.Active.Name = name
					//Save the game
					err = NFSave.Active.Save()
					if err != nil {
						_, _, _ = CustomError(window, map[string]interface{}{"window": args["window"], "message": err.Error()})
						return
					}
				}
			}
		}, window)
	}

	//Create a new save file
	newSave, err := NFSave.New(newGameScene)
	if err != nil {
		return nil, nil, err
	}
	//Set the active save to the new save
	NFSave.Active = newSave
	return nil, nil, nil
}

// SaveAs saves the game as a new save file
func SaveAs(window fyne.Window, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error) {
	//Popup a dialog asking for the save name
	saveNameEntry := widget.NewEntry()
	saveNameEntry.SetPlaceHolder("Save Name")
	saveNameEntry.SetText(NFSave.Active.Name)
	var cancelDialog *dialog.ConfirmDialog
	var confirmDialog *dialog.ConfirmDialog
	confirmDialog = dialog.NewCustomConfirm("Save As", "Save", "Cancel", saveNameEntry, func(b bool) {
		if b {
			//Walk the save file directory making sure the save name is unique, and if it is not, ask if they are sure they want to overwrite the save
			_ = filepath.WalkDir(NFSave.Directory, func(path string, d os.DirEntry, err error) error {
				//If the file is a saves.saveExtension file, check if the name matches the save name
				if filepath.Base(path) == saveNameEntry.Text+NFSave.Extension {
					//If the save name matches, ask if they are sure they want to overwrite the save
					dialog.ShowConfirm("Overwrite?", "Are you sure you want to overwrite the save?", func(b bool) {
						if b {
							//Set the save name to the entry text
							NFSave.Active.Name = saveNameEntry.Text
							//Save the game
							err := NFSave.Active.Save()
							if err != nil {
								_, _, _ = CustomError(window, map[string]interface{}{"window": args["window"], "message": err.Error()})
								return
							}
						} else {
							//Show the save as dialog again
							confirmDialog.Show()
						}
					}, window)
				}
				return nil
			})
			//Set the save name to the entry text
			NFSave.Active.Name = saveNameEntry.Text
			//Save the game
			err := NFSave.Active.Save()
			if err != nil {
				_, _, _ = CustomError(window, map[string]interface{}{"window": args["window"], "message": err.Error()})
				return
			}
		} else {
			cancelDialog.Show()
		}
	}, window)
	cancelDialog = dialog.NewConfirm("Are you Sure?", "Canceling will not save your game", func(b bool) {
		if b {
			confirmDialog.Hide()
		} else {
			confirmDialog.Show()
		}
	}, window)
	confirmDialog.Show()
	return nil, nil, nil
}

// LoadGame loads a game save file and starts the game
func LoadGame(window fyne.Window, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error) {

	type FileWithModTime struct {
		Name    string
		Path    string
		ModTime time.Time
	}
	var saveFiles []FileWithModTime
	//Walk the save directory and add each save file to the saves map with the file name being the key and the full path being the value
	_ = filepath.WalkDir(NFSave.Directory, func(path string, info os.DirEntry, err error) error {
		//If the file is a saves.saveExtension file, add it to the save map
		if filepath.Ext(path) == NFSave.Extension {
			nameWithoutExtension := strings.TrimSuffix(info.Name(), NFSave.Extension)
			fileInfo, _ := os.Stat(path)
			saveFiles = append(saveFiles, FileWithModTime{nameWithoutExtension, path, fileInfo.ModTime()})
		}
		return nil
	})
	//Sort the saves by the last modified time
	for i := 0; i < len(saveFiles); i++ {
		for j := 0; j < len(saveFiles); j++ {
			if saveFiles[i].ModTime.After(saveFiles[j].ModTime) {
				temp := saveFiles[i]
				saveFiles[i] = saveFiles[j]
				saveFiles[j] = temp
			}
		}
	}
	//Create a list of save names
	var namesList []string
	for _, file := range saveFiles {
		namesList = append(namesList, file.Name)
	}
	var savesMap = make(map[string]string)
	for _, file := range saveFiles {
		savesMap[file.Name] = file.Path
	}
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			_, _, _ = CustomError(window, map[string]interface{}{"window": args["window"], "message": err.Error()})
			return
		}
		//Get the path of the save file and load it
		path := reader.URI().Path()
		NFSave.Active, err = NFSave.Load(path)
		if err != nil {
			_, _, _ = CustomError(window, map[string]interface{}{"window": args["window"], "message": err.Error()})
			return
		}
	}, window)
	saveList := widget.NewList(
		//Length
		func() int {
			return len(namesList)
		},

		//Create Item
		func() fyne.CanvasObject {
			//Should be a save name with a timestamp and a folder button to choose a history save
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewSeparator(),
				widget.NewLabel(""),
				widget.NewButtonWithIcon("", theme.FolderIcon(), func() {}),
			)
		},

		//Update Item
		func(itemID widget.ListItemID, obj fyne.CanvasObject) {
			//Should be a save name with a timestamp and a folder button to choose a history save
			item := obj.(*fyne.Container)
			item.Objects[0].(*widget.Label).SetText(namesList[itemID])
			item.Objects[2].(*widget.Label).SetText(saveFiles[itemID].ModTime.Format("Jan 2 3:04PM"))
			item.Objects[3].(*widget.Button).OnTapped = func() {
				//Get the history files
				var historyFiles []FileWithModTime
				parentPath := filepath.Dir(saveFiles[itemID].Path)
				_ = filepath.WalkDir(parentPath+"/history", func(path string, info os.DirEntry, err error) error {
					if filepath.Ext(path) == NFSave.Extension+"history" {
						nameWithoutExtension := strings.TrimSuffix(info.Name(), NFSave.Extension)
						fileInfo, _ := os.Stat(path)
						historyFiles = append(historyFiles, FileWithModTime{nameWithoutExtension, path, fileInfo.ModTime()})
					}
					return nil
				})
				//Sort the saves by the last modified time
				for i := 0; i < len(historyFiles); i++ {
					for j := 0; j < len(historyFiles); j++ {
						if historyFiles[i].ModTime.After(historyFiles[j].ModTime) {
							temp := historyFiles[i]
							historyFiles[i] = historyFiles[j]
							historyFiles[j] = temp
						}
					}
				}
				//Create a list of save names
				var historyNamesList []string
				var historySaveMap = make(map[string]string)
				//Add the default save named Latest save to the list
				historyNamesList = append(historyNamesList, "Latest Save")
				historySaveMap["Latest Save"] = saveFiles[itemID].Path
				for _, file := range historyFiles {
					historyNamesList = append(historyNamesList, file.Name+" - "+file.ModTime.Format("Jan 2 3:04PM"))
					historySaveMap[file.Name+" - "+file.ModTime.Format("Jan 2 3:04PM")] = file.Path
				}
				dialog.ShowCustom("History", "Close", container.NewVBox(
					widget.NewLabel("Select a save file to load"),
					widget.NewSelect(historyNamesList, func(s string) {
						var err error
						NFSave.Active, err = NFSave.Load(historySaveMap[s])
						if err != nil {
							_, _, _ = CustomError(window, map[string]interface{}{"message": err.Error()})
							return
						}
					}),
				), window)
			}
		},
	)

	vbox := container.NewVBox(
		widget.NewLabel("Select a save file to load"),
		saveList,
		widget.NewButton("Open From File", func() {
			fileDialog.Show()
		}),
	)
	listDialog := dialog.NewCustomWithoutButtons("Load Game", vbox, window)
	listDialog.Show()
	return nil, nil, nil
}

// ContinueGame continues the game from the last save file
func ContinueGame(window fyne.Window, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error) {
	//Walk the save directory
	type FileWithModTime struct {
		Name    string
		Path    string
		ModTime time.Time
	}
	var saveFiles []FileWithModTime
	//Walk the save directory and add each save file to the saves map with the file name being the key and the full path being the value
	_ = filepath.WalkDir(NFSave.Directory, func(path string, info os.DirEntry, err error) error {
		//If the file is a saves.saveExtension file, add it to the save map
		if filepath.Ext(path) == NFSave.Extension {
			nameWithoutExtension := strings.TrimSuffix(info.Name(), NFSave.Extension)
			fileInfo, _ := os.Stat(path)
			saveFiles = append(saveFiles, FileWithModTime{nameWithoutExtension, path, fileInfo.ModTime()})
		}
		return nil
	})

	//if there are no save files, return an error
	if len(saveFiles) == 0 {
		return nil, nil, errors.New("no save files found")
	}

	//Sort the saves by the last modified time
	for i := 0; i < len(saveFiles); i++ {
		for j := 0; j < len(saveFiles); j++ {
			if saveFiles[i].ModTime.After(saveFiles[j].ModTime) {
				temp := saveFiles[i]
				saveFiles[i] = saveFiles[j]
				saveFiles[j] = temp
			}
		}
	}

	//Open the latest save file
	var err error
	NFSave.Active, err = NFSave.Load(saveFiles[0].Path)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}
