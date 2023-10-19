package main

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget/CalsWidgets"
	"image/color"
	"strconv"
)

var dialogBox *CalsWidgets.Dialog
var multiImage *CalsWidgets.MultiImage

func main() {
	a := app.NewWithID("com.callial.novellaforge.testing")
	w := a.NewWindow("WidgetTesting")
	dialogBox = CalsWidgets.NewDialog([]string{})
	multiImage = CalsWidgets.NewMultiImage([]*canvas.Image{})

	//Listen for f5 to refresh the preview
	w.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		if event.Name == fyne.KeyF5 {
			RefreshPreview(w)
		}
	})

	RefreshPreview(w)
	w.ShowAndRun()
}

func RefreshPreview(w fyne.Window) {

	//if we didn't find the dialog box or image then return
	if dialogBox == nil || multiImage == nil {
		return
	}

	NumberValidator := func(s string) error {
		if s == "" {
			return nil
		}
		if _, err := strconv.ParseFloat(s, 32); err != nil {
			return errors.New("tab width must be a number")
		}
		return nil
	}

	contentEntry := widget.NewMultiLineEntry()
	contentEntry.Wrapping = fyne.TextWrapWord
	addMessageButton := widget.NewButton("Add Message", func() {
		tempContent := dialogBox.AllText
		tempContent = append(tempContent, contentEntry.Text)
		dialogBox.SetContent(tempContent)
	})
	setContentButton := widget.NewButton("Set Content", func() {
		dialogBox.SetContent([]string{contentEntry.Text})
	})
	contentEntry.SetPlaceHolder("Content")

	//Content Text Styling
	labelBold := widget.NewCheck("Content Bold", func(b bool) {
		dialogBox.ContentStyle().SetBold(b, dialogBox)
	})
	labelBold.SetChecked(dialogBox.ContentStyle().Bold)
	labelItalic := widget.NewCheck("Content Italic", func(b bool) {
		dialogBox.ContentStyle().SetItalic(b, dialogBox)
	})
	labelItalic.SetChecked(dialogBox.ContentStyle().Italic)
	labelMonoSpace := widget.NewCheck("Content MonoSpace", func(b bool) {
		dialogBox.ContentStyle().SetMonospace(b, dialogBox)
	})
	labelMonoSpace.SetChecked(dialogBox.ContentStyle().Monospace)
	labelSymbol := widget.NewCheck("ContentSymbol", func(b bool) {
		dialogBox.ContentStyle().SetSymbol(b, dialogBox)
	})
	labelSymbol.SetChecked(dialogBox.ContentStyle().Symbol)

	labelTabWidth := widget.NewEntry()
	labelTabWidth.SetPlaceHolder("Tab indent amount")
	labelTabWidth.OnChanged = func(s string) {
		if s == "" {
			dialogBox.ContentStyle().SetTabWidth(0, dialogBox)
			return
		}
		width := 0
		var err error
		if width, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.ContentStyle().SetTabWidth(width, dialogBox)
	}
	labelTabWidth.SetText(strconv.Itoa(dialogBox.ContentStyle().TabWidth))
	labelTabWidth.Validator = NumberValidator

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Name Plate")
	nameEntry.OnChanged = func(s string) {
		if nameEntry.Validate() != nil {
			return
		}
		dialogBox.SetName(s)
	}
	nameEntry.SetText("Dialog Name")
	toggleName := widget.NewButton("Toggle Name", func() {
		dialogBox.SetHasName(!dialogBox.HasName)
	})

	//Name Text Styling

	nameBold := widget.NewCheck("Name Bold", func(b bool) {
		dialogBox.NameStyle().SetBold(b, dialogBox)
	})
	nameBold.SetChecked(dialogBox.NameStyle().Bold)
	nameItalic := widget.NewCheck("Name Italic", func(b bool) {
		dialogBox.NameStyle().SetItalic(b, dialogBox)
	})
	nameItalic.SetChecked(dialogBox.NameStyle().Italic)
	nameMonoSpace := widget.NewCheck("Name Monospace", func(b bool) {
		dialogBox.NameStyle().SetMonospace(b, dialogBox)
	})
	nameMonoSpace.SetChecked(dialogBox.NameStyle().Monospace)
	nameSymbol := widget.NewCheck("Name Symbol", func(b bool) {
		dialogBox.NameStyle().SetSymbol(b, dialogBox)
	})
	nameSymbol.SetChecked(dialogBox.NameStyle().Symbol)

	nameTabWidth := widget.NewEntry()
	nameTabWidth.SetPlaceHolder("Tab indent amount")
	nameTabWidth.OnChanged = func(s string) {
		if s == "" {
			dialogBox.NameStyle().SetTabWidth(0, dialogBox)
			return
		}
		if nameTabWidth.Validate() != nil {
			return
		}
		width, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			nameTabWidth.SetValidationError(errors.New("tab width must be an int"))
			return
		}
		dialogBox.NameStyle().SetTabWidth(int(width), dialogBox)
	}

	namePositionX := widget.NewEntry()
	namePositionX.SetPlaceHolder("X")
	namePositionX.OnChanged = func(s string) {
		curPosition := dialogBox.NamePosition
		if s == "" {
			curPosition.X = 0
			dialogBox.SetNamePosition(curPosition)
			return
		}
		x := 0
		var err error
		if x, err = strconv.Atoi(s); err != nil {
			return
		}
		curPosition.X = float32(x)
		dialogBox.SetNamePosition(curPosition)
	}
	namePositionY := widget.NewEntry()
	namePositionY.OnChanged = func(s string) {
		curPosition := dialogBox.NamePosition
		if s == "" {
			curPosition.Y = 0
			dialogBox.SetNamePosition(curPosition)
			return
		}
		y := 0
		var err error
		if y, err = strconv.Atoi(s); err != nil {
			return
		}
		curPosition.Y = float32(y)
		dialogBox.SetNamePosition(curPosition)
	}
	namePositionY.SetPlaceHolder("Y")
	namePositionX.Validator = NumberValidator
	namePositionY.Validator = NumberValidator
	nameAffectsLayout := widget.NewCheck("Name Affects Layout", func(b bool) {
		dialogBox.SetNameAffectsLayout(b)
	})
	nameAffectsLayout.SetChecked(dialogBox.NameAffectsLayout)

	//Content Sizing
	sizingMinWidth := widget.NewEntry()
	sizingMinWidth.SetPlaceHolder("Min Width")
	sizingMinWidth.OnChanged = func(s string) {
		if s == "" {
			dialogBox.Sizing().SetMinWidth(0, dialogBox)
			return
		}
		if sizingMinWidth.Validate() != nil {
			return
		}
		float, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return
		}
		dialogBox.Sizing().SetMinWidth(float32(float), dialogBox)
	}
	sizingMinHeight := widget.NewEntry()
	sizingMinHeight.SetPlaceHolder("Min Height")
	sizingMinHeight.OnChanged = func(s string) {
		if s == "" {
			dialogBox.Sizing().SetMinHeight(0, dialogBox)
			return
		}
		if sizingMinHeight.Validate() != nil {
			return
		}
		float, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return
		}
		dialogBox.Sizing().SetMinHeight(float32(float), dialogBox)
	}
	sizingMaxWidth := widget.NewEntry()
	sizingMaxWidth.SetPlaceHolder("Max Width")
	sizingMaxWidth.OnChanged = func(s string) {
		if s == "" {
			dialogBox.Sizing().SetMaxWidth(0, dialogBox)
			return
		}
		if sizingMaxWidth.Validate() != nil {
			return
		}
		float, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return
		}
		dialogBox.Sizing().SetMaxWidth(float32(float), dialogBox)
	}
	sizingMaxHeight := widget.NewEntry()
	sizingMaxHeight.SetPlaceHolder("Max Height")
	sizingMaxHeight.OnChanged = func(s string) {
		if s == "" {
			dialogBox.Sizing().SetMaxHeight(0, dialogBox)
			return
		}
		if sizingMaxHeight.Validate() != nil {
			return
		}
		float, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return
		}
		dialogBox.Sizing().SetMaxHeight(float32(float), dialogBox)
	}
	sizingFitWidth := widget.NewCheck("Fit Width", func(b bool) {
		dialogBox.Sizing().SetFitWidth(b, dialogBox)
	})
	sizingFitWidth.SetChecked(dialogBox.Sizing().FitWidth)
	sizingFitHeight := widget.NewCheck("Fit Height", func(b bool) {
		dialogBox.Sizing().SetFitHeight(b, dialogBox)
	})
	sizingFitHeight.SetChecked(dialogBox.Sizing().FitHeight)

	sizingMinWidth.Validator = NumberValidator
	sizingMinHeight.Validator = NumberValidator
	sizingMaxWidth.Validator = NumberValidator
	sizingMaxHeight.Validator = NumberValidator

	//Name Sizing
	nameSizingMaxWidth := widget.NewEntry()
	nameSizingMaxWidth.SetPlaceHolder("Max Width")
	nameSizingMaxWidth.OnChanged = func(s string) {
		if s == "" {
			dialogBox.NameSizing().SetMaxWidth(0, dialogBox)
			return
		}
		width := 0
		var err error
		if width, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.NameSizing().SetMaxWidth(float32(width), dialogBox)
	}
	nameSizingMaxHeight := widget.NewEntry()
	nameSizingMaxHeight.SetPlaceHolder("Max Height")
	nameSizingMaxHeight.OnChanged = func(s string) {
		if s == "" {
			dialogBox.NameSizing().SetMaxHeight(0, dialogBox)
			return
		}
		height := 0
		var err error
		if height, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.NameSizing().SetMaxHeight(float32(height), dialogBox)
	}
	nameSizingMinWidth := widget.NewEntry()
	nameSizingMinWidth.SetPlaceHolder("Min Width")
	nameSizingMinWidth.OnChanged = func(s string) {
		if s == "" {
			dialogBox.NameSizing().SetMinWidth(0, dialogBox)
			return
		}
		width := 0
		var err error
		if width, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.NameSizing().SetMinWidth(float32(width), dialogBox)
	}
	nameSizingMinHeight := widget.NewEntry()
	nameSizingMinHeight.SetPlaceHolder("Min Height")
	nameSizingMinHeight.OnChanged = func(s string) {
		if s == "" {
			dialogBox.NameSizing().SetMinHeight(0, dialogBox)
			return
		}
		height := 0
		var err error
		if height, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.NameSizing().SetMinHeight(float32(height), dialogBox)
	}
	nameSizingMaxWidth.Validator = NumberValidator
	nameSizingMaxHeight.Validator = NumberValidator
	nameSizingMinWidth.Validator = NumberValidator
	nameSizingMinHeight.Validator = NumberValidator

	//Content Padding
	paddingTop := widget.NewEntry()
	paddingTop.SetPlaceHolder("Top Padding")
	paddingTop.OnChanged = func(s string) {
		if s == "" {
			dialogBox.Padding().SetTop(0, dialogBox)
			return
		}
		top := 0
		var err error
		if top, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.Padding().SetTop(float32(top), dialogBox)
	}
	paddingBottom := widget.NewEntry()
	paddingBottom.SetPlaceHolder("Bottom Padding")
	paddingBottom.OnChanged = func(s string) {
		if s == "" {
			dialogBox.Padding().SetBottom(0, dialogBox)
			return
		}
		bottom := 0
		var err error
		if bottom, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.Padding().SetBottom(float32(bottom), dialogBox)
	}
	paddingLeft := widget.NewEntry()
	paddingLeft.SetPlaceHolder("Left Padding")
	paddingLeft.OnChanged = func(s string) {
		if s == "" {
			dialogBox.Padding().SetLeft(0, dialogBox)
			return
		}
		left := 0
		var err error
		if left, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.Padding().SetLeft(float32(left), dialogBox)
	}
	paddingRight := widget.NewEntry()
	paddingRight.SetPlaceHolder("Right Padding")
	paddingRight.OnChanged = func(s string) {}

	paddingTop.Validator = NumberValidator
	paddingBottom.Validator = NumberValidator
	paddingLeft.Validator = NumberValidator
	paddingRight.Validator = NumberValidator

	//Name Padding
	namePaddingTop := widget.NewEntry()
	namePaddingTop.SetPlaceHolder("Name Top Padding")
	namePaddingTop.OnChanged = func(s string) {
		if s == "" {
			dialogBox.NamePadding().SetTop(0, dialogBox)
			return
		}
		top := 0
		var err error
		if top, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.NamePadding().SetTop(float32(top), dialogBox)
	}
	namePaddingBottom := widget.NewEntry()
	namePaddingBottom.SetPlaceHolder("Name Bottom Padding")
	namePaddingBottom.OnChanged = func(s string) {
		if s == "" {
			dialogBox.NamePadding().SetBottom(0, dialogBox)
			return
		}
		bottom := 0
		var err error
		if bottom, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.NamePadding().SetBottom(float32(bottom), dialogBox)
	}
	namePaddingLeft := widget.NewEntry()
	namePaddingLeft.SetPlaceHolder("Name Left Padding")
	namePaddingLeft.OnChanged = func(s string) {
		if s == "" {
			dialogBox.NamePadding().SetLeft(0, dialogBox)
			return
		}
		left := 0
		var err error
		if left, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.NamePadding().SetLeft(float32(left), dialogBox)
	}
	namePaddingRight := widget.NewEntry()
	namePaddingRight.SetPlaceHolder("Name Right Padding")
	namePaddingRight.OnChanged = func(s string) {}

	namePaddingTop.Validator = NumberValidator
	namePaddingBottom.Validator = NumberValidator
	namePaddingLeft.Validator = NumberValidator
	namePaddingRight.Validator = NumberValidator
	//External Padding
	externalPaddingTop := widget.NewEntry()
	externalPaddingTop.SetPlaceHolder("Top")
	externalPaddingTop.OnChanged = func(s string) {
		if s == "" {
			dialogBox.ExternalPadding().SetTop(0, dialogBox)
			return
		}
		top := 0
		var err error
		if top, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.ExternalPadding().SetTop(float32(top), dialogBox)
	}
	externalPaddingBottom := widget.NewEntry()
	externalPaddingBottom.SetPlaceHolder("Bottom")
	externalPaddingBottom.OnChanged = func(s string) {
		if s == "" {
			dialogBox.ExternalPadding().SetBottom(0, dialogBox)
			return
		}
		bottom := 0
		var err error
		if bottom, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.ExternalPadding().SetBottom(float32(bottom), dialogBox)
	}
	externalPaddingLeft := widget.NewEntry()
	externalPaddingLeft.SetPlaceHolder("Left")
	externalPaddingLeft.OnChanged = func(s string) {
		if s == "" {
			dialogBox.ExternalPadding().SetLeft(0, dialogBox)
			return
		}
		left := 0
		var err error
		if left, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.ExternalPadding().SetLeft(float32(left), dialogBox)
	}
	externalPaddingRight := widget.NewEntry()
	externalPaddingRight.SetPlaceHolder("Right")
	externalPaddingRight.OnChanged = func(s string) {
		if s == "" {
			dialogBox.ExternalPadding().SetRight(0, dialogBox)
			return
		}
		right := 0
		var err error
		if right, err = strconv.Atoi(s); err != nil {
			return
		}
		dialogBox.ExternalPadding().SetRight(float32(right), dialogBox)
	}

	externalPaddingTop.Validator = NumberValidator
	externalPaddingBottom.Validator = NumberValidator
	externalPaddingLeft.Validator = NumberValidator
	externalPaddingRight.Validator = NumberValidator

	//Colors
	nameStrokeColorPicker := dialog.NewColorPicker("", "", func(c color.Color) { dialogBox.SetNameStrokeColor(c) }, w)
	nameFillColorPicker := dialog.NewColorPicker("", "", func(c color.Color) { dialogBox.SetNameFill(c) }, w)
	strokeColorPicker := dialog.NewColorPicker("", "", func(c color.Color) { dialogBox.SetStrokeColor(c) }, w)
	fillColorPicker := dialog.NewColorPicker("", "", func(c color.Color) { dialogBox.SetFill(c) }, w)
	setNameStrokeColorButton := widget.NewButton("Set Name Stroke Color", func() { nameStrokeColorPicker.Show() })
	setNameFillColorButton := widget.NewButton("Set Name Fill Color", func() { nameFillColorPicker.Show() })
	setStrokeColorButton := widget.NewButton("Set Stroke Color", func() { strokeColorPicker.Show() })
	setFillColorButton := widget.NewButton("Set Fill Color", func() { fillColorPicker.Show() })

	//Image Stuff
	indexEntry := widget.NewEntry()
	indexEntry.SetPlaceHolder("Index")
	indexEntry.OnChanged = func(s string) {
		if s == "" {
			return
		}
		index, err := strconv.Atoi(s)
		if err != nil {
			return
		}
		multiImage.SetIndex(index)
	}
	indexEntry.Validator = NumberValidator
	indexEntry.SetText(strconv.Itoa(multiImage.Index))

	vbox := container.NewVBox(
		widget.NewLabelWithStyle("Image", fyne.TextAlignCenter, fyne.TextStyle{}),
		indexEntry,
		widget.NewButton("Add Image", func() {
			dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					return
				}
				if reader == nil {
					return
				}
				//Check if the file is a valid image file ending in png or jpg
				if !(reader.URI().Extension() == ".png" || reader.URI().Extension() == ".jpg") {
					dialog.ShowError(errors.New("file must be a valid image file"), w)
					return
				}
				image := canvas.NewImageFromURI(reader.URI())
				multiImage.AddImage(image)
			}, w)
		}),
		container.NewGridWithColumns(2,
			widget.NewButton("Next Index", func() {
				multiImage.NextIndex()
				indexEntry.SetText(strconv.Itoa(multiImage.Index))
			}),
			widget.NewButton("Previous Index", func() {
				multiImage.PrevIndex()
				indexEntry.SetText(strconv.Itoa(multiImage.Index))
			}),
			widget.NewButton("Remove Image", func() {
				dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil {
						return
					}
					if reader == nil {
						return
					}
					//Check if the file is a valid image file ending in png or jpg
					if !(reader.URI().Extension() == ".png" || reader.URI().Extension() == ".jpg") {
						dialog.ShowError(errors.New("file must be a valid image file"), w)
						return
					}
					image := canvas.NewImageFromURI(reader.URI())
					multiImage.RemoveImageByImage(image, true)
				}, w)
			}),
			widget.NewButton("Clear Images", func() {
				multiImage.ClearImages()
				indexEntry.SetText("0")
			}),
			widget.NewButton("Remove Index", func() {
				if indexEntry.Validate() != nil {
					return
				}
				index, err := strconv.Atoi(indexEntry.Text)
				if err != nil {
					return
				}
				multiImage.RemoveImage(index)
			}),
			widget.NewButton("Insert Image", func() {
				dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil {
						return
					}
					if reader == nil {
						return
					}
					//Check if the file is a valid image file ending in png or jpg
					if !(reader.URI().Extension() == ".png" || reader.URI().Extension() == ".jpg") {
						dialog.ShowError(errors.New("file must be a valid image file"), w)
						return
					}
					image := canvas.NewImageFromURI(reader.URI())
					multiImage.InsertImage(multiImage.Index, image)
				}, w)
			}),
		),
		widget.NewLabelWithStyle("Dialog Text", fyne.TextAlignCenter, fyne.TextStyle{}),
		contentEntry,
		addMessageButton,
		setContentButton,
		setStrokeColorButton,
		setFillColorButton,
		widget.NewLabelWithStyle("Dialog Text Styling", fyne.TextAlignCenter, fyne.TextStyle{}),
		container.NewGridWithColumns(2,
			labelBold,
			labelItalic,
			labelMonoSpace,
			labelSymbol,
			labelTabWidth,
		),
		widget.NewLabelWithStyle("Dialog Sizing", fyne.TextAlignCenter, fyne.TextStyle{}),
		container.NewGridWithColumns(2,
			sizingMinWidth,
			sizingMaxWidth,
			sizingMinHeight,
			sizingMaxHeight,
			sizingFitWidth,
			sizingFitHeight,
		),
		widget.NewLabelWithStyle("Dialog Internal Padding", fyne.TextAlignCenter, fyne.TextStyle{}),
		container.NewGridWithColumns(2,
			paddingTop,
			paddingBottom,
			paddingLeft,
			paddingRight,
		),
		widget.NewLabelWithStyle("Dialog External Padding", fyne.TextAlignCenter, fyne.TextStyle{}),
		container.NewGridWithColumns(2,
			externalPaddingTop,
			externalPaddingBottom,
			externalPaddingLeft,
			externalPaddingRight,
		),
		widget.NewLabelWithStyle("Dialog Name Plate Configuration", fyne.TextAlignCenter, fyne.TextStyle{}),
		nameEntry,
		toggleName,
		setNameStrokeColorButton,
		setNameFillColorButton,
		widget.NewLabelWithStyle("Name Plate Positioning", fyne.TextAlignCenter, fyne.TextStyle{}),
		container.NewGridWithColumns(2,
			namePositionX,
			namePositionY,
			nameAffectsLayout,
		),
		widget.NewLabelWithStyle("Name Plate Styling", fyne.TextAlignCenter, fyne.TextStyle{}),
		container.NewGridWithColumns(2,
			nameBold,
			nameItalic,
			nameMonoSpace,
			nameSymbol,
		),
		widget.NewLabelWithStyle("Name Plate Sizing", fyne.TextAlignCenter, fyne.TextStyle{}),
		container.NewGridWithColumns(2,
			nameSizingMinWidth,
			nameSizingMaxWidth,
			nameSizingMinHeight,
			nameSizingMaxHeight,
		),
		widget.NewLabelWithStyle("Name Plate Internal Padding", fyne.TextAlignCenter, fyne.TextStyle{}),
		container.NewGridWithColumns(2,
			namePaddingTop,
			namePaddingBottom,
			namePaddingLeft,
			namePaddingRight,
		),
	)

	scroll := container.NewScroll(vbox)
	scroll.SetMinSize(fyne.NewSize(400, 0))
	border := container.NewBorder(
		widget.NewLabelWithStyle("Press F5 or resize the window to refresh layout changes to the preview", fyne.TextAlignCenter, fyne.TextStyle{}),
		nil,
		scroll,
		nil,
		container.NewBorder(nil, dialogBox, nil, nil, multiImage),
	)

	w.SetContent(border)
}
