package NFEditor

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2/data/binding"
	"github.com/google/uuid"

	"go.novellaforge.dev/novellaforge/pkg/NFData/NFConfig"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFFunction"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFLayout"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFScene"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFWidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
)

/*
TODO: SceneEditor
 [ ] Add in way to use hotkeys to save scene - Need to contact fyne devs to see if this is possible
 [X] Finish new widget/layout creation
 [ ] Integrate type loading from the asset files (i.e functions, layouts, widgets)
 [ ] Add in way to easily change widget to different types
 [ ] Make property editor show required and optional args
 [X] Fix up project settings
 [ ] Build Manager to build the project
 [ ] Migrate Preview to a separate window
 [ ] Add run game button to launch the game from source
 [X] Convert the function parser to make it an actual object
     so that the user can choose from added functions and specify the args
 [] Clean up tree creation and updates using tree data binding
 	[ ] Fix the old create button functions to update the new binding and add a refresh button to reload changes from disk
    [ ] Fix the weird update bug with the objects when switching scenes(Probably make both bindings global variables)
*/

var (
	objectTreeBinding    = binding.NewStringTree()
	selectedObjectTreeId = ""

	sceneTreeMap            = make(map[uuid.UUID]string) //ID to Path
	openSceneBranches       = make([]string, 0)          //Should be paths NOT UUIDs so that it can persist between uuid changes
	sceneTreeBinding        = binding.NewStringTree()
	sceneNameBinding        = binding.NewString()
	selectedSceneTreeNodeID = ""

	functions = make(map[string]NFObjects.AssetProperties)
	layouts   = make(map[string]NFObjects.AssetProperties)
	widgets   = make(map[string]NFObjects.AssetProperties)

	selectedScenePath string
	selectedScene     *NFScene.Scene
	selectedObject    interface{}
	autoSave          = true
	autoSaveTime      = 30 * time.Second
	changesMade       = false

	autoSaveTimer *time.Timer

	scenePreviewWindow fyne.Window
	previewCanvas      fyne.CanvasObject
	propertiesCanvas   fyne.CanvasObject
	selectorCanvas     fyne.CanvasObject
	objectsCanvas      fyne.CanvasObject
)

func loadAssets(initialPath string) error {
	functions = make(map[string]NFObjects.AssetProperties)
	layouts = make(map[string]NFObjects.AssetProperties)
	widgets = make(map[string]NFObjects.AssetProperties)
	initialPath = filepath.Clean(initialPath)
	err := filepath.Walk(initialPath, func(path string, info os.FileInfo, err error) error {
		asset := NFObjects.AssetProperties{}
		isFunc := filepath.Ext(path) == ".NFFunction"
		isLayout := filepath.Ext(path) == ".NFLayout"
		isWidget := filepath.Ext(path) == ".NFWidget"
		if isFunc || isLayout || isWidget {
			err := asset.Load(path)
			if err != nil {
				return err
			}
			if isFunc {
				if _, ok := functions[asset.Type]; ok {
					log.Println("Function Type already exists, overwriting: " + asset.Type)
					log.Println("To prevent this, make sure all function types are unique")
				}
				functions[asset.Type] = asset
			} else if isLayout {
				if _, ok := layouts[asset.Type]; ok {
					log.Println("Layout Type already exists, overwriting: " + asset.Type)
					log.Println("To prevent this, make sure all layout types are unique")
				}
				layouts[asset.Type] = asset
			} else if isWidget {
				if _, ok := widgets[asset.Type]; ok {
					log.Println("Widget Type already exists, overwriting: " + asset.Type)
					log.Println("To prevent this, make sure all widget types are unique")
				}
				widgets[asset.Type] = asset
			}
		}
		return nil
	})
	return err
}

func regenSceneMap(initialPath string) error {

	//Nil the maps
	newSceneTreeData := make(map[string][]string)
	newSceneTreeValue := make(map[string]string)

	type path string
	var initialUUID uuid.UUID
	//UUID to children UUID
	tempData := make(map[uuid.UUID][]uuid.UUID)
	//Path to UUID
	tempValue := make(map[path]uuid.UUID)
	//UUID to Path
	tempMap := make(map[uuid.UUID]path)
	//UUID List for unique UUIDs (Probably not needed, but I am covering all bases)
	uuidList := make([]uuid.UUID, 0)

	//First walk initializing each path with a UUID
	err := filepath.Walk(initialPath, func(walkPath string, info os.FileInfo, err error) error {
		newUUID := uuid.New()
		for slices.Contains(uuidList, newUUID) {
			newUUID = uuid.New()
		}
		uuidList = append(uuidList, newUUID)
		tempMap[newUUID] = path(walkPath)
		tempValue[path(walkPath)] = newUUID
		if walkPath == initialPath {
			initialUUID = newUUID
		}
		return nil
	})
	if err != nil {
		return err
	}

	//Second walk to create the tree
	err = filepath.Walk(initialPath, func(walkPath string, info os.FileInfo, err error) error {
		if walkPath == initialPath {
			return nil
		}
		//Get the parent path
		parentPath := filepath.Dir(walkPath)
		parentUUID := tempValue[path(parentPath)]
		//Get the current path
		currentUUID := tempValue[path(walkPath)]
		//Add the current path to the parent path
		tempData[parentUUID] = append(tempData[parentUUID], currentUUID)
		return nil
	})

	//Convert the temp maps to the new maps the InitialPath is the root and should have an id of "" in the new maps
	for key, val := range tempData {
		newKey := key.String()
		if key == initialUUID {
			newKey = ""
		}
		newSceneTreeData[newKey] = make([]string, 0)
		for _, v := range val {
			if v == initialUUID {
				continue
			}
			newSceneTreeData[newKey] = append(newSceneTreeData[newKey], v.String())
		}
	}
	for key, val := range tempMap {
		newKey := key.String()
		if key == initialUUID {
			continue
		}
		newSceneTreeValue[newKey] = tempValue[val].String()
	}

	log.Println("Reloading Scene Tree")
	//Nil the scene tree map which is uuid to path
	sceneTreeMap = make(map[uuid.UUID]string)
	for key, val := range tempMap {
		sceneTreeMap[key] = string(val) //Convert the path to a string
	}
	return sceneTreeBinding.Set(newSceneTreeData, newSceneTreeValue)
}

