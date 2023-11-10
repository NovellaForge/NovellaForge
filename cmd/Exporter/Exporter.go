package main

import (
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget/CalsWidgets"
)

func main() {
	dialog := CalsWidgets.NewNarrativeBox(false, "Lorem Ipsum", "Dolor sit amet")
	err := dialog.Export()
	if err != nil {
		panic(err)
	}
}
