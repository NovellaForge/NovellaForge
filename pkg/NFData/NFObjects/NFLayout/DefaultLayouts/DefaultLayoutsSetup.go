package DefaultLayouts

import (
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFLayout"
	"log"
)

// Import is an empty function, created to allow the inclusion of this package in other parts of the code,
// even if none of its functions are directly used.
// This ensures that the init function is executed without triggering warnings about unused imports.
//
// While it's possible to import a package for its side effects by changing its alias to _,
// using the Import function provides the added benefit of retaining direct access to the package's contents.
func Import() {}

// This init() registers the default layouts to be used within the game
// In go the init function is called when the package is imported, but in order
// to avoid unused import warnings, you can call the empty Import() function, which does nothing
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
