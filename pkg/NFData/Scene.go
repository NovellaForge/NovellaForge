package NFData

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
	"log"
	"strconv"
)

type SceneData struct {
	sceneName binding.String
	scene     fyne.CanvasObject
	layouts   map[string]fyne.CanvasObject
	widgets   map[string]fyne.CanvasObject
	Variables NFInterface
}

func NewSceneData(sceneName string) {
	data := &SceneData{
		sceneName: binding.NewString(),
		scene:     nil,
		layouts:   make(map[string]fyne.CanvasObject),
		widgets:   make(map[string]fyne.CanvasObject),
		Variables: NewNFInterface(),
	}
	_ = data.sceneName.Set(sceneName)
	ActiveSceneData = data
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

func (s *SceneData) SetScene(scene fyne.CanvasObject) {
	s.scene = scene
}

func (s *SceneData) GetScene() fyne.CanvasObject {
	return s.scene
}

func (s *SceneData) AddLayout(layoutName string, layout fyne.CanvasObject) {
	// Check if the layout already exists
	if _, ok := s.layouts[layoutName]; ok {
		//Loop through adding a number to the end until it is unique
		for i := 1; ; i++ {
			newName := layoutName + "_" + strconv.Itoa(i)
			if _, ok = s.layouts[newName]; !ok {
				log.Println("Layout", layoutName, "already exists in scene", s.GetSceneName(), "\n adding suffix_", i)
				s.layouts[newName] = layout
				return
			}
		}
	}
}

func (s *SceneData) GetLayout(layoutName string) (fyne.CanvasObject, error) {
	if layout, ok := s.layouts[layoutName]; ok {
		return layout, nil
	}
	return nil, NFError.ErrKeyNotFound
}
