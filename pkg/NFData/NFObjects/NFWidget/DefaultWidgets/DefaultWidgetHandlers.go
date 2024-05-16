package DefaultWidgets

import (
	"errors"
	"fmt"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFFunction"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFWidget"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFStyling"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
)

//TODO Create an error box widget instead of a function that creates an error box with an option to unwrap the error
// Also need to set up the new Action based function system to possibly run actions on errors
// Make it so that we can call the new Error box from either the parser or directly with a window and error passed in

// VBoxContainerHandler creates a vertical box container
func VBoxContainerHandler(window fyne.Window, _ *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	var widgetError error
	vbox := container.NewVBox()
	if len(w.Children) == 0 {
		vbox.Add(layout.NewSpacer())
	} else {
		for _, child := range w.Children {
			parsedChild, err := child.Parse(window)
			if err != nil {
				widgetError = errors.Join(widgetError, err)
				continue
			}
			vbox.Add(parsedChild)
		}
	}

	var hidden = false
	err := w.Args.Get("Hidden", &hidden)
	if err == nil {
		if hidden {
			vbox.Hide()
		} else {
			vbox.Show() //The default state is actually shown, this is just for clarity
		}
	}
	var position = vbox.Position()
	err = w.Args.Get("Position", &position)
	if err == nil {
		vbox.Move(position)
	}

	var size = vbox.Size()
	err = w.Args.Get("Size", &size)
	if err == nil {
		vbox.Resize(size)
	}
	if widgetError != nil {
		widgetError = errors.Join(NFError.NewErrWidgetParse(w.GetName(), w.GetType(), w.GetID(), "Wrapped Errors"), widgetError)
	}
	return vbox, widgetError
}

// HBoxContainerHandler creates a horizontal box container
func HBoxContainerHandler(window fyne.Window, _ *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	var widgetError error
	hbox := container.NewHBox()
	if len(w.Children) == 0 {
		hbox.Add(layout.NewSpacer())
	} else {
		for _, child := range w.Children {
			parsedChild, err := child.Parse(window)
			if err != nil {
				widgetError = errors.Join(widgetError, err)
				continue
			}
			hbox.Add(parsedChild)
		}
	}

	var hidden = false
	err := w.Args.Get("Hidden", &hidden)
	if err == nil {
		if hidden {
			hbox.Hide()
		} else {
			hbox.Show() //The default state is actually shown, this is just for clarity
		}
	}

	var position = hbox.Position()
	err = w.Args.Get("Position", &position)
	if err == nil {
		hbox.Move(position)
	}

	var size = hbox.Size()
	err = w.Args.Get("Size", &size)
	if err == nil {
		hbox.Resize(size)
	}
	if widgetError != nil {
		widgetError = errors.Join(NFError.NewErrWidgetParse(w.GetName(), w.GetType(), w.GetID(), "Wrapped Errors"), widgetError)
	}
	return hbox, widgetError
}

