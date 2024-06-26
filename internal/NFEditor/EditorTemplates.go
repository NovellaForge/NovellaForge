package NFEditor

import (
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFLayout"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFScene"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFWidget"
)

var MainMenuSceneTemplate = NFScene.New(
	"MainMenu", NFLayout.NewLayout(
		"VBox",
		NFLayout.NewChildren(
			NFWidget.New(
				"Label",
				NFWidget.NewChildren(),
				NFData.NewNFInterfaceMap(
					NFData.NewKeyVal("Text", "Main Menu"),
				),
			),
			NFWidget.New(
				"Button",
				NFWidget.NewChildren(),
				NFData.NewNFInterfaceMap(
					NFData.NewKeyVal("Text", "New Game"),
					NFData.NewKeyVal("OnTapped", "NewGame"),
					NFData.NewKeyVal("OnTappedArgs", NFData.NewNFInterfaceMap(
						NFData.NewKeyVal("NewGameScene", "NewGame"),
					)),
				),
			),
			NFWidget.New(
				"Button",
				NFWidget.NewChildren(),
				NFData.NewNFInterfaceMap(
					NFData.NewKeyVal("Text", "Load Game"),
					NFData.NewKeyVal("OnTapped", "LoadGame"),
				),
			),
			NFWidget.New(
				"Button",
				NFWidget.NewChildren(),
				NFData.NewNFInterfaceMap(
					NFData.NewKeyVal("Text", "Settings"),
					NFData.NewKeyVal("OnTapped", "Settings"),
				),
			),
			NFWidget.New(
				"Button",
				NFWidget.NewChildren(),
				NFData.NewNFInterfaceMap(
					NFData.NewKeyVal("Text", "Quit"),
					NFData.NewKeyVal("OnTapped", "Quit"),
				),
			),
		),
		NFData.NewNFInterfaceMap(),
	),
	NFData.NewNFInterfaceMap(),
)

var NewGameSceneTemplate = NFScene.New(
	"NewGame", NFLayout.NewLayout(
		"ExampleLayout",
		NFLayout.NewChildren(
			NFWidget.New(
				"ExampleWidget",
				NFWidget.NewChildren(),
				NFData.NewNFInterfaceMap(
					NFData.NewKeyVal("message", "Hello World"),
					NFData.NewKeyVal("action", "CustomFunction.ExampleFunction"),
				),
			),
		),
		NFData.NewNFInterfaceMap(),
	),
	NFData.NewNFInterfaceMap(),
)
