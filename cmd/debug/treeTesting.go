package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"log"
)

func main() {
	a := app.New()
	w := a.NewWindow("Hello")
	simpleData := map[string][]string{
		"":  {"1", "4"},
		"1": {"2", "3"},
		"4": {"5", "6"},
	}
	simpleValues := map[string]string{
		"1": "1",
		"2": "2",
		"3": "3",
		"4": "4",
		"5": "5",
		"6": "6",
	}
	simpleBinding := binding.BindStringTree(&simpleData, &simpleValues)
	simpleTree := widget.NewTreeWithData(simpleBinding,
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Branch")
		},
		func(dataItem binding.DataItem, branch bool, object fyne.CanvasObject) {
			strBind := dataItem.(binding.String)
			value, err := strBind.Get()
			if err != nil {
				log.Println(err)
				return
			}
			object.(*widget.Label).SetText(value)
		},
	)
	w.SetContent(simpleTree)
	w.ShowAndRun()

}