func refreshForm(objectKey string, form *widget.Form, object NFData.CoupledObject, window fyne.Window) {
	form.Items = nil
	form.Refresh()
	formItems := make([]*widget.FormItem, 0)
	formItems = append(formItems, FormButtons(objectKey, form, object, window))
	objectKeys := object.Keys()
	for _, key := range objectKeys {
		val, b := object.Get(key)
		if !b {
			dialog.ShowError(errors.New("failed to get key"), window)
		}
		formItems = append(formItems, CreateParamItem(key, val, object, objectKey, form, window))
	}
	form.Items = formItems
	form.Refresh()
}

func parseTime(s string) (time.Duration, error) {
	//Check if splitting off the last element results in a number and a character
	timeUnit := ""
	timeVal := 0
	_, err := fmt.Sscanf(s, "%d%s", &timeVal, &timeUnit)
	if err != nil {
		//Check for just a number
		timeVal, err = strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
	}
	switch timeUnit {
	case "m":
		timeVal = max(timeVal, 1)
		return time.Duration(timeVal) * time.Minute, nil
	case "h":
		timeVal = max(timeVal, 1)
		return time.Duration(timeVal) * time.Hour, nil
	default:
		timeVal = max(timeVal, 30)
		return time.Duration(timeVal) * time.Second, nil
	}
}

func saveScene(window fyne.Window) {
	//log.Println("Check Save Scene")
	if selectedScene != nil && selectedScenePath != "" {
		label := widget.NewButtonWithIcon("Saving Scene", theme.DocumentSaveIcon(), func() {})
		hbox := container.NewHBox(layout.NewSpacer(), label)
		vbox := container.NewVBox(hbox)
		popup := widget.NewPopUp(vbox, window.Canvas())
		popup.Show()
		err := selectedScene.Save(selectedScenePath)
		if err != nil {
			label.Text = "Error Saving Scene"
			label.Icon = theme.ErrorIcon()
			label.Refresh()
			dialog.ShowError(err, window)
		} else {
			changesMade = false
			label.Text = "Scene Saved"
			label.Icon = theme.CheckButtonCheckedIcon()
			label.Refresh()
		}
		time.Sleep(1 * time.Second)
		popup.Hide()
	}
}

func CreateSceneEditor(window fyne.Window) fyne.CanvasObject {
	UpdateMenuBar(window)
	propertiesCanvas = CreateSceneProperties(window)
	selectorCanvas = CreateSceneSelector(window)
	objectsCanvas = CreateSceneObjects(window)
	propertyScroll := container.NewScroll(propertiesCanvas)
	propertyScroll.SetMinSize(fyne.NewSize(300, 0))
	previewCanvas = CreateScenePreview(window)
	MainSplit := container.NewHSplit(
		selectorCanvas,
		container.NewHSplit(
			propertyScroll,
			objectsCanvas,
		),
	)
	MainSplit.Offset = 0.25
	return MainSplit
}

func UpdateMenuBar(window fyne.Window) {
	editMenu := fyne.NewMenu("Editor")
	preferencesItem := fyne.NewMenuItem("Preferences", func() {
		prefForm := widget.NewForm()
		autoSave = fyne.CurrentApp().Preferences().BoolWithFallback("SceneEditor_AutoSave", true)
		autoSaveCheck := widget.NewCheck("", func(b bool) {
			autoSave = b
			fyne.CurrentApp().Preferences().SetBool("SceneEditor_AutoSave", b)
		})
		autoSaveCheck.Checked = autoSave
		prefForm.Append("Auto Save", autoSaveCheck)
		autoSaveTimeText := fyne.CurrentApp().Preferences().StringWithFallback("SceneEditor_AutoSaveTime", "30s")
		autoSaveTimeEntry := widget.NewEntry()
		autoSaveTimeEntry.SetText(autoSaveTimeText)
		autoSaveTimeEntry.Validator = func(s string) error {
			_, err := parseTime(s)
			return err
		}
		autoSaveTimeEntry.OnChanged = func(s string) {
			valErr := autoSaveTimeEntry.Validate()
			if valErr == nil {
				parsedTime, err := parseTime(s)
				if err == nil {
					autoSaveTime = parsedTime
					autoSaveTimer.Stop()
					autoSaveTimer.Reset(autoSaveTime)
					fyne.CurrentApp().Preferences().SetString("SceneEditor_AutoSaveTime", s)
				}

			}
		}
		prefForm.Append("Auto Save Time", autoSaveTimeEntry)
		iconChangesMade := false
		folderIconCheck := widget.NewCheck("", func(b bool) {
			fyne.CurrentApp().Preferences().SetBool("SceneEditor_FolderIcons", b)
			iconChangesMade = true
		})
		folderIconCheck.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("SceneEditor_FolderIcons", true)
		prefForm.Append("Folder Icons", folderIconCheck)
		sceneIconCheck := widget.NewCheck("", func(b bool) {
			fyne.CurrentApp().Preferences().SetBool("SceneEditor_SceneIcons", b)
			iconChangesMade = true
		})
		sceneIconCheck.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("SceneEditor_SceneIcons", true)
		prefForm.Append("Scene Icons", sceneIconCheck)
		formDialog := dialog.NewCustom("Scene Editor Preferences", "Close", prefForm, window)
		formDialog.SetOnClosed(func() {
			if iconChangesMade {
				dialog.ShowInformation("Restart Required", "Changes to the icon settings require a restart or scene tree refresh to take effect", window)
				iconChangesMade = false
			}
		})
		formDialog.Show()
	})
	editMenu.Items = append(editMenu.Items, preferencesItem)
	saveItem := fyne.NewMenuItem("Save", func() {
		saveScene(window)
	})
	saveItem.Shortcut = NewCustomShortcut("Save", fyne.KeyS, fyne.KeyModifierControl)
	editMenu.Items = append(editMenu.Items, saveItem)
	previewSceneItem := fyne.NewMenuItem("Preview Scene", func() {
		if selectedScene != nil {
			if scenePreviewWindow == nil {
				scenePreviewWindow = fyne.CurrentApp().NewWindow("Scene Preview")
				scenePreviewWindow.SetContent(CreateScenePreview(scenePreviewWindow))
				scenePreviewWindow.Resize(fyne.NewSize(800, 600))
				scenePreviewWindow.Show()
			} else {
				scenePreviewWindow.Show()
			}
		}
	})
	editMenu.Items = append(editMenu.Items, previewSceneItem)
	runGameItem := fyne.NewMenuItem("Run Game", func() {
		//TODO: Add in code to run the game
	})
	editMenu.Items = append(editMenu.Items, runGameItem)
	mainMenu := window.MainMenu()
	//Check if a menu with the name "Editor" exists and if it does replace it
	for i, menu := range mainMenu.Items {
		if menu.Label == "Editor" {
			mainMenu.Items[i] = editMenu
			window.SetMainMenu(mainMenu)
			return
		}
	}
	mainMenu.Items = append(mainMenu.Items, editMenu)
	window.SetMainMenu(mainMenu)
}

