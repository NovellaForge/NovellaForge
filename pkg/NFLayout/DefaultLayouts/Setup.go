package DefaultLayouts

import (
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout"
	"log"
)

// Import is a function used to allow importing of the default layouts package without errors or warnings
func Import() {}

// init() registers the default layouts to be used within the game
// The init function is called when the package is imported, but in order
// to avoid unused import warnings, you can call the Import() function, which does nothing
func init() {
	Import()
	log.Println("Registering Default Layouts")

	// VBox Layout
	NFLayout.Layout{
		Type:         "VBox",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(VBoxLayoutHandler)

	// HBox Layout
	NFLayout.Layout{
		Type:         "HBox",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(HBoxLayoutHandler)

	// Grid Layout
	NFLayout.Layout{
		Type:         "Grid",
		RequiredArgs: NFData.NewNFInterface(NFData.NewKeyVal("Columns", 0)),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(GridLayoutHandler)

	// Tab Layout
	NFLayout.Layout{
		Type:         "Tab",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(TabLayoutHandler)

	// Border Layout
	NFLayout.Layout{
		Type:         "Border",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(BorderLayoutHandler)
}
