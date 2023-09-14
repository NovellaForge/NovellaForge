package editor

type Widget struct {
	Name       WidgetNameEnum         `json:"Name"`
	Properties map[string]interface{} `json:"Properties"`
	Children   []*Widget              `json:"Children"`
}

type Layout struct {
	Name       LayoutNameEnum         `json:"Name"`
	Widgets    []*Widget              `json:"Widgets"`
	Properties map[string]interface{} `json:"Properties"`
}

type Scene struct {
	Name       string `json:"Name"`
	Custom     bool   `json:"Custom"`
	Layout     Layout `json:"Layout"` // Custom layouts should be prefixed with "Custom"
	Background string `json:"Background"`
}

type SceneGroup struct {
	Name        string       //Not stored in file, is a directory name
	ChildGroups []SceneGroup //Not stored in file, is a list of scene groups in the directory
	Scenes      []Scene      //Not stored in file, is a list of scenes in the directory
}

type Function struct {
	Name    string   `json:"Name"`
	Inputs  []string `json:"Inputs"`
	Outputs []string `json:"Outputs"`
	Body    string   `json:"Body"`
}

type FunctionGroup struct {
	Name      string     //Not stored in file, is a file name
	Functions []Function //Not stored in file, is a list of functions in the file
}

type Project struct {
	GameName       string          `json:"Game Name"`
	Version        string          `json:"Version"`
	Author         string          `json:"Author"`
	Credits        string          `json:"Credits"`
	CustomLayouts  []Layout        `json:"Custom Layouts"`
	CustomWidgets  []Widget        `json:"Custom Widgets"`
	SceneGroups    []SceneGroup    //Not stored in file, is a list of scene groups in the project found in the scenes directory
	Scenes         []Scene         //Not stored in file, is a list of scenes in the scenes directory that are not in a group
	FunctionGroups []FunctionGroup //Not stored in file, is a list of function files in the project found in the functions directory
}

// LayoutNameEnum Layouts enum for the default layouts
type LayoutNameEnum string

const (
	LayoutVBox         LayoutNameEnum = "VBox"
	LayoutHBox                        = "HBox"
	LayoutGrid                        = "Grid"
	LayoutGridWrap                    = "GridWrap"
	LayoutBorder                      = "Border"
	LayoutForm                        = "Form"
	LayoutCenter                      = "Center"
	LayoutMax                         = "Max"
	LayoutTabContainer                = "TabContainer"
)

type WidgetNameEnum string

const (
	WidgetButton          WidgetNameEnum = "Button"
	WidgetLabel                          = "Label"
	WidgetEntry                          = "Entry"
	WidgetCheck                          = "Check"
	WidgetRadio                          = "Radio"
	WidgetSelect                         = "Select"
	WidgetForm                           = "Form"
	WidgetProgressBar                    = "ProgressBar"
	WidgetToolbar                        = "Toolbar"
	WidgetList                           = "List"
	WidgetTree                           = "Tree"
	WidgetTable                          = "Table"
	ContainerVBox                        = "VBox"
	ContainerHBox                        = "HBox"
	ContainerGrid                        = "Grid"
	ContainerGridWrap                    = "GridWrap"
	ContainerBorder                      = "Border"
	ContainerForm                        = "Form"
	ContainerCenter                      = "Center"
	ContainerMax                         = "Max"
	ContainerTabContainer                = "TabContainer"
)

// MainGameTemplate is the template for the <project-name>.go file in the project directory it should contain a fyne app and window and the main menu set.
const MainGameTemplate = `package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// For information on layouts and widgets see the engine documentation or go the scene editor and click the help button
type LayoutNameEnum string

type LayoutNameEnum string

const (
	LayoutVBox         LayoutNameEnum = "VBox"
	LayoutHBox                        = "HBox"
	LayoutGrid                        = "Grid"
	LayoutGridWrap                    = "GridWrap"
	LayoutBorder                      = "Border"
	LayoutForm                        = "Form"
	LayoutCenter                      = "Center"
	LayoutMax                         = "Max"
	LayoutTabContainer                = "TabContainer"
)

type WidgetNameEnum string

const (
	WidgetButton          WidgetNameEnum = "Button"
	WidgetLabel                          = "Label"
	WidgetEntry                          = "Entry"
	WidgetCheck                          = "Check"
	WidgetRadio                          = "Radio"
	WidgetSelect                         = "Select"
	WidgetForm                           = "Form"
	WidgetProgressBar                    = "ProgressBar"
	WidgetToolbar                        = "Toolbar"
	WidgetList                           = "List"
	WidgetTree                           = "Tree"
	WidgetTable                          = "Table"
	ContainerVBox                        = "VBox"
	ContainerHBox                        = "HBox"
	ContainerGrid                        = "Grid"
	ContainerGridWrap                    = "GridWrap"
	ContainerBorder                      = "Border"
	ContainerForm                        = "Form"
	ContainerCenter                      = "Center"
	ContainerMax                         = "Max"
	ContainerTabContainer                = "TabContainer"
)


func main() {
	a := app.New()
	w := a.NewWindow("Hello World")
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}`

// NewProject initializes a new project with default values
func NewProject(gameName string) Project {
	return Project{
		GameName:      gameName,
		Version:       "1.0",
		Author:        "Anonymous",
		Credits:       "Made with NovellaForge",
		CustomLayouts: []Layout{},
		CustomWidgets: []Widget{},
	}
}