func CreateSceneProperties(window fyne.Window) fyne.CanvasObject {
	err := loadAssets(filepath.Join(ActiveProject.Info.Path, "data", "assets"))
	if err != nil {
		dialog.ShowError(err, window)
	}

	//If any are 0 run the export registered functions by creating it via template

	if propertiesCanvas == nil {
		propertiesCanvas = container.NewVBox(widget.NewLabel("Scene Properties"))
	}
	properties := propertiesCanvas.(*fyne.Container)
	if selectedObject == nil {
		//Remove all but the first label
		properties.Objects = properties.Objects[:1]
	} else {
		properties.Objects = properties.Objects[:1]
		typeLabel := widget.NewLabel("Type: ")
		object, ok := selectedObject.(NFObjects.NFObject)
		if !ok {
			panic("Selected Object is not an NFObject")
		}
		typeLabel.SetText("Type: " + object.GetType())
		coupledArgs := NFData.NewCoupledInterfaceMap(object.GetArgs())
		form := widget.NewForm()
		refreshForm("Properties", form, coupledArgs, window)
		properties.Add(typeLabel)
		properties.Add(form)
	}
	propertiesCanvas.Refresh()
	return propertiesCanvas
}

func CreateSceneSelector(window fyne.Window) fyne.CanvasObject {
	//TODO add in an empty folder icon for nodes that are folders but have no children
	if selectorCanvas == nil {
		selectorCanvas = container.NewStack()
	}
	selector := selectorCanvas.(*fyne.Container)
	projectPath := ActiveProject.Info.Path
	//Go to the parent folder of the .NFProject file in the project path
	projectPath = filepath.Dir(projectPath)
	scenesFolder := filepath.Join(projectPath, "data/scenes/")
	scenesFolder = filepath.Clean(scenesFolder)
	if !fs.ValidPath(scenesFolder) {
		return container.NewVBox(widget.NewLabel("Invalid Scenes Folder"))
	}

	err := regenSceneMap(scenesFolder)
	if err != nil {
		dialog.ShowError(err, window)
		return container.NewVBox(widget.NewLabel("Error generating tree map"))
	}

	//Group Buttons
	newSceneButton := func(path string, window fyne.Window) *widget.Button {
		return widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			var newSceneDialog *dialog.CustomDialog
			entry := widget.NewEntry()
			entry.SetPlaceHolder("Scene Name")
			entry.Validator = func(s string) error {
				re := regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)
				if !re.MatchString(s) {
					return errors.New("invalid characters in scene name")
				}
				return nil
			}
			confirmButton := widget.NewButton("Create", func() {
				if entry.Validate() != nil {
					return
				}
				sceneName := entry.Text
				if sceneName == "" {
					return
				}
				scenePath := filepath.Join(path, sceneName+".NFScene")
				log.Println("Creating Scene at " + scenePath)
				newScene := NFScene.New(
					sceneName,
					NFLayout.NewLayout(
						"VBox",
						NFLayout.NewChildren(
							NFWidget.New(
								"Label",
								NFWidget.NewChildren(),
								NFData.NewNFInterfaceMap(
									NFData.NewKeyVal(
										"Text",
										"Hello World",
									),
								),
							),
						),
						NFData.NewNFInterfaceMap(),
					),
					NFData.NewNFInterfaceMap(),
				)
				err := newScene.Save(scenePath)
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				err = regenSceneMap(scenesFolder)
				if err != nil {
					dialog.ShowError(err, window)
				}
				newSceneDialog.Hide()
			})
			cancelButton := widget.NewButton("Cancel", func() {
				newSceneDialog.Hide()
			})

			entry.SetOnValidationChanged(func(e error) {
				if e != nil {
					confirmButton.Disable()
				} else {
					confirmButton.Enable()
				}
			})

			hbox := container.NewHBox(layout.NewSpacer(), cancelButton, confirmButton, layout.NewSpacer())
			content := container.NewVBox(entry, hbox)
			newSceneDialog = dialog.NewCustomWithoutButtons("Create New Scene", content, window)
			newSceneDialog.Show()
		})
	}
	newGroupButton := func(path string, window fyne.Window) *widget.Button {
		return widget.NewButtonWithIcon("", theme.FolderNewIcon(), func() {
			var newGroupDialog *dialog.CustomDialog
			entry := widget.NewEntry()
			entry.SetPlaceHolder("Group Name")
			entry.Validator = func(s string) error {
				re := regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)
				if !re.MatchString(s) {
					return errors.New("invalid characters in group name")
				}
				return nil
			}
			confirmButton := widget.NewButton("Create", func() {
				if entry.Validate() != nil {
					return
				}
				groupName := entry.Text
				if groupName == "" {
					return
				}
				groupPath := filepath.Join(path, groupName)
				log.Println("Creating Group at " + groupPath)
				err := os.MkdirAll(groupPath, os.ModePerm)
				if err != nil {
					return
				}
				err = regenSceneMap(scenesFolder)
				if err != nil {
					dialog.ShowError(err, window)
				}
				newGroupDialog.Hide()
			})
			cancelButton := widget.NewButton("Cancel", func() {
				newGroupDialog.Hide()
			})

			entry.SetOnValidationChanged(func(e error) {
				if e != nil {
					confirmButton.Disable()
				} else {
					confirmButton.Enable()
				}
			})

			hbox := container.NewHBox(layout.NewSpacer(), cancelButton, confirmButton, layout.NewSpacer())
			content := container.NewVBox(entry, hbox)
			newGroupDialog = dialog.NewCustomWithoutButtons("Create New Group", content, window)
			newGroupDialog.Show()
		})
	}
	deleteGroupButton := func(path, group string, window fyne.Window) *widget.Button {
		return widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			dialog.ShowConfirm("Delete "+group, "Are you sure you want to delete the group: "+group, func(b bool) {
				if b {
					log.Println("Delete Group at " + path)
					//Check if the group is empty
					empty := true
					err := filepath.Walk(path, func(innerPath string, info os.FileInfo, err error) error {
						if path != innerPath {
							empty = false
						}
						return nil
					})
					if err != nil {
						dialog.ShowError(err, window)
					}
					if !empty {
						dialog.ShowCustomConfirm("Keep Files", "Keep", "Delete", widget.NewLabel("Do you want to keep the files in the group?"), func(b bool) {
							if b {
								// Create an empty slice of paths to move
								paths := make([]string, 0)

								// Walk the directory and get a list of all paths directly in the directory and add them to the list
								err := filepath.Walk(path, func(innerPath string, info os.FileInfo, err error) error {
									if path != innerPath {
										paths = append(paths, innerPath)
									}
									return nil
								})

								if err != nil {
									dialog.ShowError(err, window)
									return
								}

								// If any folders are found in the list, remove files from inside the folder from the list
								for _, innerPath := range paths {
									if _, err := os.Stat(innerPath); err == nil {
										// Check if the path is a directory
										fileInfo, _ := os.Stat(innerPath)
										if fileInfo.IsDir() {
											// Remove all files from the list that are in the directory
											paths = slices.DeleteFunc(paths, func(s string) bool {
												return strings.HasPrefix(s, innerPath)
											})
											paths = append(paths, innerPath)
										}
									}
								}

								// Move all files in the list to the parent directory
								for _, innerPath := range paths {
									newPath := filepath.Join(filepath.Dir(path), filepath.Base(innerPath))
									//Check if the new path exists and if it does loop adding numbers to the end until it doesn't
									for i := 1; ; i++ {
										if _, err := os.Stat(newPath); os.IsNotExist(err) {
											break
										}
										newPath = filepath.Join(filepath.Dir(path), strings.TrimSuffix(filepath.Base(innerPath), filepath.Ext(filepath.Base(innerPath)))+"_"+strconv.Itoa(i)+filepath.Ext(filepath.Base(innerPath)))
									}
									log.Println("Moving " + innerPath + " to " + newPath)
									err := os.Rename(innerPath, newPath)
									if err != nil {
										dialog.ShowError(err, window)
										return
									}
								}

								if err != nil {
									dialog.ShowError(err, window)
								}
								err = os.RemoveAll(path)
								if err != nil {
									dialog.ShowError(err, window)
								}
							} else {
								err := os.RemoveAll(path)
								if err != nil {
									dialog.ShowError(err, window)
								}
							}
							err = regenSceneMap(scenesFolder)
							if err != nil {
								dialog.ShowError(err, window)
							}
						}, window)
					} else {
						log.Println("Group is empty, deleting group")
						err := os.Remove(path)
						if err != nil {
							dialog.ShowError(err, window)
						} else {
							err = regenSceneMap(scenesFolder)
							if err != nil {
								dialog.ShowError(err, window)
							}
						}
					}
				}
			}, window)
		})
	}
	//Scene Buttons
	copySceneButton := func(path string, window fyne.Window) *widget.Button {
		return widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
			log.Println("Copy Scene at " + path)
			//Copy the scene to the same directory with a _copy suffix
			newPath := strings.TrimSuffix(path, filepath.Ext(path)) + "_copy" + filepath.Ext(path)
			//Copy not Link the file
			fileBytes, err := os.ReadFile(path)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			err = os.WriteFile(newPath, fileBytes, os.ModePerm)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			err = regenSceneMap(scenesFolder)
			if err != nil {
				dialog.ShowError(err, window)
			}
		})
	}
	moveSceneButton := func(path string, window fyne.Window) *widget.Button {
		return widget.NewButtonWithIcon("", theme.ContentCutIcon(), func() {
			openDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				newPath := uri.Path()
				newPath = filepath.Clean(newPath)
				log.Println("Move Scene at " + path + " to " + newPath)
				//Move the scene to the new directory
				newPath = filepath.Join(newPath, filepath.Base(path))

				// Ensure the destination directory exists
				destDir := filepath.Dir(newPath)
				if _, err := os.Stat(destDir); os.IsNotExist(err) {
					err = os.MkdirAll(destDir, os.ModePerm)
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
				}

				// Check if the file already exists in the destination directory
				if _, err := os.Stat(newPath); !os.IsNotExist(err) {
					// Handle the case where the file already exists (e.g., rename, overwrite, or skip)
					// For now, let's skip the move operation
					dialog.ShowInformation("File Exists", "A file with the same name already exists in the destination directory.", window)
					return
				}

				// Move the file
				err = os.Rename(path, newPath)
				if err != nil {
					dialog.ShowError(err, window)
				} else {
					//Check if the old file still exists and if it does delete it
					if _, err := os.Stat(path); !os.IsNotExist(err) {
						err := os.Remove(path)
						if err != nil {
							dialog.ShowError(err, window)
						}
					}
				}
				err = regenSceneMap(scenesFolder)
				if err != nil {
					dialog.ShowError(err, window)
				}
			}, window)
			// Check if the path is a directory
			fileInfo, err := os.Stat(filepath.Dir(path))
			if err != nil {
				log.Println("Failed to get file info")
				return
			}
			if !fileInfo.IsDir() {
				log.Println("Path is not a directory")
				return
			}
			// Convert the path to a fyne.URI
			uri := storage.NewFileURI(filepath.Dir(path))
			URI, err := storage.ListerForURI(uri)
			if err != nil {
				log.Println("Failed to get URI")
				return
			}
			openDialog.SetLocation(URI)
			openDialog.Show()
		})
	}
	deleteSceneButton := func(path, sceneName string, window fyne.Window) *widget.Button {
		return widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			dialog.ShowConfirm("Delete "+sceneName, "Are you sure you want to delete the scene: "+sceneName, func(b bool) {
				if b {
					log.Println("Delete Scene at " + path)
					err := os.Remove(path)
					if err != nil {
						dialog.ShowError(err, window)
					} else {
						err = regenSceneMap(scenesFolder)
						if err != nil {
							dialog.ShowError(err, window)
						}
					}
				}
			}, window)
		})
	}
	var tree *widget.Tree
	tree = widget.NewTreeWithData(sceneTreeBinding,
		func(branch bool) fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("Object"))
		},
		func(item binding.DataItem, b bool, object fyne.CanvasObject) {
			strBind := item.(binding.String)
			value, err := strBind.Get()
			if err != nil {
				log.Println(err)
				return
			}
			id, err := uuid.Parse(value)
			if err != nil {
				log.Println(err)
				return
			}
			path := sceneTreeMap[id]
			base := filepath.Base(path)
			isScene := filepath.Ext(base) == ".NFScene"
			label := strings.TrimSuffix(base, filepath.Ext(base))
			if !isScene {
				folderIconPref := fyne.CurrentApp().Preferences().BoolWithFallback("SceneEditor_FolderIcons", true)
				if folderIconPref {
					//Prepend a folder icon to the label
					label = "📁 " + label
				}
			} else {
				sceneIconPref := fyne.CurrentApp().Preferences().BoolWithFallback("SceneEditor_SceneIcons", true)
				if sceneIconPref {
					//Prepend a scene icon to the label
					label = "🎬 " + label
				}
			}
			objectContainer := object.(*fyne.Container)
			objectContainer.Objects = []fyne.CanvasObject{widget.NewLabel(label)}
			open := tree.IsBranchOpen(value)
			selected := selectedSceneTreeNodeID == value
			if isScene {
				if selected {
					copyButton := copySceneButton(path, window)
					moveButton := moveSceneButton(path, window)
					deleteButton := deleteSceneButton(path, label, window)
					//Disable them all temporarily until I fix them to work with the new tree
					copyButton.Disable()
					moveButton.Disable()
					deleteButton.Disable()
					objectContainer.Objects = append(objectContainer.Objects,
						layout.NewSpacer(),
						copyButton,
						moveButton,
						deleteButton,
					)
				}
			} else {
				if selected || open {
					newGroup := newGroupButton(path, window)
					newScene := newSceneButton(path, window)
					deleteGroup := deleteGroupButton(path, label, window)
					//Disable them all temporarily until I fix them to work with the new tree
					objectContainer.Objects = append(objectContainer.Objects,
						layout.NewSpacer(),
						newGroup,
						newScene,
						deleteGroup,
					)
				}
			}
		})

	tree.OnSelected = func(id widget.TreeNodeID) {
		selectedSceneTreeNodeID = id // This is for the node not actual scene
		parsedID, err := uuid.Parse(id)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		path := sceneTreeMap[parsedID]
		info, err := os.Stat(path)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if info.IsDir() {
			return
		} else {
			log.Println("Selected Scene: " + path)
			// Close the scene preview window if it exists
			if scenePreviewWindow != nil {
				scenePreviewWindow.Close()
			}
			scene, err := NFScene.Load(path)
			if err != nil {
				log.Println(err)
				dialog.ShowError(err, window)
				return
			}
			if !reflect.DeepEqual(selectedScene, scene) {
				//Get the param list out of the sceneObjects if it exists
				if selectedScene != nil && selectedScenePath != "" && changesMade {
					dialog.ShowConfirm("Save before switching", "Do you want to save the current scene before switching?", func(b bool) {
						if b {
							saveScene(window)
						}
						changesMade = false
						selectedScenePath = path
						selectedScene = scene
						CreateScenePreview(window) // This one may not need to be here since it should hit the next one after the else, but it's here for now as a safety
					}, window)
				} else {
					changesMade = false
					selectedScenePath = path
					selectedScene = scene
					CreateScenePreview(window)
				}
			}
		}
	}
	tree.OnUnselected = func(id widget.TreeNodeID) {
		selectedSceneTreeNodeID = ""

	}
	hbox := container.NewHBox(widget.NewLabel("Scenes"), layout.NewSpacer(), newSceneButton(scenesFolder, window), newGroupButton(scenesFolder, window))
	vbox := container.NewVBox(CreateProjectSettings(window), hbox)
	border := container.NewBorder(vbox, nil, nil, nil, tree)
	scroll := container.NewVScroll(border)
	windowSize := window.Canvas().Size()
	scroll.Resize(fyne.NewSize(windowSize.Width/4, windowSize.Height))
	scroll.SetMinSize(fyne.NewSize(300, 0))
	selector.Add(scroll)
	selectorCanvas.Refresh()
	return selectorCanvas
}

