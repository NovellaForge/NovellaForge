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
	vbox := NFLayout.Layout{
		Type:         "VBox",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	vbox.Register(VBoxLayoutHandler)

	// HBox Layout
	hbox := NFLayout.Layout{
		Type:         "HBox",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	hbox.Register(HBoxLayoutHandler)

	// Grid Layout
	grid := NFLayout.Layout{
		Type:         "Grid",
		RequiredArgs: NFData.NewNFInterfaceMap(NFData.NewKeyVal("Columns", 0)),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	grid.Register(GridLayoutHandler)

	// Tab Layout
	tab := NFLayout.Layout{
		Type:         "Tab",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	tab.Register(TabLayoutHandler)

	// Border Layout
	border := NFLayout.Layout{
		Type:         "Border",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	border.Register(BorderLayoutHandler)
}
