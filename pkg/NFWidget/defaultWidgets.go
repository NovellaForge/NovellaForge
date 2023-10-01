package NFWidget

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/pkg/NFFunction"
	"regexp"
	"strconv"
	"strings"
)

// VBoxContainerHandler creates a vertical box container
func VBoxContainerHandler(_ map[string]interface{}, children []*Widget, window fyne.Window) (fyne.CanvasObject, error) {
	vbox := container.NewVBox()
	for _, child := range children {
		parsedChild, err := child.Parse(window)
		if err != nil {
			_, _, _ = NFFunction.Parse(window, "Error", map[string]interface{}{"message": err.Error()})
		}
		vbox.Add(parsedChild)
	}
	return vbox, nil
}

// HBoxContainerHandler creates a horizontal box container
func HBoxContainerHandler(_ map[string]interface{}, children []*Widget, window fyne.Window) (fyne.CanvasObject, error) {
	hbox := container.NewHBox()
	for _, child := range children {
		parsedChild, err := child.Parse(window)
		if err != nil {
			_, _, _ = NFFunction.Parse(window, "Error", map[string]interface{}{"message": err.Error()})
		}
		hbox.Add(parsedChild)
	}
	return hbox, nil
}

// FormHandler creates a form
func FormHandler(_ map[string]interface{}, children []*Widget, window fyne.Window) (fyne.CanvasObject, error) {
	form := widget.NewForm()
	for _, child := range children {
		parsedChild, err := child.Parse(window)
		if err != nil {
			_, _, _ = NFFunction.Parse(window, "Error", map[string]interface{}{"message": err.Error()})
		}
		form.Append(child.Properties["Text"].(string), parsedChild)
	}
	return form, nil
}

// LabelHandler creates a label
func LabelHandler(args map[string]interface{}, _ []*Widget, _ fyne.Window) (fyne.CanvasObject, error) {
	label := widget.NewLabel(args["Text"].(string))
	label.Wrapping = fyne.TextWrapWord
	return label, nil
}

// ButtonHandler creates a button
func ButtonHandler(args map[string]interface{}, _ []*Widget, window fyne.Window) (fyne.CanvasObject, error) {
	text := args["Text"].(string)
	action := args["Action"].(string)
	var button *widget.Button
	//if the args contain "Icon" then create a button with an icon
	if iconPath, ok := args["Icon"]; ok && iconPath != "" {
		icon := canvas.NewImageFromFile(iconPath.(string))
		button = widget.NewButtonWithIcon(text, icon.Resource, func() {
			_, _, err := NFFunction.Parse(window, action, args)
			if err != nil {
				_, _, _ = NFFunction.Parse(window, "Error", map[string]interface{}{"message": err.Error()})
			}
		})
	} else {
		button = widget.NewButton(text, func() {
			_, _, err := NFFunction.Parse(window, action, args)
			if err != nil {
				_, _, _ = NFFunction.Parse(window, "Error", map[string]interface{}{"message": err.Error()})
			}
		})
	}
	return button, nil
}