func CreateProjectSettings(window fyne.Window) fyne.CanvasObject {
	projectPath := ActiveProject.Info.Path
	projectPath = filepath.Dir(projectPath)
	configPath := filepath.Join(projectPath, "data", "Game.NFConfig")
	configPath = filepath.Clean(configPath)
	if !fs.ValidPath(configPath) {
		return widget.NewLabel("Invalid Project Config Path")
	}
	configFile, err := os.Open(configPath)
	if err != nil {
		log.Println(err)
		return widget.NewLabel("Error Loading Project Config")
	}
	//Parse the config file
	config := NFConfig.NewBlankConfig()
	err = config.Load(configFile)
	if err != nil {
		log.Println(err)
		return widget.NewLabel("Error Loading Project Config")
	}

	form := widget.NewForm()
	form.Append("Move Project", widget.NewButton("Move", func() {
		//TODO: Move the project to a new location
	}))
	nameEntry := widget.NewEntry()
	nameEntry.SetText(config.Name)
	nameEntry.OnChanged = func(s string) {
		config.Name = s
	}
	form.Append("Project Name", nameEntry)
	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetText(config.Credits)
	descriptionEntry.OnChanged = func(s string) {
		config.Credits = s
	}
	form.Append("Project Description", descriptionEntry)
	versionEntry := widget.NewEntry()
	versionEntry.SetText(config.Version)
	versionEntry.OnChanged = func(s string) {
		config.Version = s
	}
	form.Append("Project Version", versionEntry)
	authorEntry := widget.NewEntry()
	authorEntry.SetText(config.Author)
	authorEntry.OnChanged = func(s string) {
		config.Author = s
	}
	form.Append("Project Author", authorEntry)
	saveButton := widget.NewButton("Save", func() {
		err := config.Save(configPath)
		if err != nil {
			dialog.ShowError(err, window)
		}
	})
	form.Append("", saveButton)
	return form

}

