package editor

import (
	"encoding/json"
	"os"
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
