package NFEditor

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout"
	"go.novellaforge.dev/novellaforge/pkg/NFScene"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
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

func CreateSceneEditor(window fyne.Window) fyne.CanvasObject {
	MainSplit := container.NewHSplit(
		CreateSceneSelector(window),
		container.NewVSplit(CreateScenePreview(),
			container.NewHSplit(CreateSceneProperties(),
				CreateSceneObjects(),
			),
		),
	)
	MainSplit.Offset = 0.25
	return MainSplit
}

type sceneNode struct {
	Name     string
	Leaf     bool
	Parent   string
	Children []string
	FullPath string
	Selected bool
	Opened   bool
}

var sceneNodes = make(map[string]*sceneNode)
var sceneListUpdate = make(chan struct{}, 10) //Buffered to prevent blocking (10 updates can be queued up before blocking

func scanScenesFolder(rootPath string) error {
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
			Opened:   false,
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
			sceneListUpdate <- struct{}{}
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
		newSceneDialog = dialog.NewCustom("Create New Scene", "Create", content, window)
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
			sceneListUpdate <- struct{}{}
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
		newGroupDialog = dialog.NewCustom("Create New Group", "Create", content, window)
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
						sceneListUpdate <- struct{}{}
					}, window)
				} else {
					log.Println("Group is empty, deleting group")
					err := os.Remove(path)
					if err != nil {
						dialog.ShowError(err, window)
					} else {
						sceneListUpdate <- struct{}{}
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
					sceneListUpdate <- struct{}{}
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
		sceneListUpdate <- struct{}{}
	})
}

func CreateNewMoveButton(path, _ string, window fyne.Window) *widget.Button {
	return widget.NewButtonWithIcon("", theme.ContentCutIcon(), func() {
		log.Println("Move Scene at " + path)
		openDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			newPath := uri.Path()
			log.Println("Move Scene at " + path + " to " + newPath)
			//Move the scene to the new directory
			newPath = filepath.Join(newPath, filepath.Base(path))
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
			sceneListUpdate <- struct{}{}
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
					if node.Opened != open {
						if open {
							tree.CloseBranch(id)
						} else {
							tree.OpenBranch(id)
						}
					}
					if node.Opened || node.Selected {
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
		}
	}

	tree.OnUnselected = func(id widget.TreeNodeID) {
		if node, ok := sceneNodes[id]; ok {
			node.Selected = false
		}
	}

	tree.OnBranchOpened = func(id widget.TreeNodeID) {
		if node, ok := sceneNodes[id]; ok {
			node.Opened = true
		}
	}

	tree.OnBranchClosed = func(id widget.TreeNodeID) {
		if node, ok := sceneNodes[id]; ok {
			node.Opened = false
		}
	}

	go func() {
		timer := time.NewTimer(time.Minute)
		for {
			select {
			case <-timer.C:
			case <-sceneListUpdate:
			}
			err := scanScenesFolder(scenesFolder)
			if err != nil {
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

func CreateScenePreview() fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel("Scene Preview"))
}

func CreateSceneProperties() fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel("Scene Properties"))
}

func CreateSceneObjects() fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel("Scene Objects"))
}