func CreateScenePreview(window fyne.Window) fyne.CanvasObject {
	if previewCanvas == nil {
		previewCanvas = container.NewVBox(widget.NewLabel("Scene Preview"))
	}
	if autoSaveTimer == nil {
		autoSaveTimer = time.NewTimer(autoSaveTime)
		go func() {
			for {
				<-autoSaveTimer.C
				if autoSave {
					saveScene(window)
				}
			}
		}()
	} else {
		autoSaveTimer.Stop()
		autoSaveTimer.Reset(autoSaveTime)
	}
	preview := previewCanvas.(*fyne.Container)
	if selectedScene == nil {
		//Remove all but the first label
		// preview.Objects = preview.Objects[:1]
		preview.Objects[0] = container.NewVBox(widget.NewLabel("No Scene Loaded, Select a Scene to Preview"))
	} else {
		err := sceneNameBinding.Set(selectedScene.GetName())
		if err != nil {
			log.Println(err)
			dialog.ShowError(err, window)
		}
		err = updateSceneObjects(nil)
		if err != nil {
			log.Println(err)
			dialog.ShowError(err, window)
		}

		// Put an empty widget in the preview
		preview.Objects = []fyne.CanvasObject{}

		// Parse the current scene
		scene, err := selectedScene.Parse(window)
		if err != nil {
			log.Println(err)
			dialog.ShowError(err, window)
		} else {
			preview.Add(scene)
		}

	}
	previewCanvas.Refresh()
	return previewCanvas
}

