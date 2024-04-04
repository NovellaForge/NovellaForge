package main

import (
	"fyne.io/fyne/v2/app"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout/DefaultLayouts"
	"go.novellaforge.dev/novellaforge/pkg/NFScene"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget/DefaultWidgets"
	"log"
	"time"
)

func main() {
	DefaultLayouts.Import()
	DefaultWidgets.Import()

	err := NFData.GlobalBindings.CreateBinding("Time", time.Now().Format(time.RFC3339))
	if err != nil {
		panic(err)
	}
	timeBinding, err := NFData.GlobalBindings.GetStringBinding("Time")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			timeErr := timeBinding.Set(time.Now().Format(time.RFC3339))
			if timeErr != nil {
				return
			}
			time.Sleep(time.Second)
		}
	}()
	NFData.GlobalVars.Set("TestText", "Global Text")

	testScene := NFScene.NewScene("TestScene",
		NFLayout.NewLayout(
			"VBox",
			NFWidget.NewChildren(
				NFWidget.NewWithID(
					"MainLabelNoRef",
					"Label",
					nil,
					NFData.NewNFInterfaceMap(
						NFData.NewKeyVal("Text", "No Ref Text"),
					),
				),
				NFWidget.NewWithID(
					"MainLabelGlobalRef",
					"Label",
					nil,
					NFData.NewNFInterfaceMap(
						NFData.NewKeyVal("Text", NFData.NewRef(NFData.NFRefGlobal, "TestText")),
					),
				),
				NFWidget.NewWithID(
					"MainLabelGlobalBinding",
					"Label",
					nil,
					NFData.NewNFInterfaceMap(
						NFData.NewKeyVal("Text", NFData.NewRefWithBinding(NFData.NFRefGlobal, "Time", NFData.BindingString)),
					),
				),
			),
			NFData.NewNFInterfaceMap(),
		),
		NFData.NewNFInterfaceMap(),
	)

	err = testScene.Export(true)
	if err != nil {
		log.Println(err)
	}

	a := app.New()
	w := a.NewWindow("Hello")
	parse, err := testScene.Parse(w)
	if err != nil {
		panic(err)
	}
	w.SetContent(parse)
	w.ShowAndRun()
}
