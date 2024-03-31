package NFEditor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

/*
TODO: SceneEditor
 [] Scene Selector
	[] Scene List - Grabs all scenes from the project and sorts them based on folders into a tree
 [] Scene Preview - Parses the scene fully using default values for all objects
 [] Scene Other
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
		container.NewVBox(widget.NewLabel("Scene Selector")),
		container.NewVSplit(container.NewVBox(widget.NewLabel("Scene Preview")),
			container.NewHSplit(container.NewVBox(widget.NewLabel("Scene Other")),
				container.NewVBox(widget.NewLabel("Scene Objects")),
			),
		),
	)
	return MainSplit
}
