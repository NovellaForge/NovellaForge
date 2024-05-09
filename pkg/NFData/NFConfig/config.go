package NFConfig

import (
	"encoding/json"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFError"
	"io"
	"io/fs"
	"os"
)

// Game is the config for the currently running game if any
var Game = NewBlankConfig()

// NFConfig is the struct that holds all the information about a config
type NFConfig struct {
	// Name of the project
	Name string `json:"Name"`
	// Author of the project
	Author string `json:"Author"`
	// Version of the project
	Version string `json:"Version"`
	// Credits for the project
	Credits string `json:"Credits"`
	// Webpage for the project
	Webpage string `json:"Webpage"`
	// Icon for the project
	Icon string `json:"Icon"`
	//Encryption key for the project
	EncryptionKey string `json:"EncryptionKey"`
}

// NewConfig creates a new config with the given name, author, version and credits
func NewConfig(name, version, author, credits string) *NFConfig {
	return &NFConfig{
		Name:    name,
		Version: version,
		Author:  author,
		Credits: credits,
	}
}

// NewBlankConfig creates a new blank config
func NewBlankConfig() *NFConfig {
	return &NFConfig{}
}

// Load loads the config from the given file
func (c *NFConfig) Load(file fs.File) error {
	defer file.Close()
	//Read the file
	data, err := io.ReadAll(file)
	if err != nil {
		return NFError.NewErrConfigLoad(err.Error())
	}

	//Unmarshal the data
	err = json.Unmarshal(data, c)
	if err != nil {
		return NFError.NewErrConfigLoad(err.Error())
	}

	return nil
}

// Save saves the config to the given file
//
// This is an editor only function and should not be used in the final game
func (c *NFConfig) Save(path string) error {
	//Marshal the data
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return NFError.NewErrConfigSave(err.Error())
	}

	//Write the data to the file
	err = os.WriteFile(path, data, os.ModePerm)
	if err != nil {
		return NFError.NewErrConfigSave(err.Error())
	}
	return nil
}
