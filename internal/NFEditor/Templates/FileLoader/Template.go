package local

import (
	"embed"
	"go.novellaforge.dev/novellaforge/pkg/NFFS"
)

//---***DO NOT TOUCH THIS FILE UNLESS YOU KNOW FOR SURE WHAT YOU ARE DOING***---//
//---***THIS FILE IS MANAGED BY THE EDITOR AND WILL LIKELY BE OVERWRITTEN IF YOU ARE USING THE EDITOR***---//
//---***IT IS NOT RECOMMENDED TO MANUALLY EDIT THIS FILE***---//

//TODO switch these embedded hard-codings to a variable like if embedData dataToEmbed
// which contains files or dirs IN THE LOCAL DIRECTORY ONLY to embed seperated by spaces

//go:embed Local.NFConfig {{if .EmbedData}}data{{end}} {{if .EmbedAssets}}assets{{end}}
var localFS embed.FS

func init() {
	// Set the embedded filesystem to use for loading files
	NFFS.EmbedFS(localFS)
}
