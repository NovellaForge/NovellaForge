package DefaultWidgets

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFFunction"
	"go.novellaforge.dev/novellaforge/pkg/NFStyling"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
	"log"
)

// VBoxContainerHandler creates a vertical box container
func VBoxContainerHandler(window fyne.Window, _ *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	ID, Type := w.GetInfo()
	errText := fmt.Sprintf("Error parsing Widget %s of type %s: ", ID, Type)
	vbox := container.NewVBox()
	if len(w.Children) == 0 {
		vbox.Add(layout.NewSpacer())
	} else {
		for _, child := range w.Children {
			ChildID, ChildType := child.GetInfo()
			parsedChild, err := child.Parse(window)
			if err != nil {
				log.Println(errText + fmt.Sprintf("Error parsing child %s of type %s: %s", ChildID, ChildType, err.Error()))
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

	return vbox, nil
}

// HBoxContainerHandler creates a horizontal box container
func HBoxContainerHandler(window fyne.Window, _ *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	ID, Type := w.GetInfo()
	errText := fmt.Sprintf("Error parsing Widget %s of type %s: ", ID, Type)
	hbox := container.NewHBox()
	if len(w.Children) == 0 {
		hbox.Add(layout.NewSpacer())
	} else {
		for _, child := range w.Children {
			ChildID, ChildType := child.GetInfo()
			parsedChild, err := child.Parse(window)
			if err != nil {
				log.Println(errText + fmt.Sprintf("Error parsing child %s of type %s: %s", ChildID, ChildType, err.Error()))
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

	return hbox, nil
}

// FormHandler creates a form container
func FormHandler(window fyne.Window, _ *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	ID, Type := w.GetInfo()
	errText := fmt.Sprintf("Error parsing Widget %s of type %s: ", ID, Type)
	form := widget.NewForm()
	if len(w.Children) == 0 {
		log.Println(errText + "Form Empty")
		form.Append("Form Empty", widget.NewLabel(""))
	} else {
		for _, child := range w.Children {
			ChildID, ChildType := child.GetInfo()
			parsedChild, err := child.Parse(window)
			if err != nil {
				log.Println(errText + fmt.Sprintf("Error parsing child %s of type %s: %s", ChildID, ChildType, err.Error()))
				continue
			}
			var label string
			err = child.Args.Get("FormLabel", &label)
			if err != nil || label == "" {
				log.Println(errText + fmt.Sprintf("Error getting FormLabel for child %s of type %s: %s", ChildID, ChildType, err.Error()))
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

	var onSubmitted string
	err = w.Args.Get("OnSubmitted", &onSubmitted)
	if err != nil {
		log.Println(fmt.Sprintf("No OnSubmitted found for form %s", w.ID))
	} else {
		var functionArgs *NFData.NFInterfaceMap
		err = w.Args.Get("OnSubmittedArgs", &functionArgs)
		if err != nil {
			log.Println(fmt.Sprintf("No OnSubmittedArgs found for form %s", w.ID))
			functionArgs = NFData.NewNFInterfaceMap()
		}
		form.OnSubmit = func() {
			_, err := NFFunction.ParseAndRun(window, onSubmitted, functionArgs)
			if err != nil {
				log.Println(fmt.Sprintf("Error running OnSubmitted for form %s: %s", w.ID, err.Error()))
			}
		}
	}

	var onCancelled string
	err = w.Args.Get("OnCancelled", &onCancelled)
	if err != nil {
		log.Println(fmt.Sprintf("No OnCancelled found for form %s", w.ID))
	} else {
		var functionArgs *NFData.NFInterfaceMap
		err = w.Args.Get("OnCancelledArgs", &functionArgs)
		if err != nil {
			log.Println(fmt.Sprintf("No OnCancelledArgs found for form %s", w.ID))
			functionArgs = NFData.NewNFInterfaceMap()
		}
		form.OnCancel = func() {
			_, err := NFFunction.ParseAndRun(window, onCancelled, functionArgs)
			if err != nil {
				log.Println(fmt.Sprintf("Error running OnCancelled for form %s: %s", w.ID, err.Error()))
			}
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
			log.Println(fmt.Sprintf("Error getting Text for label %s: %s", w.ID, err.Error()))
			return nil, NFError.NewErrMissingArgument(w.Type, "Text")
		}
		//Check if the reference is a binding
		if reference.IsBinding {
			textBinding, err := reference.GetBinding()
			if err != nil {
				log.Println(fmt.Sprintf("Error getting Text for label %s: %s", w.ID, err.Error()))
				return nil, NFError.NewErrInvalidArgument(w.Type, "Text is an invalid binding reference")
			}
			label := widget.NewLabelWithData(textBinding.(binding.String))
			label.TextStyle = styling.TextStyle
			label.Wrapping = styling.Wrapping
			label.Refresh()
			return label, nil
		} else {
			log.Println(fmt.Sprintf("Error getting Text for label %s: %s", w.ID, err.Error()))
			return nil, NFError.NewErrInvalidArgument(w.Type, "Text is not a valid reference or binding")
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
	var onTapped string
	_ = args.Get("OnTapped", &onTapped)
	ID, _ := w.GetInfo()
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
	if onTapped != "" {
		var functionArgs *NFData.NFInterfaceMap
		err := args.Get("OnTappedArgs", &functionArgs)
		if err != nil {
			functionArgs = NFData.NewNFInterfaceMap()
		}
		functionArgs.Set("WidgetID", ID)
		functionArgs.Set("WidgetType", "Button")
		button.OnTapped = func() {
			_, err = NFFunction.ParseAndRun(window, onTapped, functionArgs)
			if err != nil {
				errText := fmt.Sprintf("Error running onTapped %s for button %s: ", onTapped, ID)
				args.Set("Error", errText+err.Error())
				_, _ = NFFunction.ParseAndRun(window, "Error", args)
			}
		}
	} else {
		log.Println(fmt.Sprintf("No OnTapped found for button %s", ID))
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

// ToolBarHandler creates a toolbar with the given widget
func ToolBarHandler(window fyne.Window, _ *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	toolbar := widget.NewToolbar()
	ID, Type := w.GetInfo()
	errText := fmt.Sprintf("Error parsing Widget %s of type %s: ", ID, Type)
	if len(w.Children) == 0 {
		log.Println(errText + "Toolbar Empty")
		return nil, errors.New("toolbar Empty")
	}
	for _, child := range w.Children {
		switch child.Type {
		case "ToolbarAction":
			var action string
			_ = child.Args.Get("Action", &action)
			var iconPath string
			err := child.Args.Get("Icon", &iconPath)
			var icon *widget.Icon
			if err != nil {
				log.Println(errText + fmt.Sprintf("Error getting Icon for toolbar %s: %s", ID, err.Error()))
				icon = widget.NewIcon(theme.CancelIcon())
			} else {
				iconObject := canvas.NewImageFromFile(iconPath)
				icon = widget.NewIcon(iconObject.Resource)
			}
			toolbarAction := widget.NewToolbarAction(icon.Resource, func() {})
			if action != "" {
				var functionArgs *NFData.NFInterfaceMap
				err = child.Args.Get("OnActivatedArgs", &functionArgs)
				if err != nil {
					log.Println(fmt.Sprintf("Error getting OnActivatedArgs for toolbar %s: %s", ID, err.Error()))
					functionArgs = NFData.NewNFInterfaceMap()
				}
				toolbarAction.OnActivated = func() {
					_, err = NFFunction.ParseAndRun(window, action, functionArgs)
					if err != nil {
						eText := fmt.Sprintf("%s Error running action %s for toolbar %s: ", errText, action, child.ID)
						child.Args.Set("Error", eText+err.Error())
						_, _ = NFFunction.ParseAndRun(window, "Error", child.Args)
					}
				}
			}
			toolbar.Append(toolbarAction)
		case "ToolbarSeparator":
			toolbarSeparator := widget.NewToolbarSeparator()
			toolbar.Append(toolbarSeparator)
		case "ToolbarSpacer":
			toolbarSpacer := widget.NewToolbarSpacer()
			toolbar.Append(toolbarSpacer)
		}
	}
	return toolbar, nil
}

// EntryHandler creates an entry with an optional onchange function
func EntryHandler(window fyne.Window, args *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	entry := widget.NewEntry()
	var placeHolder string
	err := args.Get("PlaceHolder", &placeHolder)
	if err != nil {
		log.Println(fmt.Sprintf("No placeholder found for entry %s", w.ID))
	} else {
		entry.SetPlaceHolder(placeHolder)
	}
	var text string
	err = args.Get("Text", &text)
	if err != nil {
		log.Println(fmt.Sprintf("No text found for entry %s", w.ID))
	} else {
		entry.SetText(text)
	}
	var onChanged string
	err = args.Get("OnChanged", &onChanged)
	if err != nil {
		log.Println(fmt.Sprintf("No OnChanged found for entry %s", w.ID))
	} else {
		var functionArgs *NFData.NFInterfaceMap
		err = args.Get("OnChangedArgs", &functionArgs)
		if err != nil {
			log.Println(fmt.Sprintf("No OnChangedArgs found for entry %s", w.ID))
			functionArgs = NFData.NewNFInterfaceMap()
		}
		entry.OnChanged = func(s string) {
			functionArgs.Set("Value", s)
			_, err := NFFunction.ParseAndRun(window, onChanged, functionArgs)
			if err != nil {
				log.Println(fmt.Sprintf("Error running OnChanged for entry %s: %s", w.ID, err.Error()))
			}
		}
	}
	var onSubmitted string
	err = args.Get("OnSubmitted", &onSubmitted)
	if err != nil {
		log.Println(fmt.Sprintf("No OnSubmitted found for entry %s", w.ID))
	} else {
		var functionArgs *NFData.NFInterfaceMap
		err = args.Get("OnSubmittedArgs", &functionArgs)
		if err != nil {
			log.Println(fmt.Sprintf("No OnSubmittedArgs found for entry %s", w.ID))
			functionArgs = NFData.NewNFInterfaceMap()
		}
		entry.OnSubmitted = func(s string) {
			functionArgs.Set("Value", s)
			_, err := NFFunction.ParseAndRun(window, onSubmitted, functionArgs)
			if err != nil {
				log.Println(fmt.Sprintf("Error running OnSubmitted for entry %s: %s", w.ID, err.Error()))
			}
		}
	}

	var onCursorChanged string
	err = args.Get("OnCursorChanged", &onCursorChanged)
	if err != nil {
		log.Println(fmt.Sprintf("No OnCursorChanged found for entry %s", w.ID))
	} else {
		var functionArgs *NFData.NFInterfaceMap
		err = args.Get("OnCursorChangedArgs", &functionArgs)
		if err != nil {
			log.Println(fmt.Sprintf("No OnCursorChangedArgs found for entry %s", w.ID))
			functionArgs = NFData.NewNFInterfaceMap()
		}
		entry.OnCursorChanged = func() {
			_, err := NFFunction.ParseAndRun(window, onCursorChanged, functionArgs)
			if err != nil {
				log.Println(fmt.Sprintf("Error running OnCursorChanged for entry %s: %s", w.ID, err.Error()))
			}
		}
	}
	return entry, nil
}

// PasswordEntryHandler creates a password entry with an optional onchange function
func PasswordEntryHandler(window fyne.Window, args *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	entry := widget.NewPasswordEntry()
	var placeHolder string
	err := args.Get("PlaceHolder", &placeHolder)
	if err != nil {
		log.Println(fmt.Sprintf("No placeholder found for entry %s", w.ID))
	} else {
		entry.SetPlaceHolder(placeHolder)
	}
	var text string
	err = args.Get("Text", &text)
	if err != nil {
		log.Println(fmt.Sprintf("No text found for entry %s", w.ID))
	} else {
		entry.SetText(text)
	}
	var onChanged string
	err = args.Get("OnChanged", &onChanged)
	if err != nil {
		log.Println(fmt.Sprintf("No OnChanged found for entry %s", w.ID))
	} else {
		var functionArgs *NFData.NFInterfaceMap
		err = args.Get("OnChangedArgs", &functionArgs)
		if err != nil {
			log.Println(fmt.Sprintf("No OnChangedArgs found for entry %s", w.ID))
			functionArgs = NFData.NewNFInterfaceMap()
		}
		functionArgs.Set("WidgetID", w.ID)
		functionArgs.Set("WidgetType", "PasswordEntry")
		entry.OnChanged = func(s string) {
			functionArgs.Set("Value", s)
			_, err := NFFunction.ParseAndRun(window, onChanged, functionArgs)
			if err != nil {
				log.Println(fmt.Sprintf("Error running OnChanged for entry %s: %s", w.ID, err.Error()))
			}
		}
	}
	var onSubmitted string
	err = args.Get("OnSubmitted", &onSubmitted)
	if err != nil {
		log.Println(fmt.Sprintf("No OnSubmitted found for entry %s", w.ID))
	} else {
		var functionArgs *NFData.NFInterfaceMap
		err = args.Get("OnSubmittedArgs", &functionArgs)
		if err != nil {
			log.Println(fmt.Sprintf("No OnSubmittedArgs found for entry %s", w.ID))
			functionArgs = NFData.NewNFInterfaceMap()
		}
		functionArgs.Set("WidgetID", w.ID)
		functionArgs.Set("WidgetType", "PasswordEntry")
		entry.OnSubmitted = func(s string) {
			functionArgs.Set("Value", s)
			_, err := NFFunction.ParseAndRun(window, onSubmitted, functionArgs)
			if err != nil {
				log.Println(fmt.Sprintf("Error running OnSubmitted for entry %s: %s", w.ID, err.Error()))
			}
		}
	}

	var onCursorChanged string
	err = args.Get("OnCursorChanged", &onCursorChanged)
	if err != nil {
		log.Println(fmt.Sprintf("No OnCursorChanged found for entry %s", w.ID))
	} else {
		var functionArgs *NFData.NFInterfaceMap
		err = args.Get("OnCursorChangedArgs", &functionArgs)
		if err != nil {
			log.Println(fmt.Sprintf("No OnCursorChangedArgs found for entry %s", w.ID))
			functionArgs = NFData.NewNFInterfaceMap()
		}
		functionArgs.Set("WidgetID", w.ID)
		functionArgs.Set("WidgetType", "PasswordEntry")
		entry.OnCursorChanged = func() {
			_, err := NFFunction.ParseAndRun(window, onCursorChanged, functionArgs)
			if err != nil {
				log.Println(fmt.Sprintf("Error running OnCursorChanged for entry %s: %s", w.ID, err.Error()))
			}
		}
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

	return entry, nil
}

// SliderHandler creates a slider with an optional onchange function
func SliderHandler(window fyne.Window, args *NFData.NFInterfaceMap, w *NFWidget.Widget) (fyne.CanvasObject, error) {
	var sMin float64
	err := args.Get("Min", &sMin)
	if err != nil {
		log.Println(fmt.Sprintf("No Min found for slider %s", w.ID))
		return nil, err
	}
	var sMax float64
	err = args.Get("Max", &sMax)
	if err != nil {
		log.Println(fmt.Sprintf("No Max found for slider %s", w.ID))
		return nil, err
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

	var onChanged string
	err = args.Get("OnChanged", &onChanged)
	if err != nil {
		log.Println(fmt.Sprintf("No OnChanged found for slider %s", w.ID))
	} else {
		var functionArgs *NFData.NFInterfaceMap
		err = args.Get("OnChangedArgs", &functionArgs)
		if err != nil {
			log.Println(fmt.Sprintf("No OnChangedArgs found for slider %s", w.ID))
			functionArgs = NFData.NewNFInterfaceMap()
		}
		functionArgs.Set("WidgetID", w.ID)
		functionArgs.Set("WidgetType", "Slider")
		slider.OnChanged = func(f float64) {
			functionArgs.Set("Value", f)
			_, err = NFFunction.ParseAndRun(window, onChanged, functionArgs)
			if err != nil {
				log.Println(fmt.Sprintf("Error running OnChanged for slider %s: %s", w.ID, err.Error()))
			}
		}
	}
	return slider, nil
}

//TODO We need to add in more optional arguments for those that can support them and add in more default widgets
