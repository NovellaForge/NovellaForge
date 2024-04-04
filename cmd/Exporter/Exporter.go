package main

import (
	"go.novellaforge.dev/novellaforge/internal/NFEditor"
)

func main() {
	NFEditor.MainMenuSceneTemplate.Export(true)
	NFEditor.NewGameSceneTemplate.Export(true)
}
