package NFEditor

import (
	"github.com/NovellaForge/NovellaForge/pkg/NFLayout"
	"github.com/NovellaForge/NovellaForge/pkg/NFScene"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
)

var MainMenuSceneTemplate = NFScene.Scene{
	Name: "MainMenu",
	Layout: NFLayout.Layout{
		Type: "VBox",
		Children: []NFWidget.Widget{
			{
				Type: "Label",
				Properties: map[string]interface{}{
					"Text": "Main Menu",
				},
			},
			{
				Type: "Button",
				Properties: map[string]interface{}{
					"Text":   "New Game",
					"Action": "NewGame",
				},
			},
			{
				Type: "Button",
				Properties: map[string]interface{}{
					"Text":   "Load Game",
					"Action": "LoadGame",
				},
			},
			{
				Type: "Button",
				Properties: map[string]interface{}{
					"Text":   "Settings",
					"Action": "Settings",
				},
			},
			{
				Type: "Button",
				Properties: map[string]interface{}{
					"Text":   "Quit",
					"Action": "Quit",
				},
			},
		},
	},
	Properties: nil,
}

var NewGameSceneTemplate = NFScene.Scene{
	Name: "NewGame",
	Layout: NFLayout.Layout{
		Type: "VBox",
	},
}
