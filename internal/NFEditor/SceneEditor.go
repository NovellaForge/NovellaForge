package NFEditor

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	"github.com/google/uuid"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

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

type sceneNode struct {
	UUID uuid.UUID
	Name string
	Dir  bool
	Path string
}

var (
	selectedObjectTreeId    = ""
	selectedSceneTreeNodeID = ""
	sceneNameBinding        = binding.NewString()

	functions = make(map[string]NFData.AssetProperties)
	layouts   = make(map[string]NFData.AssetProperties)
	widgets   = make(map[string]NFData.AssetProperties)

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

func generateTreeMap(initialPath string) (map[string][]string, map[string]string, map[string]*sceneNode, error) {
	treeMap := make(map[string]*sceneNode)      // Maps UUID to sceneNode
	parentChildMap := make(map[string][]string) //Maps parent directory to a slice of UUIDs of its children
	valueMap := make(map[string]string)         //Maps UUID to UUID (for value map, it seems redundant, but it's needed for the StringTree)
	pathToUUID := make(map[string]string)       //Maps relative path to UUID

	err := filepath.Walk(initialPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		//Skip the initial path
		if path == initialPath {
			return nil
		}

		dir, file := filepath.Split(path)

		// remove trailing slash from dir
		dir = filepath.Clean(dir)

		if info.IsDir() || filepath.Ext(path) == ".NFScene" {
			newNode := &sceneNode{
				UUID: uuid.New(),
				Name: file,
				Dir:  info.IsDir(),
				Path: path,
			}
			uuidStr := newNode.UUID.String()
			treeMap[uuidStr] = newNode
			pathToUUID[path] = uuidStr // Store the path to UUID mapping

			// Add the new node to the parent's slice in the parent-child map
			if dir == initialPath {
				parentChildMap[""] = append(parentChildMap[""], uuidStr)
			} else {
				parentChildMap[dir] = append(parentChildMap[dir], uuidStr)
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, nil, err
	}

	// Convert the parent-child map to use the UUIDs of the nodes
	idMap := make(map[string][]string)
	for parentPath, childrenUUIDs := range parentChildMap {
		parentID := pathToUUID[parentPath] // Use the path to UUID map here
		idMap[parentID] = childrenUUIDs
	}

	// Create the value map
	for _, node := range treeMap {
		valueMap[node.UUID.String()] = node.UUID.String()
	}

	return idMap, valueMap, treeMap, nil
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
		dialog.ShowCustom("Scene Editor Preferences", "Close", prefForm, window)
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

	functions = make(map[string]NFData.AssetProperties)
	layouts = make(map[string]NFData.AssetProperties)
	widgets = make(map[string]NFData.AssetProperties)
	//Walk the assets folder for all .NFLayout files
	assetsFolder := filepath.Join(filepath.Dir(ActiveProject.Info.Path), "data/assets/")
	assetsFolder = filepath.Clean(assetsFolder)
	err := filepath.Walk(assetsFolder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch filepath.Ext(path) {
		case ".NFLayout":
			l, lErr := NFLayout.Load(path)
			if lErr != nil {
				errors.Join(err, errors.New(fmt.Sprintf("Error loading layout at %s", path)))
			} else {
				if _, ok := layouts[l.Type]; !ok {
					errors.Join(err, errors.New(fmt.Sprintf("Layout Type %s already exists, if you are using third party widgets/layouts the developer has not properly namespaced", l.Type)))
				} else {
					layouts[l.Type] = l
				}
			}
		case ".NFWidget":
			w, wErr := NFWidget.Load(path)
			if wErr != nil {
				errors.Join(err, errors.New(fmt.Sprintf("Error loading widget at %s", path)))
			} else {
				if _, ok := widgets[w.Type]; !ok {
					errors.Join(err, errors.New(fmt.Sprintf("Widget Type %s already exists, if you are using third party widgets/layouts the developer has not properly namespaced", w.Type)))
				} else {
					widgets[w.Type] = w
				}
			}
		case ".NFFunction":
			f, fErr := NFFunction.Load(path)
			if fErr != nil {
				errors.Join(err, errors.New(fmt.Sprintf("Error loading function at %s", path)))
			} else {
				if _, ok := functions[f.Type]; !ok {
					errors.Join(err, errors.New(fmt.Sprintf("Function Type %s already exists, if you are using third party widgets/layouts the developer has not properly namespaced", f.Type)))
				} else {
					functions[f.Type] = f
				}
			}
		}
		return err
	})
	if err != nil {
		log.Println(err)
		dialog.ShowError(err, window)
	}
	propertiesCanvas.Refresh()
	return propertiesCanvas
}

func CreateSceneSelector(window fyne.Window) fyne.CanvasObject {
	//TODO finish converting buttons to new tree
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

	idMap, valueMap, treeMap, err := generateTreeMap(scenesFolder)
	if err != nil {
		dialog.ShowError(err, window)
		return container.NewVBox(widget.NewLabel("Error generating tree map"))
	}

	//Group Buttons
	newSceneButton := func(path, text string, window fyne.Window) *widget.Button {
		return widget.NewButtonWithIcon(text, theme.ContentAddIcon(), func() {
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
				newScene := NFScene.NewScene(
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
				CreateSceneSelector(window)
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
	newGroupButton := func(path, text string, window fyne.Window) *widget.Button {
		return widget.NewButtonWithIcon(text, theme.FolderNewIcon(), func() {
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
				CreateSceneSelector(window)
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
	deleteGroupButton := func(path, text string, window fyne.Window) *widget.Button {
		return widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			dialog.ShowConfirm("Delete "+text, "Are you sure you want to delete the group: "+text, func(b bool) {
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
								//Get all the files in the group and move them to the parent directory
								err := filepath.Walk(path, func(innerPath string, info os.FileInfo, err error) error {
									if path != innerPath {
										newPath := filepath.Join(filepath.Dir(path), info.Name())
										//Check if the new path exists and if it does loop adding numbers to the end until it doesn't
										for i := 1; ; i++ {
											if _, err := os.Stat(newPath); os.IsNotExist(err) {
												break
											}
											newPath = filepath.Join(filepath.Dir(path), strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))+"_"+strconv.Itoa(i)+filepath.Ext(info.Name()))
										}
										err := os.Rename(innerPath, newPath)
										if err != nil {
											return err
										}
									}
									return nil
								})
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
							CreateSceneSelector(window)
						}, window)
					} else {
						log.Println("Group is empty, deleting group")
						err := os.Remove(path)
						if err != nil {
							dialog.ShowError(err, window)
						} else {
							CreateSceneSelector(window)
						}
					}
				}
			}, window)
		})
	}
	//Scene Buttons
	copySceneButton := func(path, _ string, window fyne.Window) *widget.Button {
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
			CreateSceneSelector(window)
		})
	}
	moveSceneButton := func(path, _ string, window fyne.Window) *widget.Button {
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
				CreateSceneSelector(window)
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
	deleteSceneButton := func(path, text string, window fyne.Window) *widget.Button {
		return widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			dialog.ShowConfirm("Delete "+text, "Are you sure you want to delete the scene: "+text, func(b bool) {
				if b {
					log.Println("Delete Scene at " + path)
					err := os.Remove(path)
					if err != nil {
						dialog.ShowError(err, window)
					} else {
						CreateSceneSelector(window)
					}
				}
			}, window)
		})
	}

	createItem := func(branch bool) fyne.CanvasObject {
		return container.NewHBox(widget.NewLabel("Object"))
	}

	treeBinding := binding.BindStringTree(&idMap, &valueMap)
	var tree *widget.Tree
	tree = widget.NewTreeWithData(treeBinding, createItem,
		func(item binding.DataItem, b bool, object fyne.CanvasObject) {
			strBind := item.(binding.String)
			value, err := strBind.Get()
			if err != nil {
				log.Println(err)
				return
			}
			node := treeMap[value]
			if node == nil {
				log.Println("Node not found")
				return
			}
			base := filepath.Base(node.Path)
			isScene := filepath.Ext(base) == ".NFScene"
			noExt := strings.TrimSuffix(base, filepath.Ext(base))
			objectContainer := object.(*fyne.Container)
			objectContainer.Objects = []fyne.CanvasObject{widget.NewLabel(noExt)}
			open := tree.IsBranchOpen(value)
			selected := selectedSceneTreeNodeID == value
			if isScene {
				if selected {
					copyButton := copySceneButton(node.Path, noExt, window)
					moveButton := moveSceneButton(node.Path, noExt, window)
					deleteButton := deleteSceneButton(node.Path, noExt, window)
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
					newGroup := newGroupButton(node.Path, noExt, window)
					newScene := newSceneButton(node.Path, noExt, window)
					deleteGroup := deleteGroupButton(node.Path, noExt, window)
					//Disable them all temporarily until I fix them to work with the new tree
					newGroup.Disable()
					newScene.Disable()
					deleteGroup.Disable()
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
		node := treeMap[id]
		if node == nil {
			log.Println("Node not found")
			return
		}
		if node.Dir {
			return
		} else {
			scenePath := node.Path
			log.Println("Selected Scene: " + scenePath)
			// Close the scene preview window if it exists
			if scenePreviewWindow != nil {
				scenePreviewWindow.Close()
			}
			scene, err := NFScene.Load(scenePath)
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
						selectedScenePath = scenePath
						selectedScene = scene
						CreateScenePreview(window) // This one may not need to be here since it should hit the next one after the else, but it's here for now as a safety
					}, window)
				} else {
					changesMade = false
					selectedScenePath = scenePath
					selectedScene = scene
					CreateScenePreview(window)
				}
			}
		}
	}
	tree.OnUnselected = func(id widget.TreeNodeID) {
		selectedSceneTreeNodeID = ""

	}
	hbox := container.NewHBox(widget.NewLabel("Scenes"), layout.NewSpacer(), newSceneButton(scenesFolder, "", window), newGroupButton(scenesFolder, "", window))
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
		CreateSceneObjects(window)
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

