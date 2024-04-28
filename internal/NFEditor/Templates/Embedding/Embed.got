package game

import (
	"embed"
	"go.novellaforge.dev/novellaforge/pkg/NFFS"
)

//go:embed {{.EmbedFiles}}
var embeddedFS embed.FS

func init() {
	NFFS.EmbedFS(embeddedFS)
}




