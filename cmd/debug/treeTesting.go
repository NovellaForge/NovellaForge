package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"log"
	"strconv"
)

func main() {
	a := app.New()
	w := a.NewWindow("Hello")
	simpleData := map[string][]string{
		"":  {"1", "4"},
		"1": {"2", "3"},
		"4": {"5", "6"},
	}
	simpleValues := map[string]int{
		"1": 5,
		"2": 4,
		"3": 3,
		"4": 2,
		"5": 1,
		"6": 0,
	}

	simpleBinding := binding.BindIntTree(&simpleData, &simpleValues)
	simpleTree := widget.NewTreeWithData(simpleBinding,
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Branch")
		},
		func(dataItem binding.DataItem, branch bool, object fyne.CanvasObject) {
			strBind, ok := dataItem.(binding.String)
			if !ok {
				log.Println("DataItem is not a string")
			} else {
				log.Println("DataItem: ", strBind)
				value, err := strBind.Get()
				if err != nil {
					log.Println(err)
					return
				}
				log.Println("Value: ", value)
				object.(*widget.Label).SetText(value)
			}
			intBind, ok := dataItem.(binding.Int)
			if !ok {
				log.Println("DataItem is not an int")
			} else {
				log.Println("DataItem: ", intBind)
				value, err := intBind.Get()
				if err != nil {
					log.Println(err)
					return
				}
				log.Println("Value: ", value)
				object.(*widget.Label).SetText(strconv.Itoa(value))
			}
		},
	)

	w.SetContent(simpleTree)
	w.ShowAndRun()

}
