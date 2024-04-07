package NFConfig

import (
	"encoding/json"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
	"io"
	"io/fs"
	"os"
)

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
	// UseEmbeddedAssets determines if the project should use embedded assets
	UseEmbeddedAssets bool `json:"UseEmbeddedAssets"`
	// UseEmbeddedData determines if the project should use embedded data
	UseEmbeddedData bool `json:"UseEmbeddedData"`
}

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

// NewConfig creates a new config with the given name, author, version, credits, webpage, icon, and use embedded assets and data
func NewConfig(name, version, author, credits string) *NFConfig {
	return &NFConfig{
		Name:              name,
		Version:           version,
		Author:            author,
		Credits:           credits,
		Webpage:           "",
		Icon:              "",
		UseEmbeddedAssets: false,
		UseEmbeddedData:   false,
	}
}
