package editor

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"os"
	"path/filepath"
)

type ProjectInfo struct {
	Name         string `json:"Name"`
	LastOpenDate string `json:"LastOpenDate"`
}

// ReadProjectInfo reads the project info from the project file
func ReadProjectInfo() ([]ProjectInfo, error) {
	//Check if the project file exists
	if _, err := os.Stat("projects/projects.nf"); os.IsNotExist(err) {
		//if it doesn't exist, create it
		err = os.WriteFile("projects/projects.nf", []byte("[]"), 0644)
		if err != nil {
			return []ProjectInfo{}, err
		} else {
			return []ProjectInfo{}, nil
		}
	}

	//if it does exist, read it
	file, err := os.ReadFile("projects/projects.nf")
	if err != nil {
		return []ProjectInfo{}, err
	}

	//unmarshal the json into a slice of structs
	var projects []ProjectInfo
	err = json.Unmarshal(file, &projects)
	if err != nil {
		return []ProjectInfo{}, err
	}

	//return the slice of structs
	return projects, nil
}

// OpenRecentProject opens a project from the recent projects list
func OpenRecentProject(info ProjectInfo, window fyne.Window) error {
	if info.Name == "" {
		return ErrInvalidInput
	}
	return OpenProject(info.Name, false, window)
}

// OpenProject opens a project with the given name and a bool if it is not a name but a path
func OpenProject(name string, isPath bool, window fyne.Window) (err error) {
	if name == "" {
		return ErrInvalidInput
	}
	var file []byte
	if isPath {
		//Check if the path is absolute or relative
		if !filepath.IsAbs(name) {
			name, err = filepath.Abs(name)
			if err != nil {
				return err
			}
		}
		file, err = os.ReadFile(name)
		if err != nil {
			return err
		}
	} else {
		file, err = os.ReadFile("projects/" + name + "/" + name + ".NFProject")
		if err != nil {
			return err
		}
	}
	project, err := DeserializeProject(file)
	if err != nil {
		return err
	}
	return LoadProject(project, window)
}

func DeserializeProject(file []byte) (Project, error) {
	project := Project{}
	err := json.Unmarshal(file, &project)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

// LoadProject takes a deserialized project and loads it into the editor loading the scenes and functions as well
func LoadProject(project Project, window fyne.Window) error {
	return ErrNotImplemented
}
