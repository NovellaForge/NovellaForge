package DefaultWidgets

import (
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
	"log"
)

func Import() {}

func init() {
	Import()
	log.Println("Registering Default Widgets")
	NFWidget.Register("VBoxContainer", VBoxContainerHandler)
	NFWidget.Register("HBoxContainer", HBoxContainerHandler)
	NFWidget.Register("Form", FormHandler)
	NFWidget.Register("Label", LabelHandler)
	NFWidget.Register("Button", ButtonHandler)
	NFWidget.Register("ToolBar", ToolBarHandler)
	NFWidget.Register("Entry", EntryHandler)
	NFWidget.Register("PasswordEntry", PasswordEntryHandler)
	NFWidget.Register("Slider", SliderHandler)
}
