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

	NFFunction.Function{
		Name:         "Quit",
		Type:         "Quit",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(Quit)

	NFFunction.Function{
		Name:         "Error",
		Type:         "Error",
		RequiredArgs: NFData.NewNFInterface(NFData.NewKeyVal("Error", "This should be an error message in a string format")),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(CustomError)

	NFFunction.Function{
		Name:         "New Game",
		Type:         "NewGame",
		RequiredArgs: NFData.NewNFInterface(NFData.NewKeyVal("NewGameScene", "This should be the name of the scene to start the game with. THIS IS CASE SENSITIVE")),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(NewGame)

	NFFunction.Function{
		Name:         "Save As",
		Type:         "SaveAs",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(SaveAs)

	NFFunction.Function{
		Name:         "Load Game",
		Type:         "LoadGame",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(LoadGame)

	NFFunction.Function{
		Name:         "Continue Game",
		Type:         "ContinueGame",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(ContinueGame)
}