//TODO Fix the tree not updating fully

func updateSceneObjects(tree *widget.Tree) error {
	if selectedScene == nil {
		return nil
	}
	childrenMap, count := selectedScene.FetchAll()
	log.Println("Scene has", count, "objects")
	err := selectedScene.Validate()
	if err != nil {
		selectedScene.MakeId()
		return updateSceneObjects(tree)
	}
	sceneObjectData := make(map[string][]string)
	sceneObjectValue := make(map[string]string)
	for parent, children := range childrenMap {
		sceneObjectValue[parent.String()] = parent.String()
		sceneObjectData[parent.String()] = make([]string, 0)
		for _, child := range children {
			sceneObjectValue[child.GetID().String()] = child.GetID().String()
			sceneObjectData[parent.String()] = append(sceneObjectData[parent.String()], child.GetID().String())
		}
	}
	err = objectTreeBinding.Set(sceneObjectData, sceneObjectValue)
	if tree != nil && err == nil {
		tree.OpenAllBranches()
		tree.Refresh()
	}
	return err
}

// TODO Delete makes tree go invisible
// Add Does not appear to work

func CreateSceneObjects(window fyne.Window) fyne.CanvasObject {
	if objectsCanvas == nil {
		objectsCanvas = container.NewStack()
	}
	stack := objectsCanvas.(*fyne.Container)
	sceneObjectsLabel := widget.NewLabelWithData(sceneNameBinding)
	topBox := container.NewHBox(sceneObjectsLabel, layout.NewSpacer())
	if selectedScene == nil {
		err := sceneNameBinding.Set("No Scene Loaded")
		if err != nil {
			log.Println(err)
			dialog.ShowError(err, window)
		}
	}
	var bindingTree *widget.Tree
	previewSceneButton := widget.NewButtonWithIcon("Preview Scene", theme.MediaSkipNextIcon(), func() {
		if selectedScene != nil {
			if scenePreviewWindow == nil {
				scenePreviewWindow = fyne.CurrentApp().NewWindow("Scene Preview")
				scenePreviewWindow.SetContent(previewCanvas)
				scenePreviewWindow.Resize(fyne.NewSize(800, 600))
				scenePreviewWindow.Show()
			} else {
				scenePreviewWindow.Show()
			}
		} else {
			dialog.ShowInformation("No Scene Selected", "You must select a scene to preview", window)
		}
	})
	topBox.Add(previewSceneButton)
	runGameButton := widget.NewButtonWithIcon("Run Game", theme.MediaPlayIcon(), func() {
		//Todo: Add in code to run the game
	})
	topBox.Add(runGameButton)
	refreshObjectsButton := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		err := updateSceneObjects(bindingTree)
		if err != nil {
			log.Println(err)
			dialog.ShowError(err, window)
		}
	})
	topBox.Add(refreshObjectsButton)
	err := updateSceneObjects(bindingTree)
	if err != nil {
		log.Println(err)
		dialog.ShowError(err, window)
	}

	createAddItemButton := func(level int, object NFObjects.NFObject) *widget.Button {
		return widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			if level < 2 {
				return
			}
			var addDialog *dialog.CustomDialog
			rendered := object.(NFObjects.NFRendered)
			createChildButton := widget.NewButton("Add Child", func() {
				children := NFLayout.NewChildren()
				args := NFData.NewNFInterfaceMap()
				if level == 3 {
					if selectedScene.Layout != nil {
						log.Println("Cannot add another layout to the scene")
					} else {
						newLayout := NFLayout.New("VBox", children, args)
						//Loop while the newUUIDStr is in the valueMap
						ids := selectedScene.FetchIDs()
						for slices.Contains(ids, newLayout.GetID()) {
							newLayout = NFLayout.New("VBox", children, args)
						}
						selectedScene.Layout = newLayout
						changesMade = true
					}
				} else {
					args.Set("Text", "Hello World")
					newWidget := NFWidget.New("Label", children, args)
					//Loop while the newUUIDStr is in the valueMap
					ids := selectedScene.FetchIDs()
					for slices.Contains(ids, newWidget.GetID()) {
						newWidget = NFWidget.New("Label", children, args)
					}
					rendered.AddChild(newWidget)
					changesMade = true
				}
				if addDialog != nil {
					addDialog.Hide()
				}
				err = updateSceneObjects(bindingTree)
				if err != nil {
					log.Println(err)
					dialog.ShowError(err, window)
				}
			})
			createFunctionButton := widget.NewButton("Add Action", func() {
				if level != 2 {
					return
				}
				args := NFData.NewNFInterfaceMap()
				newFunction := NFFunction.New("OnParse", "HelloWorld", args)
				//Loop while the newUUIDStr is in the valueMap
				ids := selectedScene.FetchIDs()
				for slices.Contains(ids, newFunction.GetID()) {
					newFunction = NFFunction.New("Function", "HelloWorld", args)
				}
				rendered.AddFunction(newFunction)
				changesMade = true
				if addDialog != nil {
					addDialog.Hide()
				}
				err = updateSceneObjects(bindingTree)
				if err != nil {
					log.Println(err)
					dialog.ShowError(err, window)
				}
			})
			addDialog = dialog.NewCustom("Add Item", "Close", container.NewHBox(createChildButton, createFunctionButton), window)
			addDialog.Show()
		})
	}

	createDeleteItemButton := func(level int, object NFObjects.NFObject) *widget.Button {
		return widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			id := object.GetID()
			if level != 2 || id == selectedScene.Layout.GetID() {
				return
			}
			idStr := id.String()
			vbox := container.NewVBox(widget.NewLabel("Are you sure you want to delete this item?"),
				widget.NewLabel("ID: "+idStr),
				widget.NewLabel("Name: "+object.GetName()),
				widget.NewLabel("Type: "+object.GetType()),
				widget.NewLabel("This action cannot be undone"),
				widget.NewLabel("This will also delete all children and functions of this item"),
				layout.NewSpacer(),
			)
			dialog.ShowCustomConfirm("Delete Item", "Delete", "Cancel", vbox, func(b bool) {
				if b {
					if id == selectedScene.Layout.GetID() {
						return
					} else {
						err := selectedScene.DeleteChild(id, true)
						if err != nil {
							return
						}
					}
					changesMade = true
					CreateSceneObjects(window)
				}
			}, window)
		})
	}

	bindingTree = widget.NewTreeWithData(objectTreeBinding,
		func(bool) fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("Object"))
		},
		func(value binding.DataItem, branch bool, node fyne.CanvasObject) {
			strBinding := value.(binding.String)
			idStr, err := strBinding.Get()
			if err != nil {
				log.Println(err)
				return
			}
			id, err := uuid.Parse(idStr)
			if err != nil {
				return //This means the id is not a uuid and cannot be parsed into an object from the map
			}
			if node == nil {
				return //Something has gone terribly wrong
			}

			object := selectedScene.GetByID(id)
			if object == nil {
				return //This means the object is not in the scene
			}
			level := 0
			switch object.(type) {
			case NFObjects.NFRoot:
				level = 3
			case NFObjects.NFRendered:
				level = 2
			case NFObjects.NFObject:
				level = 1
			}
			if level == 0 {
				return //This means the object is not a valid object
			}
			isOpen := bindingTree.IsBranchOpen(idStr)
			node.(*fyne.Container).Objects = []fyne.CanvasObject{widget.NewLabel(object.GetName())}
			if !(selectedObjectTreeId == idStr || isOpen) {
				return
			}
			addItemButton := createAddItemButton(level, object)
			deleteItemButton := createDeleteItemButton(level, object)
			node.(*fyne.Container).Objects = append(node.(*fyne.Container).Objects, layout.NewSpacer(), addItemButton, deleteItemButton)
			switch level {
			case 1:
				addItemButton.Enable()
				deleteItemButton.Enable()
			case 2:
				addItemButton.Enable()
				if id != selectedScene.Layout.GetID() {
					deleteItemButton.Enable()
				}
			case 3:
				if selectedScene.Layout == nil {
					addItemButton.Enable()
				}
			default:
				deleteItemButton.Disable()
				addItemButton.Disable()
			}
		})
	bindingTree.OnSelected = func(id widget.TreeNodeID) {
		selectedObjectTreeId = id
		bindingTree.RefreshItem(id)
		selectedObject = selectedScene.GetByID(uuid.MustParse(id))
		CreateSceneProperties(window)
	}
	bindingTree.OnUnselected = func(id widget.TreeNodeID) {
		oldId := selectedObjectTreeId
		selectedObjectTreeId = ""
		bindingTree.RefreshItem(oldId)
	}
	bindingTree.OpenAllBranches()
	if selectedObjectTreeId != "" {
		bindingTree.Select(selectedObjectTreeId)
	}
	bindingTree.Refresh()
	border := container.NewBorder(topBox, nil, nil, nil, bindingTree)
	stack.Objects = nil
	stack.Refresh()
	stack.Add(border)
	objectsCanvas.Refresh()
	return objectsCanvas
}

