package CalsWidgets

import (
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
	"log"
)

func Import() { log.Println("Importing Cals Widgets") }

func init() {
	Import()
	NFWidget.Register("Dialog", DialogHandler)
}
