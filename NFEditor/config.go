package NFEditor

const (
	Version = "0.0.1"
	Icon    = "assets/icons/editor.png"
	Author  = "The Novella Forge Team"
)

var WindowTitle = "Novella Forge" + " " + Version

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