func CreateSceneObjects(window fyne.Window) fyne.CanvasObject {
	if objectsCanvas == nil {
		objectsCanvas = container.NewStack()
	}
	stack := objectsCanvas.(*fyne.Container)
	treeStack := container.NewStack()
	sceneObjectsLabel := widget.NewLabelWithData(sceneNameBinding)
	topBox := container.NewHBox(sceneObjectsLabel, layout.NewSpacer())
	err := sceneNameBinding.Set("No Scene Loaded")
	if err != nil {
		log.Println(err)
		dialog.ShowError(err, window)
		return nil
	}
	if selectedScene != nil {
		err = sceneNameBinding.Set(selectedScene.GetName())
		if err != nil {
			log.Println(err)
			dialog.ShowError(err, window)
		}
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

		childrenMap, count := selectedScene.FetchAll()
		idMap := make(map[string][]string)
		valueMap := make(map[string]string, count)
		//Add the scene id to the ids map tied to the "" key
		idMap[""] = append(idMap[""], selectedScene.GetID().String())
		for parent, children := range childrenMap {
			valueMap[parent.String()] = parent.String()
			idMap[parent.String()] = make([]string, 0)
			for _, child := range children {
				valueMap[child.GetID().String()] = child.GetID().String()
				idMap[parent.String()] = append(idMap[parent.String()], child.GetID().String())
			}
		}
		log.Println(idMap)
		treeBinding := binding.BindStringTree(&idMap, &valueMap)

		createItem := func(bool) fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("Object"))
		}
		var bindingTree *widget.Tree
		bindingTree = widget.NewTreeWithData(treeBinding, createItem, UpdateObjectTreeItem(window, bindingTree, idMap, valueMap, treeBinding))
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
		treeStack.Add(bindingTree)
	}
	border := container.NewBorder(topBox, nil, nil, nil, treeStack)
	stack.Objects = nil
	stack.Refresh()
	stack.Add(border)
	objectsCanvas.Refresh()
	return objectsCanvas
}

