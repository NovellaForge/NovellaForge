package DefaultWidgets

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
	"go.novellaforge.dev/novellaforge/pkg/NFFunction"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
)

// VBoxContainerHandler creates a vertical box container
func VBoxContainerHandler(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error) {
	vbox := container.NewVBox()
	var children []NFWidget.Widget
	for _, value := range args {
		switch value.(type) {
		case []NFWidget.Widget:
			children = value.([]NFWidget.Widget)
		}
	}
	if len(children) == 0 {
		vbox.Add(layout.NewSpacer())
	} else {
		for _, child := range children {
			parsedChild, err := child.Parse(window)
			if err != nil {
				_, _, _ = NFFunction.ParseAndRun(window, "Error", map[string]interface{}{"message": err.Error()})
			}
			vbox.Add(parsedChild)
		}
	}
	return vbox, nil
}

// HBoxContainerHandler creates a horizontal box container
func HBoxContainerHandler(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error) {
	hbox := container.NewHBox()
	var children []NFWidget.Widget
	for _, value := range args {
		switch value.(type) {
		case []NFWidget.Widget:
			children = value.([]NFWidget.Widget)
		}
	}
	if len(children) == 0 {
		hbox.Add(layout.NewSpacer())
	} else {
		for _, child := range children {
			parsedChild, err := child.Parse(window)
			if err != nil {
				_, _, _ = NFFunction.ParseAndRun(window, "Error", map[string]interface{}{"message": err.Error()})
			}
			hbox.Add(parsedChild)
		}
	}
	return hbox, nil
}

// FormHandler creates a form container
func FormHandler(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error) {
	form := widget.NewForm()
	var children []NFWidget.Widget
	for _, value := range args {
		switch value.(type) {
		case []NFWidget.Widget:
			children = value.([]NFWidget.Widget)
		}
	}
	if len(children) == 0 {
		form.Append("Form Empty", widget.NewLabel(""))
	} else {
		for _, child := range children {
			parsedChild, err := child.Parse(window)
			if err != nil {
				_, _, _ = NFFunction.ParseAndRun(window, "Error", map[string]interface{}{"message": err.Error()})
			}
			if _, ok := child.Properties["Text"]; !ok {
				child.Properties["Text"] = ""
			}
			form.Append(child.Properties["Text"].(string), parsedChild)
		}
	}
	return form, nil
}

// LabelHandler creates a label
func LabelHandler(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error) {
	properties := args[0].(map[string]interface{})
	//Check if the text property exists, if it doesn't, return an error
	if _, ok := properties["Text"]; !ok {
		return nil, NFError.ErrMissingArgument
	}
	label := widget.NewLabel(properties["Text"].(string))
	label.Wrapping = fyne.TextWrapWord
	return label, nil
}

// ButtonHandler creates a button
func ButtonHandler(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error) {
	properties := args[0].(map[string]interface{})
	//If Text does not exist, set it to be blank
	if _, ok := properties["Text"]; !ok {
		properties["Text"] = ""
	}
	text := properties["Text"].(string)
	//Do the same for Action
	if _, ok := properties["Action"]; !ok {
		properties["Action"] = ""
	}
	action := properties["Action"].(string)
	var button *widget.Button
	//if the args contain "Icon" then create a button with an icon
	if iconPath, ok := properties["Icon"]; ok && iconPath != "" {
		icon := canvas.NewImageFromFile(iconPath.(string))
		button = widget.NewButtonWithIcon(text, icon.Resource, func() {
			_, _, err := NFFunction.ParseAndRun(window, action, properties)
			if err != nil {
				_, _, _ = NFFunction.ParseAndRun(window, "Error", map[string]interface{}{"message": err.Error()})
			}
		})
	} else {
		button = widget.NewButton(text, func() {
			_, _, err := NFFunction.ParseAndRun(window, action, properties)
			if err != nil {
				_, _, _ = NFFunction.ParseAndRun(window, "Error", map[string]interface{}{"message": err.Error()})
			}
		})
	}
	return button, nil
}

// ToolBarHandler creates a toolbar with the given children
func ToolBarHandler(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error) {
	//Loop through the children and switch on their type
	toolbar := widget.NewToolbar()
	var children []NFWidget.Widget
	for _, value := range args {
		switch value.(type) {
		case []NFWidget.Widget:
			children = value.([]NFWidget.Widget)
		}
	}
	if len(children) == 0 {
		return nil, errors.New("toolbar must have children")
	}
	for _, child := range children {
		switch child.Type {
		case "ToolbarAction":
			action := child.Properties["Action"].(string)
			childArgs := child.Properties["Args"].(map[string]interface{})
			iconStr := child.Properties["Icon"].(string)
			//Create an icon from the icon string path
			iconObject := canvas.NewImageFromFile(iconStr)
			//Create a new toolbar action with the icon
			icon := widget.NewIcon(iconObject.Resource)
			toolbarAction := widget.NewToolbarAction(icon.Resource, func() {
				_, _, err := NFFunction.ParseAndRun(window, action, childArgs)
				if err != nil {
					_, _, _ = NFFunction.ParseAndRun(window, "Error", map[string]interface{}{"message": err.Error()})
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
	return toolbar, nil
}

//TODO make the onchanged optional and add a validator

// EntryHandler creates an entry with an optional onchange function
func EntryHandler(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error) {
	properties := args[0].(map[string]interface{})
	entry := widget.NewEntry()
	entry.SetPlaceHolder(properties["PlaceHolder"].(string))
	entry.SetText(properties["Text"].(string))
	entry.OnChanged = func(s string) {
		_, _, err := NFFunction.ParseAndRun(window, properties["OnChanged"].(string), map[string]interface{}{"Value": s})
		if err != nil {
			_, _, _ = NFFunction.ParseAndRun(window, "Error", map[string]interface{}{"message": err.Error()})
		}
	}
	return entry, nil
}

// PasswordEntryHandler creates a password entry with an optional onchange function
func PasswordEntryHandler(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error) {
	properties := args[0].(map[string]interface{})
	entry := widget.NewPasswordEntry()
	entry.SetPlaceHolder(properties["PlaceHolder"].(string))
	entry.SetText(properties["Text"].(string))
	if properties["Action"] != nil {
		entry.OnChanged = func(s string) {
			_, _, err := NFFunction.ParseAndRun(window, properties["Action"].(string), map[string]interface{}{"Value": s})
			if err != nil {
				_, _, _ = NFFunction.ParseAndRun(window, "Error", map[string]interface{}{"message": err.Error()})
			}
		}
	}
	return entry, nil
}

// SliderHandler creates a slider with an optional onchange function
func SliderHandler(window fyne.Window, args ...interface{}) (fyne.CanvasObject, error) {
	properties := args[0].(map[string]interface{})
	sMin := properties["Min"].(float64)
	sMax := properties["Max"].(float64)
	slider := widget.NewSlider(sMin, sMax)
	if properties["Action"] != nil {
		slider.OnChanged = func(f float64) {
			_, _, err := NFFunction.ParseAndRun(window, properties["Action"].(string), map[string]interface{}{"Value": f})
			if err != nil {
				_, _, _ = NFFunction.ParseAndRun(window, "Error", map[string]interface{}{"message": err.Error()})
			}
		}
	}
	return slider, nil
}

//TODO Select, Tree, List, Table, and more
