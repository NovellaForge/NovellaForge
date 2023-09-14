package editor

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"path/filepath"
)

var editorVersion = "0.0.1"
var WindowTitle = "Novella Forge" + " V" + editorVersion

type ConfigType struct {
	Version     string             `toml:"version"`
	Resolution  map[string]float32 `toml:"resolution"`
	Fullscreen  bool               `toml:"fullscreen"`
	LastProject string             `toml:"last_project"`
}

var Config ConfigType

// NewConfig creates a new configuration struct
func NewConfig(version string, resolution map[string]float32, fullscreen bool) ConfigType {
	return ConfigType{
		Version:    version,
		Resolution: resolution,
		Fullscreen: fullscreen,
	}
}

// SaveConfig saves the configuration to a TOML file
func SaveConfig(cfg ConfigType) error {
	configDir := "configs"
	configFile := filepath.Join(configDir, "config.toml")

	// Create or open the file
	file, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new TOML encoder
	encoder := toml.NewEncoder(file)

	// Encode the config struct to TOML and write it to the file
	if err := encoder.Encode(cfg); err != nil {
		return err
	}

	return nil
}

// LoadConfig loads the configuration file for the editor or game
func LoadConfig(version string) (ConfigType, error) {
	configDir := "configs"
	configFile := filepath.Join(configDir, "config.toml")
	var cfg ConfigType

	// Check if configs directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		log.Printf("Directory %s does not exist. Creating it.", configDir)
		err := os.Mkdir(configDir, 0755)
		if err != nil {
			return cfg, fmt.Errorf("failed to create directory: %v", err)
		}
	}

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Printf("Config file %s does not exist. Creating it with default values.", configFile)
		err := SaveConfig(NewConfig(version, map[string]float32{"width": 1280, "height": 720}, false))
		if err != nil {
			return cfg, fmt.Errorf("failed to create default config: %v", err)
		}
	}

	// Load config
	if _, err := toml.DecodeFile(configFile, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to load config: %v", err)
	}

	// Perform any additional checks or operations on the loaded config here,
	// For example, checking the version
	if cfg.Version != version {
		log.Printf("Warning: Config version mismatch, checking for missing values or unneeded values. Expected %s, got %s", version, cfg.Version)
		//TODO: Check for missing values or unneeded values
	}

	return cfg, nil
}
