package DefaultWidgets

import (
	"fyne.io/fyne/v2"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
	"log"
)

func Import() {}

func init() {
	Import()
	log.Println("Registering Default Widgets")

	// VBoxContainerHandler
	vbox := NFWidget.Widget{
		Type:         "VBoxContainer",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Hidden", false),
		),
	}
	vbox.Register(VBoxContainerHandler)

	// HBoxContainerHandler
	hbox := NFWidget.Widget{
		Type:         "HBoxContainer",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Hidden", false),
		),
	}
	hbox.Register(HBoxContainerHandler)

	// FormHandler
	form := NFWidget.Widget{
		Type:         "Form",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Hidden", false),
			NFData.NewKeyVal("SubmitText", ""),
			NFData.NewKeyVal("CancelText", ""),
			NFData.NewKeyVal("OnSubmitted", ""),
			NFData.NewKeyVal("OnSubmittedArgs", NFData.NewNFInterfaceMap()),
			NFData.NewKeyVal("OnCancelled", ""),
			NFData.NewKeyVal("OnCancelledArgs", NFData.NewNFInterfaceMap()),
			NFData.NewKeyVal("Enabled", true),
		),
	}
	form.Register(FormHandler)

	// LabelHandler
	label := NFWidget.Widget{
		Type:         "Label",
		RequiredArgs: NFData.NewNFInterfaceMap(NFData.NewKeyVal("Text", "")),
		OptionalArgs: NFData.NewNFInterfaceMap(),
	}
	label.Register(LabelHandler)

	// ButtonHandler
	button := NFWidget.Widget{
		Type:         "Button",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(
			NFData.NewKeyVal("Text", ""),
			NFData.NewKeyVal("Icon", ""),
			NFData.NewKeyVal("OnTapped", ""),
			NFData.NewKeyVal("OnTappedArgs", NFData.NewNFInterfaceMap()),
			NFData.NewKeyVal("Hidden", false),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Enabled", true),
		),
	}
	button.Register(ButtonHandler)

	// ToolBarHandler
	toolbar := NFWidget.Widget{
		Type:         "ToolBar",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Hidden", false),
		),
	}
	toolbar.Register(ToolBarHandler)

	// EntryHandler
	entry := NFWidget.Widget{
		Type:         "Entry",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(
			NFData.NewKeyVal("PlaceHolder", ""),
			NFData.NewKeyVal("Text", ""),
			NFData.NewKeyVal("OnChanged", ""),
			NFData.NewKeyVal("OnChangedArgs", NFData.NewNFInterfaceMap()),
			NFData.NewKeyVal("OnSubmitted", ""),
			NFData.NewKeyVal("OnSubmittedArgs", NFData.NewNFInterfaceMap()),
			NFData.NewKeyVal("OnCursorChanged", ""),
			NFData.NewKeyVal("OnCursorChangedArgs", NFData.NewNFInterfaceMap()),
			NFData.NewKeyVal("Hidden", false),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Enabled", true),
		),
	}
	entry.Register(EntryHandler)

	// PasswordEntryHandler
	passwordEntry := NFWidget.Widget{
		Type:         "PasswordEntry",
		RequiredArgs: NFData.NewNFInterfaceMap(),
		OptionalArgs: NFData.NewNFInterfaceMap(
			NFData.NewKeyVal("PlaceHolder", ""),
			NFData.NewKeyVal("Text", ""),
			NFData.NewKeyVal("OnChanged", ""),
			NFData.NewKeyVal("OnChangedArgs", NFData.NewNFInterfaceMap()),
			NFData.NewKeyVal("OnSubmitted", ""),
			NFData.NewKeyVal("OnSubmittedArgs", NFData.NewNFInterfaceMap()),
			NFData.NewKeyVal("OnCursorChanged", ""),
			NFData.NewKeyVal("OnCursorChangedArgs", NFData.NewNFInterfaceMap()),
			NFData.NewKeyVal("Hidden", false),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Enabled", true),
		),
	}
	passwordEntry.Register(PasswordEntryHandler)

	// SliderHandler
	slider := NFWidget.Widget{
		Type: "Slider",
		RequiredArgs: NFData.NewNFInterfaceMap(
			NFData.NewKeyVal("Min", 0.0),
			NFData.NewKeyVal("Max", 1.0),
		),
		OptionalArgs: NFData.NewNFInterfaceMap(
			NFData.NewKeyVal("Value", 0.0),
			NFData.NewKeyVal("Step", 0.1),
			NFData.NewKeyVal("OnChanged", ""),
			NFData.NewKeyVal("OnChangedArgs", NFData.NewNFInterfaceMap()),
			NFData.NewKeyVal("Hidden", false),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Enabled", true),
		),
	}
	slider.Register(SliderHandler)
}