// FormHandler creates a form container
func FormHandler(window fyne.Window, _ *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	var widgetError error
	form := widget.NewForm()
	if len(w.Children) == 0 {
		log.Println("Form Empty")
		form.Append("Form Empty", widget.NewLabel(""))
	} else {
		for _, child := range w.Children {
			parsedChild, err := child.Parse(window)
			if err != nil {
				widgetError = errors.Join(widgetError, err)
				continue
			}
			var label string
			err = child.Args.Get("FormLabel", &label)
			if err != nil || label == "" {
				childErr := NFError.NewErrWidgetParse(child.GetName(), child.GetType(), child.GetID(), "Error Getting Form Label")
				widgetError = errors.Join(widgetError, childErr)
				form.Append("Unknown", parsedChild)
			} else {
				form.Append(label, parsedChild)
			}
		}
	}

	var hidden = false
	err := w.Args.Get("Hidden", &hidden)
	if err == nil {
		if hidden {
			form.Hide()
		} else {
			form.Show() //The default state is actually shown, this is just for clarity
		}
	}

	var position = form.Position()
	err = w.Args.Get("Position", &position)
	if err == nil {
		form.Move(position)
	}

	var size = form.Size()
	err = w.Args.Get("Size", &size)
	if err == nil {
		form.Resize(size)
	}

	var submitText string
	err = w.Args.Get("SubmitText", &submitText)
	if err == nil {
		form.SubmitText = submitText
	}

	var cancelText string
	err = w.Args.Get("CancelText", &cancelText)
	if err == nil {
		form.CancelText = cancelText
	}

	form.OnSubmit = func() {
		results, err := w.RunAction("OnSubmit", window, nil)
		if err != nil {
			errText := fmt.Sprintf("Error running OnSubmit for form %s: ", w.GetName())
			results.Set("Error", errText+err.Error())
			_, _ = NFFunction.ParseAndRun(window, "Error", results)
		}
	}

	form.OnCancel = func() {
		results, err := w.RunAction("OnCancel", window, nil)
		if err != nil {
			errText := fmt.Sprintf("Error running OnCancel for form %s: ", w.GetName())
			results.Set("Error", errText+err.Error())
			_, _ = NFFunction.ParseAndRun(window, "Error", results)
		}
	}

	var enabled = true
	err = w.Args.Get("Enabled", &enabled)
	if err == nil {
		if enabled {
			form.Enable()
		} else {
			form.Disable()
		}
	}

	return form, nil
}

// LabelHandler creates a label
func LabelHandler(_ fyne.Window, args *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	var styling NFStyling.NFStyling
	err := args.Get("Styling", &styling)
	if err != nil {
		styling = NFStyling.NewTextStyling()
	}
	var text string
	err = args.Get("Text", &text)
	if err != nil {
		var reference NFData.NFReference
		err = args.Get("Text", &reference)
		if err != nil {
			return nil, NFError.NewErrWidgetParse(w.GetName(), w.GetType(), w.GetID(), "Error Getting Text")
		}
		//Check if the reference is a binding
		if reference.IsBinding {
			textBinding, err := reference.GetBinding()
			if err != nil {
				return nil, NFError.NewErrWidgetParse(w.GetName(), w.GetType(), w.GetID(), "Error Getting Text Binding")
			}
			label := widget.NewLabelWithData(textBinding.(binding.String))
			label.TextStyle = styling.TextStyle
			label.Wrapping = styling.Wrapping
			label.Refresh()
			return label, nil
		} else {
			return nil, NFError.NewErrWidgetParse(w.GetName(), w.GetType(), w.GetID(), "Error Getting Text")
		}
	}
	label := widget.NewLabel(text)
	label.TextStyle = styling.TextStyle
	label.Wrapping = styling.Wrapping
	label.Refresh()
	return label, nil
}

// ButtonHandler creates a button
func ButtonHandler(window fyne.Window, args *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	var button *widget.Button
	var text string
	_ = args.Get("Text", &text)
	var iconPath string
	_ = args.Get("Icon", &iconPath)
	if iconPath != "" {
		icon := canvas.NewImageFromFile(iconPath)
		button = widget.NewButtonWithIcon(text, icon.Resource, func() {})
		var iconPlacement = button.IconPlacement
		err := w.Args.Get("IconPlacement", &iconPlacement)
		if err == nil {
			button.IconPlacement = iconPlacement
		}

	} else {
		button = widget.NewButton(text, func() {})
	}

	button.OnTapped = func() {
		results, err := w.RunAction("OnTapped", window, nil)
		if err != nil {
			errText := fmt.Sprintf("Error running OnTapped for button %s: ", w.GetName())
			results.Set("Error", errText+err.Error())
			_, _ = NFFunction.ParseAndRun(window, "Error", results)
		}
	}

	var hidden = button.Hidden
	err := w.Args.Get("Hidden", &hidden)
	if err == nil {
		if hidden {
			button.Hide()
		} else {
			button.Show() //The default state is actually shown, this is just for clarity
		}
	}

	var position = button.Position()
	err = w.Args.Get("Position", &position)
	if err == nil {
		button.Move(position)
	}

	var size = button.Size()
	err = w.Args.Get("Size", &size)
	if err == nil {
		button.Resize(size)
	}

	var enabled = true
	err = w.Args.Get("Enabled", &enabled)
	if err == nil {
		if enabled {
			button.Enable()
		} else {
			button.Disable()
		}
	}

	return button, nil
}

