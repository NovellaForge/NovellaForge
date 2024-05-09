package NFEditor

import (
	"errors"
	"fmt"
	"go.novellaforge.dev/novellaforge/pkg/NFConfig"
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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFFunction"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout"
	"go.novellaforge.dev/novellaforge/pkg/NFScene"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
)

/*
TODO: SceneEditor
 Need to add in mutexes for all the channel updates !!! Important !!!
 [X] Project Settings
	[X] Project Name
	[X] Move Project
	[X] Project Description
	[X] Project Version
	[X] Project Author
	[] Anything else we can think of
	[X] To make this easy move the project info to an embedded .NFConfig file in the project folder
 [] Scene Editor
 	[] Scene Saving on Key Press and via Button and Auto Save
		[X] Auto Save
		[X] Save Button
		[O] Save on Key Press (Not Implemented I could not get fyne key press events to work I will ask in the fyne discord soon)
 	[X] Scene Selector
		[X] Scene List - Grabs all scenes from the project and sorts them based on folders into a tree
		[-] Scene Preview - Parses the scene fully using default values for all objects (Basic is done but could be extended to allow better control)
	[] Scene Properties
		[] Lists the scene name and object id of the selected object at the top
		[X] Lists all properties of the selected object
		[X] Allows for editing of the properties limiting to allowed types/values
	[X] Scene Objects
		[X] Lists all objects in the scene
		[X] Allows for adding/removing objects
		[X] Allows for selecting objects to edit properties
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

	sceneNodes   = make(map[string]*sceneNode)
	sceneObjects = make(map[string]*sceneNode)

	sceneListUpdate       = make(chan struct{})
	scenePreviewUpdate    = make(chan struct{})
	sceneObjectsUpdate    = make(chan struct{})
	scenePropertiesUpdate = make(chan struct{})
	propertyTypesUpdate   = make(chan struct{})

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

	previewCanvas    fyne.CanvasObject
	propertiesCanvas fyne.CanvasObject
	objectsCanvas    fyne.CanvasObject
)

func CreateSceneEditor(window fyne.Window) fyne.CanvasObject {
	updateMainMenuBar(window)
	propertyScroll := container.NewScroll(CreateSceneProperties(window))
	propertyScroll.SetMinSize(fyne.NewSize(300, 0))
	MainSplit := container.NewHSplit(
		CreateSceneSelector(window),
		container.NewVSplit(CreateScenePreview(window),
			container.NewHSplit(propertyScroll,
				CreateSceneObjects(window),
			),
		),
	)
	MainSplit.Offset = 0.25
	return MainSplit
}

func updateMainMenuBar(window fyne.Window) {
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
	mainMenu := window.MainMenu()
	mainMenu.Items = append(mainMenu.Items, editMenu)
	window.SetMainMenu(mainMenu)
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
	log.Println("Check Save Scene")
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
			sceneListUpdate <- emptyData
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
			sceneListUpdate <- emptyData
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
						sceneListUpdate <- emptyData
					}, window)
				} else {
					log.Println("Group is empty, deleting group")
					err := os.Remove(path)
					if err != nil {
						dialog.ShowError(err, window)
					} else {
						sceneListUpdate <- emptyData
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
					sceneListUpdate <- emptyData
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
		sceneListUpdate <- emptyData
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
			sceneListUpdate <- emptyData
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

func CreateSceneSelector(window fyne.Window) fyne.CanvasObject {
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
							scenePreviewUpdate <- emptyData
						}, window)
					} else {
						changesMade = false
						selectedScenePath = scenePath
						selectedScene = scene
						scenePreviewUpdate <- emptyData
					}
				}
			}
		}
	}

	tree.OnUnselected = func(id widget.TreeNodeID) {
		if node, ok := sceneNodes[id]; ok {
			node.Selected = false
		}
	}

	go func() {
		for {
			<-sceneListUpdate
			err := scanScenesFolder(scenesFolder)
			if err != nil {
				log.Println(err)
				dialog.ShowError(err, window)
				return
			}
			tree.Refresh()
		}
	}()
	hbox := container.NewHBox(widget.NewLabel("Scenes"), layout.NewSpacer(), CreateNewSceneButton(scenesFolder, "", window), CreateNewGroupButton(scenesFolder, "", window))
	vbox := container.NewVBox(CreateProjectSettings(window), hbox)
	border := container.NewBorder(vbox, nil, nil, nil, tree)
	scroll := container.NewVScroll(border)
	windowSize := window.Canvas().Size()
	scroll.Resize(fyne.NewSize(windowSize.Width/4, windowSize.Height))
	scroll.SetMinSize(fyne.NewSize(300, 0))
	return scroll
}

func CreateProjectSettings(window fyne.Window) fyne.CanvasObject {
	projectPath := ActiveProject.Info.Path
	projectPath = filepath.Dir(projectPath)
	configPath := filepath.Join(projectPath, "data", ".NFConfig")
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

func countChildren(n interface{}) int {
	//Go all the way down the tree and count the children
	count := 1
	var l *NFLayout.Layout
	var w *NFWidget.Widget
	switch v := n.(type) {
	case *NFLayout.Layout:
		l = v
	case *NFWidget.Widget:
		w = v
	}

	if l != nil {
		for _, child := range l.Children {
			count += countChildren(child)
		}
	} else if w != nil {
		for _, child := range w.Children {
			count += countChildren(child)
		}
	}
	return count
}

func CreateScenePreview(window fyne.Window) fyne.CanvasObject {
	previewCanvas = container.NewVBox(widget.NewLabel("Scene Preview"))
	autoSaveTimer = time.NewTimer(autoSaveTime)
	go func() {
		for {
			<-autoSaveTimer.C
			if autoSave {
				saveScene(window)
			}
		}
	}()
	go func() {
		for {
			<-scenePreviewUpdate
			//Reset the auto save timer
			autoSaveTimer.Stop()
			autoSaveTimer.Reset(autoSaveTime)
			preview := previewCanvas.(*fyne.Container)
			if selectedScene == nil {
				//Remove all but the first label
				// preview.Objects = preview.Objects[:1]
				preview.Objects[0] = container.NewVBox(widget.NewLabel("No Scene Loaded, Select a Scene to Preview"))
			} else {
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
			sceneObjectsUpdate <- emptyData
		}
	}()

	scenePreviewUpdate <- emptyData
	return previewCanvas
}

func CreateSceneObjects(window fyne.Window) fyne.CanvasObject {
	sceneObjectsLabel := container.NewVBox(widget.NewLabel("Scene Objects"))
	objectsCanvas = container.NewBorder(sceneObjectsLabel, nil, nil, nil, widget.NewLabel("No Scene Loaded"))
	go func() {
		for {
			<-sceneObjectsUpdate
			if selectedScene == nil {
				objectsCanvas.(*fyne.Container).Objects[0] = widget.NewLabel("No Scene Loaded")
			} else {
				// Get scene info
				layoutType := selectedScene.Layout.Type
				layoutChildrenCount := countChildren(selectedScene.Layout)
				layoutTypeLabel := widget.NewLabel("Layout Type: " + layoutType)
				layoutChildrenCountLabel := widget.NewLabel(strconv.Itoa(layoutChildrenCount) + " Scene Objects:")

				sceneObjectsLabel.Objects = []fyne.CanvasObject{layoutTypeLabel, layoutChildrenCountLabel}

				//Iterate over the scene objects and add them to the list
				sceneObjects = fetchChildren(selectedScene.Layout)
				var tree *widget.Tree
				tree = widget.NewTree(
					func(id widget.TreeNodeID) []widget.TreeNodeID {
						if id == "" {
							branchNodes := make([]widget.TreeNodeID, 0)
							leafNodes := make([]widget.TreeNodeID, 0)
							for nodeID, node := range sceneObjects {
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
						} else {
							//Get the children of the node
							if node, ok := sceneObjects[id]; ok {
								branchNodes := make([]widget.TreeNodeID, 0)
								leafNodes := make([]widget.TreeNodeID, 0)
								for _, childID := range node.Children {
									if sceneObjects[childID].Leaf {
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
						}
						return []string{}
					},
					func(id widget.TreeNodeID) bool {
						if node, ok := sceneObjects[id]; ok {
							return !node.Leaf
						}
						return true
					},
					func(b bool) fyne.CanvasObject {
						if b {
							hbox := container.NewHBox(
								widget.NewLabel("Group"),
								widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {}),
							)
							return hbox
						} else {
							hbox := container.NewHBox(
								widget.NewLabel("Object"),
							)
							return hbox
						}
					},
					func(id widget.TreeNodeID, b bool, object fyne.CanvasObject) {
						if node, ok := sceneObjects[id]; ok {
							if b {
								open := tree.IsBranchOpen(id)
								if open || node.Selected {
									object.(*fyne.Container).Objects = []fyne.CanvasObject{
										widget.NewLabel(node.Name),
										widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
											//Add a new object to the scene
											//Get the parent of the node
											parentNode, ok := sceneObjects[node.Parent]
											if !ok {
												log.Println("Parent Node not found")
												return
											}
											//Get the parent object
											var parentObject interface{}
											if parentNode.Data != nil {
												parentObject = parentNode.Data
											} else {
												log.Println("Parent Object not found")
												return
											}
											//Check if the parent object is a layout or a widget

											//Check if it is a widget or a layout
											if obj, ok := parentObject.(NFData.NFObject); ok {
												//Add a new child widget
												//Create an ID that doesn't exist in the parent
												newID := obj.GetType()
												count := 0
												for {
													innerID := newID + "_" + strconv.Itoa(count)
													_, ok := sceneObjects[innerID]
													if !ok {
														newID = innerID
														break
													}
													count++
												}
												newObject := NFWidget.NewWithID(newID, "Null", NFWidget.NewChildren(), NFData.NewNFInterfaceMap())
												//Add the new object to the parent
												obj.AddChild(newObject)
											} else {
												log.Println("Parent Object is not an NFObject")
												return
											}
										}),
										layout.NewSpacer(),
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
					if node, ok := sceneObjects[id]; ok {
						if !node.Selected {
							node.Selected = true
							selectedObject = node.Data
							scenePropertiesUpdate <- emptyData
						}
					}
				}

				tree.OnUnselected = func(id widget.TreeNodeID) {
					if node, ok := sceneObjects[id]; ok {
						node.Selected = false
					}
				}
				objectsCanvas.(*fyne.Container).Objects[0] = tree
			}
			objectsCanvas.Refresh()
			selectedObject = nil
			scenePropertiesUpdate <- emptyData
		}
	}()
	sceneObjectsUpdate <- emptyData
	return objectsCanvas
}

func fetchChildren(n interface{}, parent ...string) map[string]*sceneNode {
	children := make(map[string]*sceneNode)
	var l *NFLayout.Layout
	var w *NFWidget.Widget
	switch v := n.(type) {
	case *NFLayout.Layout:
		l = v
	case *NFWidget.Widget:
		w = v
	}
	if l != nil {
		children["MainLayout"] = &sceneNode{
			Name:     "MainLayout",
			Leaf:     false,
			Parent:   "",
			Children: nil,
			FullPath: "",
			Selected: false,
			Data:     l,
		}
		for i, child := range l.Children {
			id := l.Type + "_" + strconv.Itoa(i)
			childID, _ := child.GetInfo()
			children[id] = &sceneNode{
				Name:     childID,
				Leaf:     false,
				Parent:   "MainLayout",
				Children: nil,
				FullPath: "",
				Selected: false,
				Data:     child,
			}
			if len(child.Children) > 0 {
				childChildren := fetchChildren(child, id)
				for k, v := range childChildren {
					children[id].Children = append(children[id].Children, k)
					children[k] = v
				}
			} else {
				children[id].Leaf = true
			}
			children["MainLayout"].Children = append(children["MainLayout"].Children, id)
		}
	} else if w != nil {
		for i, child := range w.Children {
			if len(parent) < 1 {
				log.Println("Parent ID not found")
				return nil
			}
			id := parent[0] + "_" + strconv.Itoa(i)
			children[id] = &sceneNode{
				Name:     w.ID,
				Leaf:     false,
				Parent:   parent[0],
				Children: nil,
				FullPath: "",
				Selected: false,
				Data:     w,
			}
			if len(w.Children) > 0 {
				childChildren := fetchChildren(child)
				for k, v := range childChildren {
					children[id].Children = append(children[id].Children, k)
					children[k] = v
				}
			} else {
				//Set leaf to true
				children[id].Leaf = true
			}
		}
	}
	return children
}

func CreateSceneProperties(window fyne.Window) fyne.CanvasObject {
	propertiesCanvas = container.NewVBox(widget.NewLabel("Scene Properties"))
	go func() {
		for {
			<-scenePropertiesUpdate
			properties := propertiesCanvas.(*fyne.Container)
			if selectedObject == nil {
				//Remove all but the first label
				properties.Objects = properties.Objects[:1]
			} else {
				properties.Objects = properties.Objects[:1]
				typeLabel := widget.NewLabel("Type: ")
				object, ok := selectedObject.(NFData.NFObject)
				if !ok {
					panic("Selected Object is not an NFObject")
				}
				typeLabel.SetText("Type: " + object.GetType())
				coupledArgs := NFData.NewCoupledInterfaceMap(object.GetArgs())
				form := widget.NewForm()
				RefreshForm("Properties", form, coupledArgs, window)
				properties.Add(typeLabel)
				properties.Add(form)
			}
			propertiesCanvas.Refresh()
		}
	}()
	go func() {
		for {
			<-propertyTypesUpdate
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
				return
			}
		}
	}()

	propertyTypesUpdate <- emptyData
	scenePropertiesUpdate <- emptyData
	return propertiesCanvas
}

func RefreshForm(objectKey string, form *widget.Form, object NFData.CoupledObject, window fyne.Window) {
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
			RefreshForm(parentObjectKey, parentForm, parentObject, window)
		}()
	})
	copyButton := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		parentObject.Copy(key)
		changesMade = true
		RefreshForm(parentObjectKey, parentForm, parentObject, window)
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
		RefreshForm(parentObjectKey, parentForm, parentObject, window)
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
				RefreshForm(parentObjectKey, parentForm, parentObject, window)
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
		RefreshForm(key, innerForm, coupledObject, window)
		scrollBox := container.NewVScroll(innerForm)
		scrollBox.SetMinSize(fyne.NewSize(400, 250))
		editDialog := dialog.NewCustomConfirm("Editing "+valType.String(), "Save", "Cancel", container.NewCenter(scrollBox), func(b bool) {
			if b {
				parentObject.Set(key, coupledObject.Object())
				RefreshForm(parentObjectKey, parentForm, parentObject, window)
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