// CreateParamItem creates a FormItem for a parameter
//
// Current Issues: The entries do not properly display validation issues or disable save buttons, but nothing will be saved if the entry is invalid, so it's not a critical issue
func CreateParamItem(key string, val interface{}, parentObject NFData.CoupledObject, parentObjectKey string, parentForm *widget.Form, window fyne.Window) *widget.FormItem {
	valType := NFData.GetValueType(val)
	var coupledObject NFData.CoupledObject
	var formObject fyne.CanvasObject
	switch valType {
	case NFData.FloatType, NFData.IntType, NFData.StringType, NFData.BooleanType:
		entry := widget.NewEntry()
		entry.SetText(fmt.Sprintf("%v", val))
		entry.OnChanged = func(s string) {
			err := entry.Validate()
			if err != nil {
				return
			}
			v, err := parentObject.ParseValue(key, s)
			if err != nil {
				return
			}
			parentObject.Set(key, v)
			changesMade = true
		}
		formObject = entry
	case NFData.PropertyType:
		innerArgs, ok := val.(*NFData.NFInterfaceMap)
		if !ok {
			valMap, ok := val.(map[string]interface{})
			if !ok {
				valMap, ok = val.(NFData.CustomMap)
				if !ok {
					panic("Failed to convert property")
				}
			}
			//Check if it has a Data key
			if _, ok = valMap["Data"]; !ok {
				panic("Property does not have a Data key")
			}
			innerArgs = NFData.NewNFInterfaceFromMap(valMap)
		}
		coupledObject = NFData.NewCoupledInterfaceMap(innerArgs)
	case NFData.SliceType:
		convertedObject, ok := val.([]interface{})
		if !ok {
			convertedObject, ok = val.(NFData.CustomSlice)
			if !ok {
				panic("Failed to convert slice")
			}
		}
		coupledObject = NFData.NewCoupledSlice(convertedObject)
	case NFData.MapType:
		convertedObject, ok := val.(map[string]interface{})
		if !ok {
			convertedObject, ok = val.(NFData.CustomMap)
			if !ok {
				panic("Failed to convert map")
			}
		}
		coupledObject = NFData.NewCoupledMap(convertedObject)
	default:
		formObject = widget.NewLabel(valType.String())
	}

	//All objects need a delete button and a context menu button
	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		waitChan := make(chan struct{})
		parentObject.Delete(key, waitChan, window)
		go func() {
			<-waitChan
			changesMade = true
			refreshForm(parentObjectKey, parentForm, parentObject, window)
		}()
	})
	copyButton := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		parentObject.Copy(key)
		changesMade = true
		refreshForm(parentObjectKey, parentForm, parentObject, window)
	})
	contextButton := widget.NewButtonWithIcon("", theme.MenuIcon(), func() {})
	buttonBox := container.NewHBox(deleteButton, copyButton, contextButton)
	changeTypeSelect := widget.NewSelect(NFData.GetTypesString(), func(s string) {})
	changeTypeSelect.Selected = valType.String()
	changeKeyEntry := widget.NewEntry()
	changeKeyEntry.SetText(key)
	saveKeyButton := widget.NewButton("Apply", func() {})
	saveTypeButton := widget.NewButton("Apply", func() {})
	keyGrid := container.NewGridWithColumns(2, changeKeyEntry, saveKeyButton)
	typeGrid := container.NewGridWithColumns(2, changeTypeSelect, saveTypeButton)
	keyFormItem := widget.NewFormItem("Set Key: ", keyGrid)
	typeFormItem := widget.NewFormItem("Set Type: ", typeGrid)
	closeButton := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {})
	contextForm := widget.NewForm(keyFormItem, typeFormItem)
	contextDialog := dialog.NewCustomWithoutButtons("Context Menu", container.NewVBox(contextForm, layout.NewSpacer(), closeButton), window)
	saveKeyButton.OnTapped = func() {
		newKey := changeKeyEntry.Text
		b := parentObject.SetKey(key, newKey)
		if !b {
			dialog.ShowError(errors.New("failed to change key"), window)
		}
		changesMade = true
		refreshForm(parentObjectKey, parentForm, parentObject, window)
		contextDialog.Hide()
	}
	saveTypeButton.OnTapped = func() {
		newType := NFData.GetType(changeTypeSelect.Selected)
		if newType == valType {
			return
		}
		dialog.ShowConfirm("Change Type", "Are you sure you want to change the type of this item? This **WILL** delete any data it currently has", func(b bool) {
			if b {
				parentObject.SetType(key, newType)
				changesMade = true
				refreshForm(parentObjectKey, parentForm, parentObject, window)
			}
			contextDialog.Hide()
		}, window)
	}
	contextButton.OnTapped = func() {
		contextDialog.Resize(fyne.NewSize(400, 400))
		contextDialog.Show()
	}
	closeButton.OnTapped = func() {
		contextDialog.Hide()
	}

	if coupledObject == nil {
		//If the object is not a map or slice we can just add it to the form with an entry
		gridBox := container.NewGridWithColumns(2, formObject)
		gridBox.Add(buttonBox)
		return widget.NewFormItem(key, gridBox)
	} else {
		//If the object is a map or slice we need to create a new form for it
		innerForm := widget.NewForm()
		refreshForm(key, innerForm, coupledObject, window)
		scrollBox := container.NewVScroll(innerForm)
		scrollBox.SetMinSize(fyne.NewSize(400, 250))
		editDialog := dialog.NewCustomConfirm("Editing "+valType.String(), "Save", "Cancel", container.NewCenter(scrollBox), func(b bool) {
			if b {
				parentObject.Set(key, coupledObject.Object())
				refreshForm(parentObjectKey, parentForm, parentObject, window)
			}
		}, window)
		editButton := widget.NewButton(valType.String(), func() {
			editDialog.Show()
		})
		gridBox := container.NewGridWithColumns(2, editButton, buttonBox)
		return widget.NewFormItem(key, gridBox)
	}
}

func FormButtons(parentKey string, parentForm *widget.Form, parentObject NFData.CoupledObject, window fyne.Window) *widget.FormItem {
	button := widget.NewButtonWithIcon("Add Item...", theme.ContentAddIcon(), func() {
		newKeyInterface := parentObject.Add()
		newKey := ""
		switch v := newKeyInterface.(type) {
		case string:
			newKey = v
		case int:
			newKey = strconv.Itoa(v)
		default:
			dialog.ShowError(errors.New("failed to add new key"), window)
		}
		newVal, b := parentObject.Get(newKey)
		if !b {
			dialog.ShowError(errors.New("failed to add new key"), window)
		}
		parentForm.AppendItem(CreateParamItem(newKey, newVal, parentObject, parentKey, parentForm, window))
		changesMade = true
	})
	buttonBox := container.NewHBox(button, layout.NewSpacer())
	buttonGrid := container.NewGridWithColumns(2, buttonBox)
	return widget.NewFormItem(parentKey, buttonGrid)
}
