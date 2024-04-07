package NFConfig

type Config struct {
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

// NewConfig creates a new config with the given name, author, version, credits, webpage, icon, and use embedded assets and data
func NewConfig(name, version, author, credits string) *Config {
	return &Config{
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
