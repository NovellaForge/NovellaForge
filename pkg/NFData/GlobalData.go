package NFData

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var GlobalVars = NewRemoteNFInterfaceMap()
var GlobalBindings = NewNFBindingMap()
var ActiveSceneData *NFSceneData

type AssetProperties struct {
	Type         string              `json:"Type"`
	RequiredArgs map[string][]string `json:"RequiredArgs"`
	OptionalArgs map[string][]string `json:"OptionalArgs"`
}

func (a *AssetProperties) Load(path string) error {
	//Check if the path is a file
	if filepath.Ext(path) == "" {
		return errors.New("invalid file type")
	}
	//Read the file
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	//Unmarshal the json file
	err = json.Unmarshal(file, a)
	if err != nil {
		return err
	}
	return nil
}
