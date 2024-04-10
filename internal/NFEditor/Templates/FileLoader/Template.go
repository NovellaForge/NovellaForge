package game

import (
	"embed"
	"go.novellaforge.dev/novellaforge/pkg/NFFS"
)

//---***DO NOT TOUCH THIS FILE UNLESS YOU KNOW FOR SURE WHAT YOU ARE DOING***---//
//---***THIS FILE IS MANAGED BY THE EDITOR AND WILL LIKELY BE OVERWRITTEN IF YOU ARE USING THE EDITOR***---//
//---***IT IS NOT RECOMMENDED TO MANUALLY EDIT THIS FILE***---//

//go:embed Game.NFConfig {{if .Embed}}{{.EmbedFiles}}{{end}}
var gameFS embed.FS

func init() {
	// Set the embedded filesystem to use for loading files
	NFFS.EmbedFS(gameFS)
}
