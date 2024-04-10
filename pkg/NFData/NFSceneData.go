package NFData

import (
	"fyne.io/fyne/v2/data/binding"
)

//TODO add a method to the NFScenes so that bindings can be initialized when a scene is parsed so that they don't need to be manually created on load

type NFSceneData struct {
	sceneName binding.String
	Layouts   *NFInterfaceMap
	Variables *NFInterfaceMap
	Bindings  *NFBindingMap
}

func NewSceneData(sceneName string) *NFSceneData {
	data := &NFSceneData{
		sceneName: binding.NewString(),
		Layouts:   NewNFInterfaceMap(),
		Variables: NewRemoteNFInterfaceMap(),
		Bindings:  NewNFBindingMap(),
	}
	_ = data.sceneName.Set(sceneName)
	return data
}

func (s *NFSceneData) SetSceneName(sceneName string) {
	err := s.sceneName.Set(sceneName)
	if err != nil {
		return
	}
}

func (s *NFSceneData) GetSceneName() string {
	name, err := s.sceneName.Get()
	if err != nil {
		return ""
	}
	return name
}

// GetSceneNameBinding returns the binding.String for the scene name
func (s *NFSceneData) GetSceneNameBinding() binding.String {
	return s.sceneName
}
