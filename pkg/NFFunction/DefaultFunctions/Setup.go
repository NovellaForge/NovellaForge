package DefaultFunctions

import (
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFFunction"
	"log"
)

// Import is a function that exists to allow importing of this package even if you don't directly use any of its functions,
// while still running the init function without disabling the unused import warning,
// You can also import a functions package for side effects by changing its alias to _ but using the import function allows you to still use the package directly
func Import() {}

func init() {
	Import()
	log.Println("Registering Default Functions")

	quit := NFFunction.Function{
		Name:         "Quit",
		Type:         "Quit",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	quit.Register(Quit)

	customError := NFFunction.Function{
		Name:         "Error",
		Type:         "Error",
		RequiredArgs: NFData.NewNFInterfaceMap(NFData.NewKeyVal("Error", "This should be an error message in a string format")),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	customError.Register(CustomError)

	newGame := NFFunction.Function{
		Name:         "New Game",
		Type:         "NewGame",
		RequiredArgs: NFData.NewNFInterfaceMap(NFData.NewKeyVal("NewGameScene", "This should be the name of the scene to start the game with. THIS IS CASE SENSITIVE")),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	newGame.Register(NewGame)

	saveAs := NFFunction.Function{
		Name:         "Save As",
		Type:         "SaveAs",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	saveAs.Register(SaveAs)

	loadGame := NFFunction.Function{
		Name:         "Load Game",
		Type:         "LoadGame",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	loadGame.Register(LoadGame)

	continueGame := NFFunction.Function{
		Name:         "Continue Game",
		Type:         "ContinueGame",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	continueGame.Register(ContinueGame)
}
