package CalsWidgets

import (
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
	"log"
)

func Import() { log.Println("Importing Cals Widgets") }

func init() {
	Import()
	NFWidget.Register("NarrativeBox", DialogHandler)
}
