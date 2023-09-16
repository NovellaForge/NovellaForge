package editor

type Widget struct {
	Name       string                 `json:"Name"`
	Children   []*Widget              `json:"Children"`
	Properties map[string]interface{} `json:"Properties"`
}

type Layout struct {
	Name       string                 `json:"Name"`
	Children   []*Widget              `json:"Widgets"`
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
	GameName string `json:"Game Name"`
	Version  string `json:"Version"`
	Author   string `json:"Author"`
	Credits  string `json:"Credits"`
}

// MainGameTemplate is the template for the <project-name>.go file in the project directory it should contain a fyne app and window and the main menu set.
const MainGameTemplate = ``

//TODO create the default layouts and default functions files both should use Annotations to get the parameters and return values of the functions and the types of the widgets
