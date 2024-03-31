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
	NFWidget.Widget{
		Type:         "VBoxContainer",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Hidden", false),
		),
	}.Register(VBoxContainerHandler)

	// HBoxContainerHandler
	NFWidget.Widget{
		Type:         "HBoxContainer",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Hidden", false),
		),
	}.Register(HBoxContainerHandler)

	// FormHandler
	NFWidget.Widget{
		Type:         "Form",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Hidden", false),
			NFData.NewKeyVal("SubmitText", ""),
			NFData.NewKeyVal("CancelText", ""),
			NFData.NewKeyVal("OnSubmitted", ""),
			NFData.NewKeyVal("OnSubmittedArgs", NFData.NewNFInterface()),
			NFData.NewKeyVal("OnCancelled", ""),
			NFData.NewKeyVal("OnCancelledArgs", NFData.NewNFInterface()),
			NFData.NewKeyVal("Enabled", true),
		),
	}.Register(FormHandler)

	// LabelHandler
	NFWidget.Widget{
		Type:         "Label",
		RequiredArgs: NFData.NewNFInterface(NFData.NewKeyVal("Text", "")),
		OptionalArgs: NFData.NewNFInterface(),
	}.Register(LabelHandler)

	// ButtonHandler
	NFWidget.Widget{
		Type:         "Button",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(
			NFData.NewKeyVal("Text", ""),
			NFData.NewKeyVal("Icon", ""),
			NFData.NewKeyVal("OnTapped", ""),
			NFData.NewKeyVal("OnTappedArgs", NFData.NewNFInterface()),
			NFData.NewKeyVal("Hidden", false),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Enabled", true),
		),
	}.Register(ButtonHandler)

	// ToolBarHandler
	NFWidget.Widget{
		Type:         "ToolBar",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Hidden", false),
		),
	}.Register(ToolBarHandler)

	// EntryHandler
	NFWidget.Widget{
		Type:         "Entry",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(
			NFData.NewKeyVal("PlaceHolder", ""),
			NFData.NewKeyVal("Text", ""),
			NFData.NewKeyVal("OnChanged", ""),
			NFData.NewKeyVal("OnChangedArgs", NFData.NewNFInterface()),
			NFData.NewKeyVal("OnSubmitted", ""),
			NFData.NewKeyVal("OnSubmittedArgs", NFData.NewNFInterface()),
			NFData.NewKeyVal("OnCursorChanged", ""),
			NFData.NewKeyVal("OnCursorChangedArgs", NFData.NewNFInterface()),
			NFData.NewKeyVal("Hidden", false),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Enabled", true),
		),
	}.Register(EntryHandler)

	// PasswordEntryHandler
	NFWidget.Widget{
		Type:         "PasswordEntry",
		RequiredArgs: NFData.NewNFInterface(),
		OptionalArgs: NFData.NewNFInterface(
			NFData.NewKeyVal("PlaceHolder", ""),
			NFData.NewKeyVal("Text", ""),
			NFData.NewKeyVal("OnChanged", ""),
			NFData.NewKeyVal("OnChangedArgs", NFData.NewNFInterface()),
			NFData.NewKeyVal("OnSubmitted", ""),
			NFData.NewKeyVal("OnSubmittedArgs", NFData.NewNFInterface()),
			NFData.NewKeyVal("OnCursorChanged", ""),
			NFData.NewKeyVal("OnCursorChangedArgs", NFData.NewNFInterface()),
			NFData.NewKeyVal("Hidden", false),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Enabled", true),
		),
	}.Register(PasswordEntryHandler)

	// SliderHandler
	NFWidget.Widget{
		Type: "Slider",
		RequiredArgs: NFData.NewNFInterface(
			NFData.NewKeyVal("Min", 0.0),
			NFData.NewKeyVal("Max", 1.0),
		),
		OptionalArgs: NFData.NewNFInterface(
			NFData.NewKeyVal("Value", 0.0),
			NFData.NewKeyVal("Step", 0.1),
			NFData.NewKeyVal("OnChanged", ""),
			NFData.NewKeyVal("OnChangedArgs", NFData.NewNFInterface()),
			NFData.NewKeyVal("Hidden", false),
			NFData.NewKeyVal("Position", fyne.NewPos(0, 0)),
			NFData.NewKeyVal("Size", fyne.NewSize(0, 0)),
			NFData.NewKeyVal("Enabled", true),
		),
	}.Register(SliderHandler)
}
