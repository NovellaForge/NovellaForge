package NFEditor

import (
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout"
	"go.novellaforge.dev/novellaforge/pkg/NFScene"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget"
)

var MainMenuSceneTemplate = NFScene.Scene{
	Name: "MainMenu",
	Layout: NFLayout.Layout{
		Type: "VBox",
		Children: []NFWidget.Widget{
			{
				Type: "Label",
				Args: NFData.NewNFInterface(
					NFData.KeyVal{
						Key:   "Text",
						Value: "Main Menu",
					},
				),
			},
			{
				Type: "Button",
				Args: NFData.NewNFInterface(
					NFData.KeyVal{
						Key:   "Text",
						Value: "New Game",
					},
					NFData.KeyVal{
						Key:   "OnTapped",
						Value: "NewGame",
					},
				),
			},
			{
				Type: "Button",
				Args: NFData.NewNFInterface(
					NFData.KeyVal{
						Key:   "Text",
						Value: "Load Game",
					},
					NFData.KeyVal{
						Key:   "OnTapped",
						Value: "LoadGame",
					},
				),
			},
			{
				Type: "Button",
				Args: NFData.NewNFInterface(
					NFData.KeyVal{
						Key:   "Text",
						Value: "Settings",
					},
					NFData.KeyVal{
						Key:   "OnTapped",
						Value: "Settings",
					},
				),
			},
			{
				Type: "Button",
				Args: NFData.NewNFInterface(
					NFData.KeyVal{
						Key:   "Text",
						Value: "Quit",
					},
					NFData.KeyVal{
						Key:   "OnTapped",
						Value: "Quit",
					},
				),
			},
		},
	},
	Args: NFData.NewNFInterface(),
}

var NewGameSceneTemplate = NFScene.Scene{
	Name: "NewGame",
	Layout: NFLayout.Layout{
		Type: "ExampleLayout",
		Children: []NFWidget.Widget{
			{
				Type: "ExampleWidget",
				Args: NFData.NewNFInterface(
					NFData.KeyVal{
						Key:   "message",
						Value: "Hello World",
					},
					NFData.KeyVal{
						Key:   "action",
						Value: "CustomFunction.ExampleFunction",
					},
				),
			},
		},
	},
	Args: NFData.NewNFInterface(),
}
