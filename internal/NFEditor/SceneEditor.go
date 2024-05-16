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
	"sort"
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
 [ ] Convert the function parser to make it an actual object
     so that the user can choose from added functions and specify the args
*/

type sceneNode struct {
	Name     string
	Leaf     bool
	Parent   string
	Children []string
	FullPath string
	Selected bool
	Data     interface{}
}

var (
	emptyData = struct{}{}

	sceneNodes        = make(map[string]*sceneNode)
	objectTreeBinding = binding.NewStringTree()
	selectedObjectID  = ""
	sceneNameBinding  = binding.NewString()

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

func scanScenesFolder(rootPath string) error {
	sceneNodes = make(map[string]*sceneNode)
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create a new node
		node := &sceneNode{
			Name:     info.Name(),
			Leaf:     !info.IsDir(), // Set Leaf to true if it's a file, false if it's a directory
			FullPath: path,
			Selected: false,
		}

		// Use the full path from the root folder as the id removing leading and trailing slashes and replacing the rest with underscores
		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		id := strings.ReplaceAll(relativePath, "\\", "_")
		if info.IsDir() {
			id = "group_" + id
		} else {
			id = "scene_" + strings.TrimSuffix(id, filepath.Ext(info.Name()))
			node.Name = strings.TrimSuffix(node.Name, filepath.Ext(node.Name))
		}

		// Add the node to its parent's children and set the parent of the node
		parentPath := filepath.Dir(path)
		if rootPath != parentPath {
			relativePath, err = filepath.Rel(rootPath, parentPath)
			if err != nil {
				return err
			}
			parentID := "group_" + strings.ReplaceAll(relativePath, "\\", "_")
			if parentID != "group_.." && parentID != "group_." {
				parentNode, ok := sceneNodes[parentID]
				if !ok {
					// If the parent node does not exist yet, create it
					parentNode = &sceneNode{
						Name: filepath.Base(parentPath),
						Leaf: false,
					}
				}
				parentNode.Children = append(parentNode.Children, id)
				sceneNodes[parentID] = parentNode // Update the parent node in the map
				node.Parent = parentID
			}
		}

		// Add the node to the map if it's not the root directory
		if id != "group_." {
			sceneNodes[id] = node
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
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
		return container.NewVBox(widget.NewLabel("Invalid Scenes Folder Path"))
	}
	err := scanScenesFolder(scenesFolder)
	if err != nil {
		log.Println(err)
		return container.NewVBox(widget.NewLabel("Error scanning scenes folder"))
	}
	var tree *widget.Tree
	tree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				branchNodes := make([]widget.TreeNodeID, 0)
				leafNodes := make([]widget.TreeNodeID, 0)
				for nodeID, node := range sceneNodes {
					if node.Parent == "" {
						if node.Leaf {
							leafNodes = append(leafNodes, nodeID)
						} else {
							branchNodes = append(branchNodes, nodeID)
						}
					}
				}
				sort.Strings(branchNodes)
				sort.Strings(leafNodes)
				nodes := make([]widget.TreeNodeID, 0)
				nodes = append(nodes, branchNodes...)
				nodes = append(nodes, leafNodes...)
				return nodes
			}
			if node, ok := sceneNodes[id]; ok {
				branchNodes := make([]widget.TreeNodeID, 0)
				leafNodes := make([]widget.TreeNodeID, 0)
				for _, childID := range node.Children {
					if sceneNodes[childID].Leaf {
						leafNodes = append(leafNodes, childID)
					} else {
						branchNodes = append(branchNodes, childID)
					}
				}
				sort.Strings(branchNodes)
				sort.Strings(leafNodes)
				nodes := make([]widget.TreeNodeID, 0)
				nodes = append(nodes, branchNodes...)
				nodes = append(nodes, leafNodes...)
				return nodes
			}
			return nil
		},
		func(id widget.TreeNodeID) bool {
			if node, ok := sceneNodes[id]; ok {
				return !node.Leaf
			}
			return true
		},
		func(b bool) fyne.CanvasObject {
			if b {
				hbox := container.NewHBox(
					widget.NewLabel("Group"),
				)
				return hbox
			} else {
				hbox := container.NewHBox(
					widget.NewLabel("Scene"),
				)
				return hbox
			}
		},
		func(id widget.TreeNodeID, b bool, object fyne.CanvasObject) {
			if node, ok := sceneNodes[id]; ok {
				if b {
					open := tree.IsBranchOpen(id)
					if open || node.Selected {
						object.(*fyne.Container).Objects = []fyne.CanvasObject{
							widget.NewLabel(node.Name),
							layout.NewSpacer(),
							CreateNewSceneButton(node.FullPath, "", window),
							CreateNewGroupButton(node.FullPath, "", window),
							CreateNewGroupDeleteButton(node.FullPath, node.Name, window),
						}
					} else {
						object.(*fyne.Container).Objects = []fyne.CanvasObject{
							widget.NewLabel(node.Name),
							layout.NewSpacer(),
						}
					}
				} else {
					if node.Selected {
						object.(*fyne.Container).Objects = []fyne.CanvasObject{
							widget.NewLabel(node.Name),
							layout.NewSpacer(),
							CreateNewCopyButton(node.FullPath, node.Name, window),
							CreateNewMoveButton(node.FullPath, node.Name, window),
							CreateNewDeleteButton(node.FullPath, node.Name, window),
						}
					} else {
						object.(*fyne.Container).Objects = []fyne.CanvasObject{
							widget.NewLabel(node.Name),
							layout.NewSpacer(),
						}
					}
				}
			}
		},
	)
	tree.OnSelected = func(id widget.TreeNodeID) {
		if node, ok := sceneNodes[id]; ok {
			node.Selected = true
			if node.Leaf {
				scenePath := node.FullPath
				log.Println("Selected Scene: " + scenePath)
				// Send empty struct to sceneChangeEvent
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
					}
				}
				CreateScenePreview(window)
			}
		}
	}
	tree.OnUnselected = func(id widget.TreeNodeID) {
		if node, ok := sceneNodes[id]; ok {
			node.Selected = false
		}
	}
	hbox := container.NewHBox(widget.NewLabel("Scenes"), layout.NewSpacer(), CreateNewSceneButton(scenesFolder, "", window), CreateNewGroupButton(scenesFolder, "", window))
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

