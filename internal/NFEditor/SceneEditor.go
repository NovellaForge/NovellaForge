package NFEditor

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/internal/NFEditor/EditorWidgets"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFFunction"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout"
	"go.novellaforge.dev/novellaforge/pkg/NFScene"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
)

/*
TODO: SceneEditor
 [] Project Settings
	[] Project Name
	[] Move Project
	[] Project Icon
	[] Project Description
	[] Project Version
	[] Project Author
	[] Anything else we can think of
	[] To make this easy move the project info to an embedded .NFConfig file in the project folder
 [] Scene Editor
 [] Scene Selector
	[] Scene List - Grabs all scenes from the project and sorts them based on folders into a tree
 [] Scene Preview - Parses the scene fully using default values for all objects
 [] Scene Properties
	[] Lists the scene name and object id of the selected object at the top
	[] Lists all properties of the selected object
	[] Allows for editing of the properties limiting to allowed types/values
 [] Scene Objects
	[] Lists all objects in the scene
	[] Allows for adding/removing objects
	[] Allows for selecting objects to edit properties
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
	emptyData             = struct{}{}
	sceneNodes            = make(map[string]*sceneNode)
	sceneObjects          = make(map[string]*sceneNode)
	sceneListUpdate       = make(chan struct{})
	scenePreviewUpdate    = make(chan struct{})
	sceneObjectsUpdate    = make(chan struct{})
	scenePropertiesUpdate = make(chan struct{})
	propertyTypesUpdate   = make(chan struct{})
	functions             = make(map[string]NFData.AssetProperties)
	layouts               = make(map[string]NFData.AssetProperties)
	widgets               = make(map[string]NFData.AssetProperties)
	selectedScene         *NFScene.Scene
	selectedObject        interface{}

	previewCanvas    fyne.CanvasObject
	propertiesCanvas fyne.CanvasObject
	objectsCanvas    fyne.CanvasObject
)

func CreateSceneEditor(window fyne.Window) fyne.CanvasObject {
	MainSplit := container.NewHSplit(
		CreateSceneSelector(window),
		container.NewVSplit(CreateScenePreview(window),
			container.NewHSplit(CreateSceneProperties(window),
				CreateSceneObjects(),
			),
		),
	)
	MainSplit.Offset = 0.25
	return MainSplit
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
	scenesFolder := filepath.Join(projectPath, "game/data/scenes/")
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

				//Sort both the branch and leaf nodes alphabetically
				for i := 0; i < len(branchNodes); i++ {
					for j := i + 1; j < len(branchNodes); j++ {
						if sceneNodes[branchNodes[i]].Name > sceneNodes[branchNodes[j]].Name {
							branchNodes[i], branchNodes[j] = branchNodes[j], branchNodes[i]
						}
					}
				}
				for i := 0; i < len(leafNodes); i++ {
					for j := i + 1; j < len(leafNodes); j++ {
						if sceneNodes[leafNodes[i]].Name > sceneNodes[leafNodes[j]].Name {
							leafNodes[i], leafNodes[j] = leafNodes[j], leafNodes[i]
						}
					}
				}
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
				//Sort both the branch and leaf nodes alphabetically
				for i := 0; i < len(branchNodes); i++ {
					for j := i + 1; j < len(branchNodes); j++ {
						if sceneNodes[branchNodes[i]].Name > sceneNodes[branchNodes[j]].Name {
							branchNodes[i], branchNodes[j] = branchNodes[j], branchNodes[i]
						}
					}
				}
				for i := 0; i < len(leafNodes); i++ {
					for j := i + 1; j < len(leafNodes); j++ {
						if sceneNodes[leafNodes[i]].Name > sceneNodes[leafNodes[j]].Name {
							leafNodes[i], leafNodes[j] = leafNodes[j], leafNodes[i]
						}
					}
				}
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
				selectedScene = scene
				scenePreviewUpdate <- emptyData
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
	border := container.NewBorder(hbox, nil, nil, nil, tree)
	scroll := container.NewVScroll(border)
	windowSize := window.Canvas().Size()
	scroll.Resize(fyne.NewSize(windowSize.Width/4, windowSize.Height))
	scroll.SetMinSize(fyne.NewSize(300, 0))
	return scroll
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
	go func() {
		for {
			<-scenePreviewUpdate
			preview := previewCanvas.(*fyne.Container)
			if selectedScene == nil {
				//Remove all but the first label
				// preview.Objects = preview.Objects[:1]
				preview.Objects[0] = container.NewVBox(widget.NewLabel("No Scene Loaded, Select a Scene to Preview"))
			} else {
				log.Println("Updating Scene Preview")
				// Put an empty widget in the preview
				preview.Objects = []fyne.CanvasObject{}

				// Parse the current scene
				scene, err := selectedScene.Parse(window)
				if err != nil {
					log.Println(err)
					dialog.ShowError(err, window)
					return
				}
				preview.Add(scene)

			}
			previewCanvas.Refresh()
			sceneObjectsUpdate <- emptyData
		}
	}()

	scenePreviewUpdate <- emptyData
	return previewCanvas
}

func CreateSceneObjects() fyne.CanvasObject {
	// objectsCanvas = container.NewBorder(widget.NewLabel("Scene Objects"), nil, nil, nil, widget.NewLabel("No Scene Loaded"))
	sceneObjectsLabel := container.NewVBox(widget.NewLabel("Scene Objects"))
	objectsCanvas = container.NewBorder(sceneObjectsLabel, nil, nil, nil, widget.NewLabel("No Scene Loaded"))
	go func() {
		for {
			<-sceneObjectsUpdate
			//TODO unselect the current loaded object
			if selectedScene == nil {
				objectsCanvas.(*fyne.Container).Objects[0] = widget.NewLabel("No Scene Loaded")
			} else {
				log.Println("Updating Scene Objects")

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

							//Sort both the branch and leaf nodes alphabetically
							for i := 0; i < len(branchNodes); i++ {
								for j := i + 1; j < len(branchNodes); j++ {
									if sceneObjects[branchNodes[i]].Name > sceneObjects[branchNodes[j]].Name {
										branchNodes[i], branchNodes[j] = branchNodes[j], branchNodes[i]
									}
								}
							}
							for i := 0; i < len(leafNodes); i++ {
								for j := i + 1; j < len(leafNodes); j++ {
									if sceneObjects[leafNodes[i]].Name > sceneObjects[leafNodes[j]].Name {
										leafNodes[i], leafNodes[j] = leafNodes[j], leafNodes[i]
									}
								}
							}
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
								//Sort both the branch and leaf nodes alphabetically
								for i := 0; i < len(branchNodes); i++ {
									for j := i + 1; j < len(branchNodes); j++ {
										if sceneObjects[branchNodes[i]].Name > sceneObjects[branchNodes[j]].Name {
											branchNodes[i], branchNodes[j] = branchNodes[j], branchNodes[i]
										}
									}
								}
								for i := 0; i < len(leafNodes); i++ {
									for j := i + 1; j < len(leafNodes); j++ {
										if sceneObjects[leafNodes[i]].Name > sceneObjects[leafNodes[j]].Name {
											leafNodes[i], leafNodes[j] = leafNodes[j], leafNodes[i]
										}
									}
								}
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
						node.Selected = true
						//TODO add selected object properties to the scene properties
						selectedObject = node.Data
						scenePropertiesUpdate <- emptyData
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
			children[id] = &sceneNode{
				Name:     l.Type,
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
			return
			properties := propertiesCanvas.(*fyne.Container)
			if selectedObject == nil {
				//Remove all but the first label
				properties.Objects = properties.Objects[:1]
			} else {
				log.Println("Updating Scene Properties")
				properties.Objects = properties.Objects[:1]
				form := widget.NewForm()
				//TODO add the properties of the selected object to the form

				switch v := selectedObject.(type) {
				case *NFLayout.Layout:
					for k, _ := range v.Args.Data {
						//TODO FIX THIS
						typedParam := EditorWidgets.NewTypedParameter(k, "", EditorWidgets.String, window)
						form.Append(k, typedParam)
					}

				case *NFWidget.Widget:

				}

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
			assetsFolder := filepath.Join(ActiveProject.Info.Path, "game/assets/")
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
	return container.NewVBox(widget.NewLabel("Scene Properties"))
}
