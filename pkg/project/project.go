package project

import "encoding/json"

type Content struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type Scene struct {
	ID          string    `json:"id"`
	SceneType   string    `json:"sceneType"`
	ContentList []Content `json:"contentList"`
}

type Project struct {
	GameName string  `json:"gameName"`
	Version  string  `json:"version"`
	Author   string  `json:"author"`
	Credits  string  `json:"credits"`
	Scenes   []Scene `json:"scenes"`
}

// SerializeProject Serialize Project to JSON
func SerializeProject(project Project) (string, error) {
	jsonData, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// DeserializeProject Deserialize JSON to Project
func DeserializeProject(jsonData string) (Project, error) {
	var project Project
	err := json.Unmarshal([]byte(jsonData), &project)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}