func CreateNewSceneButton(path, text string, window fyne.Window) *widget.Button {
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

func CreateNewGroupButton(path, text string, window fyne.Window) *widget.Button {
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

func CreateNewGroupDeleteButton(path, text string, window fyne.Window) *widget.Button {
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

func CreateNewDeleteButton(path, text string, window fyne.Window) *widget.Button {
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

func CreateNewCopyButton(path, _ string, window fyne.Window) *widget.Button {
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

func CreateNewMoveButton(path, _ string, window fyne.Window) *widget.Button {
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
		children, count := selectedScene.FetchAll()
		err = CreateObjectNodes(window, children, count)
		if err != nil {
			log.Println(err)
			dialog.ShowError(err, window)
			os.Exit(1)
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

func CreateSceneObjects(window fyne.Window) fyne.CanvasObject {
	if objectsCanvas == nil {
		objectsCanvas = container.NewStack()
	}
	stack := objectsCanvas.(*fyne.Container)
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
	}
	objectTree := widget.NewTreeWithData(objectTreeBinding,
		func(bool) fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Object"),
			)
		},
		func(item binding.DataItem, b bool, object fyne.CanvasObject) {
			log.Println("Creating String Tree Item")
			stringItem, ok := item.(binding.String)
			if !ok {
				log.Println("Item is not a string")
				return
			}
			itemID, gErr := stringItem.Get()
			if gErr != nil {
				log.Println(gErr)
				return
			}
			itemUUID, uErr := uuid.Parse(itemID)
			if uErr != nil {
				//Log the error and return
				log.Println(uErr)
				return
			}
			itemObject := selectedScene.GetByID(itemUUID)
			if itemObject == nil {
				//Log the error and return
				log.Println("Object not found")
				return
			}
			itemName := itemObject.GetName()
			object.(*fyne.Container).Objects = []fyne.CanvasObject{
				widget.NewLabel(itemName),
			}
			createChildButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
				newWidget := NFWidget.New("Null", NFWidget.NewChildren(), NFData.NewNFInterfaceMap())
				err := objectTreeBinding.Append(itemID, newWidget.GetID().String(), newWidget.GetName())
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
			})
			deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				err := selectedScene.DeleteChild(itemUUID, true)
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				children, count := selectedScene.FetchAll()
				err = CreateObjectNodes(window, children, count)
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
			})
			if b {
				if itemID == selectedObjectID && selectedScene.GetID() != itemUUID {
					object.(*fyne.Container).Objects = []fyne.CanvasObject{
						widget.NewLabel(itemName),
						createChildButton,
						deleteButton,
						layout.NewSpacer(),
					}
				}
			} else {
				if itemID == selectedObjectID && selectedScene.GetID() != itemUUID {
					object.(*fyne.Container).Objects = []fyne.CanvasObject{
						widget.NewLabel(itemID),
						createChildButton,
						deleteButton,
						layout.NewSpacer(),
					}
				}
			}
		},
	)
	objectTree.OnSelected = func(id widget.TreeNodeID) {
		selectedObjectID = id
	}
	objectTree.Refresh()
	border := container.NewBorder(topBox, nil, nil, nil, objectTree)
	stack.Objects = nil
	stack.Refresh()
	stack.Objects = append(stack.Objects, border)
	stack.Refresh()
	return objectsCanvas
}

func CreateObjectNodes(window fyne.Window, children map[uuid.UUID][]NFObjects.NFObject, count int) error {
	ids := make(map[string][]string)
	v := make(map[string]string)
	log.Println("Creating Object Nodes from " + strconv.Itoa(count) + " children")
	for parent, childSlice := range children {
		for _, child := range childSlice {
			childID := child.GetID()
			ids[parent.String()] = append(ids[parent.String()], childID.String())
			//Check if the v map already has the childID and if it does panic
			if _, ok := v[childID.String()]; ok {
				log.Println(children)
				//Log the duplicate ID and return an error
				log.Println("Duplicate ID found: " + childID.String()) //TODO Need to fix the duplicate ID issue so that scene loading fixes itself
				return errors.New("duplicate ID found")
			}
			v[childID.String()] = childID.String()
		}
		if parent == selectedScene.GetID() {
			log.Println("Scene found")
			ids[""] = append(ids[""], parent.String())
		}
	}
	err := objectTreeBinding.Set(ids, v)
	if err != nil {
		return err
	}
	CreateSceneObjects(window)
	return nil
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
