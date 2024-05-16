package DefaultFunctions

import (
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFFunction"
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
	log.Println("Registering Default Functions")

	quit := NFFunction.Function{
		Type:         "Quit",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	quit.Register(Quit)

	customError := NFFunction.Function{
		Type:         "Error",
		RequiredArgs: NFData.NewNFInterfaceMap(NFData.NewKeyVal("Error", "This should be an error message in a string format")),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	customError.Register(CustomError)

	newGame := NFFunction.Function{
		Type:         "NewGame",
		RequiredArgs: NFData.NewNFInterfaceMap(NFData.NewKeyVal("NewGameScene", "This should be the name of the scene to start the game with. THIS IS CASE SENSITIVE")),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	newGame.Register(NewGame)

	saveAs := NFFunction.Function{
		Type:         "SaveAs",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	saveAs.Register(SaveAs)

	loadGame := NFFunction.Function{
		Type:         "LoadGame",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	loadGame.Register(LoadGame)

	continueGame := NFFunction.Function{
		Type:         "ContinueGame",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	continueGame.Register(ContinueGame)
}