// ToolBarHandler creates a toolbar with the given children
func ToolBarHandler(_ map[string]interface{}, children []*Widget, window fyne.Window) (fyne.CanvasObject, error) {
	//Loop through the children and switch on their type
	toolbar := widget.NewToolbar()
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
				_, _, err := NFFunction.Parse(window, action, childArgs)
				if err != nil {
					_, _, _ = NFFunction.Parse(window, "Error", map[string]interface{}{"message": err.Error()})
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

func ImageHandler(args map[string]interface{}, _ []*Widget, _ fyne.Window) (fyne.CanvasObject, error) {
	path := args["Path"]
	minSizeX := args["MinSizeX"]
	minSizeY := args["MinSizeY"]
	sizeX := args["SizeX"]
	sizeY := args["SizeY"]

	//Attempt to validate all the arguments we are using
	//Attempt to convert the path to a string with error checking
	path, ok := path.(string)
	if !ok {
		return nil, errors.New("invalid image path")
	}

	//Attempt to convert the minSizeX to a float32 with error checking
	minSizeX, ok = minSizeX.(float32)
	if !ok {
		return nil, errors.New("invalid image minSizeX")
	}

	//Attempt to convert the minSizeY to a float with error checking
	minSizeY, ok = minSizeY.(float32)
	if !ok {
		return nil, errors.New("invalid image minSizeY")
	}

	//Attempt to convert the sizeX to a float with error checking
	sizeX, ok = sizeX.(float32)
	if !ok {
		return nil, errors.New("invalid image sizeX")
	}

	//Attempt to convert the sizeY to a float with error checking
	sizeY, ok = sizeY.(float32)
	if !ok {
		return nil, errors.New("invalid image sizeY")
	}
	//make sure the path ends in jpg, jpeg, or png
	if !strings.HasSuffix(path.(string), ".jpg") && !strings.HasSuffix(path.(string), ".jpeg") && !strings.HasSuffix(path.(string), ".png") {
		return nil, errors.New("invalid image path")
	}

	//load the image
	image := canvas.NewImageFromFile(path.(string))
	image.FillMode = canvas.ImageFillContain
	image.SetMinSize(fyne.NewSize(
		minSizeX.(float32),
		minSizeY.(float32),
	))

	image.Resize(fyne.NewSize(
		sizeX.(float32),
		sizeY.(float32),
	))

	return image, nil

}

func EntryHandler(properties map[string]interface{}, _ []*Widget, window fyne.Window) (fyne.CanvasObject, error) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(properties["PlaceHolder"].(string))
	entry.SetText(properties["Text"].(string))
	entry.OnChanged = func(s string) {
		_, _, err := NFFunction.Parse(window, properties["Action"].(string), map[string]interface{}{"Value": s})
		if err != nil {
			_, _, _ = NFFunction.Parse(window, "Error", map[string]interface{}{"message": err.Error()})
		}
	}
	return entry, nil
}

func NumEntryHandler(properties map[string]interface{}, _ []*Widget, window fyne.Window) (fyne.CanvasObject, error) {
	entryMin := properties["Min"].(float64)
	entryMax := properties["Max"].(float64)
	entry := widget.NewEntry()
	entry.SetPlaceHolder(properties["PlaceHolder"].(string))
	entry.SetText(properties["Text"].(string))
	entry.OnChanged = func(s string) {
		err := entry.Validate()
		if err != nil {
			//Clear the entry
			entry.SetText("")
			return
		}

		if properties["Action"] != nil {
			_, _, err = NFFunction.Parse(window, properties["Action"].(string), map[string]interface{}{"Value": s})
			if err != nil {
				_, _, _ = NFFunction.Parse(window, "Error", map[string]interface{}{"message": err.Error()})
			}
		}
	}
	//Validate the entry is a number
	entry.Validator = func(s string) error {
		//Check if the string can be cast to a float using a regex
		match, err := regexp.MatchString(`^-?\d*\.?\d*$`, s)
		if err != nil {
			return err
		}
		//Check if the number is within the min and max
		if match && entryMin < entryMax {
			//cast the string to a float
			num, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return err
			}
			if num < entryMin || num > entryMax {
				return errors.New("number out of range")
			}
		}

		if !match {
			return errors.New("invalid number")
		}
		return nil
	}
	return entry, nil
}

func PasswordEntryHandler(properties map[string]interface{}, _ []*Widget, window fyne.Window) (fyne.CanvasObject, error) {
	entry := widget.NewPasswordEntry()
	entry.SetPlaceHolder(properties["PlaceHolder"].(string))
	entry.SetText(properties["Text"].(string))
	if properties["Action"] != nil {
		entry.OnChanged = func(s string) {
			_, _, err := NFFunction.Parse(window, properties["Action"].(string), map[string]interface{}{"Value": s})
			if err != nil {
				_, _, _ = NFFunction.Parse(window, "Error", map[string]interface{}{"message": err.Error()})
			}
		}
	}
	return entry, nil
}

func SliderHandler(properties map[string]interface{}, _ []*Widget, window fyne.Window) (fyne.CanvasObject, error) {
	sMin := properties["Min"].(float64)
	sMax := properties["Max"].(float64)
	slider := widget.NewSlider(sMin, sMax)
	if properties["Action"] != nil {
		slider.OnChanged = func(f float64) {
			_, _, err := NFFunction.Parse(window, properties["Action"].(string), map[string]interface{}{"Value": f})
			if err != nil {
				_, _, _ = NFFunction.Parse(window, "Error", map[string]interface{}{"message": err.Error()})
			}
		}
	}
	return slider, nil
}
