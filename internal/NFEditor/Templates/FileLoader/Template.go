package game

import (
	"embed"
	"go.novellaforge.dev/novellaforge/pkg/NFFS"
)

//TODO Allow full editing of this file by creating a new file with an init on build that will allow embedding more files

//---***You can add any embeded files located within the game directory simply by adding them after the Game.NFConfig file***---//
//---***You need to separate each file with a space, not a comma or newline***---//
//---***Example: Game.NFConfig Game.NFScene Game.NFLayout or even directories like Game.NFConfig data assets***---//
//---***The file will be loaded into the embedded filesystem and can be accessed using the NFFS package***---//

//go:embed Game.NFConfig
var gameFS embed.FS

// Import This needs to be run for the embedded filesystem to work
func Import() {}

func init() {
	// Set the embedded filesystem to use for loading files
	NFFS.EmbedFS(gameFS)
}
