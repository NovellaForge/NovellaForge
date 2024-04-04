package NFData

import (
	"fyne.io/fyne/v2/data/binding"
)

type SceneData struct {
	sceneName binding.String
	Layouts   *NFInterfaceMap
	Variables *NFInterfaceMap
	Bindings  *NFBindingMap
}

func NewSceneData(sceneName string) *SceneData {
	data := &SceneData{
		sceneName: binding.NewString(),
		Layouts:   NewNFInterfaceMap(),
		Variables: NewRemoteNFInterfaceMap(),
		Bindings:  NewNFBindingMap(),
	}
	_ = data.sceneName.Set(sceneName)
	return data
}

func (s *SceneData) SetSceneName(sceneName string) {
	err := s.sceneName.Set(sceneName)
	if err != nil {
		return
	}
}

func (s *SceneData) GetSceneName() string {
	name, err := s.sceneName.Get()
	if err != nil {
		return ""
	}
	return name
}

// GetSceneNameBinding returns the binding.String for the scene name
func (s *SceneData) GetSceneNameBinding() binding.String {
	return s.sceneName
}
