package local

import (
	"embed"
)

//---***DO NOT TOUCH THIS FILE UNLESS YOU KNOW FOR SURE WHAT YOU ARE DOING***---//
//---***THIS FILE IS MANAGED BY THE EDITOR AND WILL LIKELY BE OVERWRITTEN IF YOU ARE USING THE EDITOR***---//
//---***IT IS NOT RECOMMENDED TO MANUALLY EDIT THIS FILE***---//

//go:embed Local.NFConfig {{if .EmbedData}}data{{end}} {{if .EmbedAssets}}assets{{end}}
var CombinedFS embed.FS