// ImageHandler creates an image
func ImageHandler(window fyne.Window, args *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	var path string
	err := args.Get("Path", &path)
	if err != nil {
		return nil, NFError.NewErrWidgetParse(w.GetName(), w.GetType(), w.GetID(), "Error Getting Image Path")
	}

	// Check if there is an image at the path, or at assets/image/path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = "assets/image/" + path
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, NFError.NewErrWidgetParse(w.GetName(), w.GetType(), w.GetID(), "Error Getting Image From Path")
		}
	}

	image := canvas.NewImageFromFile(path)
	var hidden = false
	err = w.Args.Get("Hidden", &hidden)
	if err == nil {
		if hidden {
			image.Hide()
		} else {
			image.Show() //The default state is actually shown, this is just for clarity
		}
	}

	var position = image.Position()
	err = w.Args.Get("Position", &position)
	if err == nil {
		image.Move(position)
	}

	// Set the size of the image if it is provided
	var size = image.Size()
	err = w.Args.Get("MinSize", &size)
	if err == nil {
		image.Resize(size)
	}

	scroll := container.NewVScroll(image)
	scroll.SetMinSize(image.MinSize())
	image.FillMode = canvas.ImageFillContain
	image.Refresh()

	return scroll, nil

}

// ToolBarHandler creates a toolbar with the given widget
func ToolBarHandler(window fyne.Window, _ *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	toolbar := widget.NewToolbar()
	var widgetError error
	if len(w.Children) == 0 {
		widgetError = errors.Join(widgetError, NFError.NewErrWidgetParse(w.GetName(), w.GetType(), w.GetID(), "No Children Found"))
	}
	for _, child := range w.Children {
		switch child.Type {
		case "ToolbarAction":
			var iconPath string
			err := child.Args.Get("Icon", &iconPath)
			var icon *widget.Icon
			if err != nil {
				log.Println(NFError.NewErrWidgetParse(child.GetName(), child.GetType(), child.GetID(), "Error Getting Icon Path"))
				icon = widget.NewIcon(theme.BrokenImageIcon())
			} else {
				iconObject := canvas.NewImageFromFile(iconPath)
				icon = widget.NewIcon(iconObject.Resource)
			}
			toolbarAction := widget.NewToolbarAction(icon.Resource, func() {
				results, err := child.RunAction("OnActivated", window, nil)
				if err != nil {
					errText := fmt.Sprintf("Error running OnActivated for toolbar action %s: ", child.GetName())
					results.Set("Error", errText+err.Error())
					_, _ = NFFunction.ParseAndRun(window, "Error", results)
				}
			})
			toolbar.Append(toolbarAction)
		case "ToolbarSeparator":
			toolbarSeparator := widget.NewToolbarSeparator()
			toolbar.Append(toolbarSeparator)
		case "ToolbarSpacer":
			toolbarSpacer := widget.NewToolbarSpacer()
			toolbar.Append(toolbarSpacer)
		}
	}
	return toolbar, widgetError
}

