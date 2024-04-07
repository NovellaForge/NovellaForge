package main

import (
	"go.novellaforge.dev/novellaforge/internal/NFEditor"
)

func main() {
	_ = NFEditor.MainMenuSceneTemplate.Export()
	_ = NFEditor.NewGameSceneTemplate.Export()
}
