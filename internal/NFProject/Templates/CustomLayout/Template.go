package CustomLayout

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/pkg/NFLayout"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
	"log"
)

func init() {
	NFLayout.Register("ExampleLayout", ExampleLayoutHandler)
}

func Register() {
	//init is run when the package is imported, so this is just a dummy function to make sure the init function is run
	log.Printf("Registering ExampleLayouts")
}

func ExampleLayoutHandler(window fyne.Window, args map[string]interface{}, children []NFWidget.Widget) (fyne.CanvasObject, error) {
	vbox := container.NewVBox()
	vbox.Add(widget.NewLabel("Example Layout"))
	for _, child := range children {
		parsedChild, err := child.Parse(window)
		if err != nil {
			return nil, err
		}
		vbox.Add(parsedChild)
	}
	return vbox, nil
}