// EntryHandler creates an entry with an optional onchange function
func EntryHandler(window fyne.Window, args *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	entry := widget.NewEntry()
	var placeHolder string
	err := args.Get("PlaceHolder", &placeHolder)
	if err == nil {
		entry.SetPlaceHolder(placeHolder)
	}
	var text string
	err = args.Get("Text", &text)
	if err == nil {
		entry.SetText(text)
	}

	var hidden = entry.Hidden
	err = w.Args.Get("Hidden", &hidden)
	if err == nil {
		if hidden {
			entry.Hide()
		} else {
			entry.Show() //The default state is actually shown, this is just for clarity
		}
	}

	var position = entry.Position()
	err = w.Args.Get("Position", &position)
	if err == nil {
		entry.Move(position)
	}

	var size = entry.Size()
	err = w.Args.Get("Size", &size)
	if err == nil {
		entry.Resize(size)
	}

	var enabled = true
	err = w.Args.Get("Enabled", &enabled)
	if err == nil {
		if enabled {
			entry.Enable()
		} else {
			entry.Disable()
		}
	}

	entry.OnChanged = func(s string) {
		newArgs := NFData.NewNFInterfaceMap(NFData.NewKeyVal("EntryValue", s))
		results, err := w.RunAction("OnChanged", window, newArgs)
		if err != nil {
			errText := fmt.Sprintf("Error running OnChanged for entry %s: ", w.GetName())
			results.Set("Error", errText+err.Error())
			_, _ = NFFunction.ParseAndRun(window, "Error", results)
		}
	}

	//TODO other entry actions

	return entry, nil
}

// PasswordEntryHandler creates a password entry with an optional onchange function
func PasswordEntryHandler(window fyne.Window, args *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	entry := widget.NewPasswordEntry()
	var placeHolder string
	err := args.Get("PlaceHolder", &placeHolder)
	if err == nil {
		entry.SetPlaceHolder(placeHolder)
	}
	var text string
	err = args.Get("Text", &text)
	if err == nil {
		entry.SetText(text)
	}

	var hidden = entry.Hidden
	err = w.Args.Get("Hidden", &hidden)
	if err == nil {
		if hidden {
			entry.Hide()
		} else {
			entry.Show() //The default state is actually shown, this is just for clarity
		}
	}

	var position = entry.Position()
	err = w.Args.Get("Position", &position)
	if err == nil {
		entry.Move(position)
	}

	var size = entry.Size()
	err = w.Args.Get("Size", &size)
	if err == nil {
		entry.Resize(size)
	}

	var enabled = true
	err = w.Args.Get("Enabled", &enabled)
	if err == nil {
		if enabled {
			entry.Enable()
		} else {
			entry.Disable()
		}
	}

	entry.OnChanged = func(s string) {
		newArgs := NFData.NewNFInterfaceMap(NFData.NewKeyVal("EntryValue", s))
		results, err := w.RunAction("OnChanged", window, newArgs)
		if err != nil {
			errText := fmt.Sprintf("Error running OnChanged for entry %s: ", w.GetName())
			results.Set("Error", errText+err.Error())
			_, _ = NFFunction.ParseAndRun(window, "Error", results)
		}
	}

	//TODO Other entry actions

	return entry, nil
}

// SliderHandler creates a slider with an optional onchange function
func SliderHandler(window fyne.Window, args *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	var sMin float64
	err := args.Get("Min", &sMin)
	if err != nil {
		return nil, NFError.NewErrWidgetParse(w.GetName(), w.GetType(), w.GetID(), "Error Getting Min")
	}
	var sMax float64
	err = args.Get("Max", &sMax)
	if err != nil {
		return nil, NFError.NewErrWidgetParse(w.GetName(), w.GetType(), w.GetID(), "Error Getting Max")
	}
	slider := widget.NewSlider(sMin, sMax)

	var sValue float64
	err = args.Get("Value", &sValue)
	if err == nil {
		slider.Value = sValue
	}

	var sStep float64
	err = args.Get("Step", &sStep)
	if err == nil {
		slider.Step = sStep
	}

	slider.OnChanged = func(f float64) {
		newArgs := NFData.NewNFInterfaceMap(NFData.NewKeyVal("SliderValue", f))
		results, err := w.RunAction("OnChanged", window, newArgs)
		if err != nil {
			errText := fmt.Sprintf("Error running OnChanged for slider %s: ", w.GetName())
			results.Set("Error", errText+err.Error())
			_, _ = NFFunction.ParseAndRun(window, "Error", results)
		}
	}
	return slider, nil
}

//TODO We need to add in more optional arguments for those that can support them and add in more default widgets