//TODO Clean this up and move it back inside the CreateSceneObjects function

func UpdateObjectTreeItem(window fyne.Window, tree *widget.Tree, idMap map[string][]string, valueMap map[string]string, treeBinding binding.StringTree) func(binding.DataItem, bool, fyne.CanvasObject) {
	return func(value binding.DataItem, branch bool, node fyne.CanvasObject) {
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

		addItemButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
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
						newUUIDStr := newLayout.GetID().String()
						//Loop while the newUUIDStr is in the valueMap
						for _, ok := valueMap[newUUIDStr]; ok; _, ok = valueMap[newUUIDStr] {
							newLayout = NFLayout.New("VBox", children, args)
							newUUIDStr = newLayout.GetID().String()
						}
						selectedScene.Layout = newLayout
						changesMade = true
						idMap[idStr] = append(idMap[idStr], newLayout.GetID().String())
						valueMap[newLayout.GetID().String()] = newLayout.GetID().String()
					}
				} else {
					args.Set("Text", "Hello World")
					newWidget := NFWidget.New("Label", children, args)
					//Loop while the newUUIDStr is in the valueMap
					newUUIDStr := newWidget.GetID().String()
					for _, ok := valueMap[newUUIDStr]; ok; _, ok = valueMap[newUUIDStr] {
						newWidget = NFWidget.New("Label", children, args)
						newUUIDStr = newWidget.GetID().String()
					}
					rendered.AddChild(newWidget)
					changesMade = true
					idMap[idStr] = append(idMap[idStr], newWidget.GetID().String())
					valueMap[newWidget.GetID().String()] = newWidget.GetID().String()
				}
				if addDialog != nil {
					addDialog.Hide()
				}
				err := treeBinding.Set(idMap, valueMap)
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
				newUUIDStr := newFunction.GetID().String()
				for _, ok := valueMap[newUUIDStr]; ok; _, ok = valueMap[newUUIDStr] {
					newFunction = NFFunction.New("Function", "HelloWorld", args)
					newUUIDStr = newFunction.GetID().String()
				}
				rendered.AddFunction(newFunction)
				changesMade = true
				idMap[idStr] = append(idMap[idStr], newFunction.GetID().String())
				valueMap[newFunction.GetID().String()] = newFunction.GetID().String()
				if addDialog != nil {
					addDialog.Hide()
				}
				err := treeBinding.Set(idMap, valueMap)
				if err != nil {
					log.Println(err)
					dialog.ShowError(err, window)
				}
			})
			addDialog = dialog.NewCustom("Add Item", "Close", container.NewHBox(createChildButton, createFunctionButton), window)
			addDialog.Show()
		})
		deleteItemButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			if level != 2 || id == selectedScene.Layout.GetID() {
				return
			}
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
						delete(idMap, idStr)
						delete(valueMap, idStr)
					}
					changesMade = true
					err := treeBinding.Set(idMap, valueMap)
					if err != nil {
						log.Println(err)
						dialog.ShowError(err, window)
					}
				}
			}, window)
		})

		node.(*fyne.Container).Objects = []fyne.CanvasObject{widget.NewLabel(object.GetName())}
		addItemButton.Disable()
		deleteItemButton.Disable()
		if level == 2 && selectedObjectTreeId == idStr {
			addItemButton.Enable()
			if id != selectedScene.Layout.GetID() {
				deleteItemButton.Enable()
			}
			node.(*fyne.Container).Objects = append(node.(*fyne.Container).Objects, addItemButton, deleteItemButton)
		} else if level == 1 {
			deleteItemButton.Enable()
			node.(*fyne.Container).Objects = append(node.(*fyne.Container).Objects, deleteItemButton)
		}
	}
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
