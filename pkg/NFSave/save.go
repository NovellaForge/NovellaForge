package NFSave

import (
	"encoding/json"
	"errors"
	"github.com/NovellaForge/NovellaForge/pkg/NFEncryption"
	"os"
	"path/filepath"
	"time"
)

const Extension = ".novella" // This is the extension that will be used for save files make sure it has the dot at the beginning
var Active *Save
var Directory = ""
var ErrSaveNameNotSet = errors.New("save name not yet set")
var SaveEncryption = false
var SaveHistory = true
var SaveEncryptionKey = "NovellaForge"

type Save struct {
	Name       string             `json:"Name"`
	Scene      string             `json:"Scene"`
	Time       time.Time          `json:"Saved At"`
	IntData    map[string]int     `json:"IntData,omitempty"`
	FloatData  map[string]float64 `json:"FloatData,omitempty"`
	StringData map[string]string  `json:"StringData,omitempty"`
	BoolData   map[string]bool    `json:"BoolData,omitempty"`
}

// GetActive and SetActive are used to get and set the active save
func GetActive() *Save {
	return Active
}

// SetActive is used to set the active save
func SetActive(save *Save) {
	Active = save
}

// GetSaveEncryption and SetSaveEncryption are used to get and set the save encryption
func GetSaveEncryption() bool {
	return SaveEncryption
}

// SetSaveEncryption is used to set the save encryption
func SetSaveEncryption(value bool) {
	SaveEncryption = value
}

// GetSaveHistory and SetSaveHistory are used to get and set the save history
func GetSaveHistory() bool {
	return SaveHistory
}

// SetSaveHistory is used to set the save history
func SetSaveHistory(value bool) {
	SaveHistory = value
}

// GetSaveEncryptionKey and SetSaveEncryptionKey are used to get and set the save encryption key
func GetSaveEncryptionKey() string {
	return SaveEncryptionKey
}

// SetSaveEncryptionKey is used to set the save encryption key
func SetSaveEncryptionKey(value string) {
	SaveEncryptionKey = value
}

// New creates a new save file
func New(scene string) (*Save, error) {
	//Create the save struct
	save := Save{
		Name:       "",
		Scene:      scene,
		Time:       time.Now(),
		IntData:    map[string]int{},
		FloatData:  map[string]float64{},
		StringData: map[string]string{},
		BoolData:   map[string]bool{},
	}

	return &save, nil
}

// Load loads a save file
func Load(savePath string) (*Save, error) {
	//Check if it is a valid save file
	if filepath.Ext(savePath) != Extension && filepath.Ext(savePath) != Extension+"history" {
		return nil, errors.New("invalid save file")
	}
	//Try to load the file at the path
	fileBytes, err := os.ReadFile(savePath)
	if err != nil {
		return nil, err
	}
	//Decrypt the fileBytes if needed
	if SaveEncryption {
		fileBytes, err = NFEncryption.Decrypt(fileBytes, SaveEncryptionKey)
		if err != nil {
			return nil, err
		}
	}
	//Decode the fileBytes into a save struct
	save := Save{}
	err = json.Unmarshal(fileBytes, &save)
	if err != nil {
		return nil, err
	}

	return &save, nil
}

func (s *Save) Save() error {
	err := os.MkdirAll(Directory, os.ModePerm)
	if err != nil {
		return err
	}

	//Check if the save file has a name
	if s.Name == "" {
		return ErrSaveNameNotSet
	}

	saveIsNew := false
	//Check if the save already exists
	var SaveFile *os.File
	_, err = os.Stat(Directory + s.Name + "/save" + Extension)
	if err != nil {
		//If the save does not exist, create a new save folder
		err = os.MkdirAll(Directory+s.Name, os.ModePerm)
		if err != nil {
			return err
		}
		saveIsNew = true
		//Create a new save file
		SaveFile, err = os.Create(Directory + s.Name + "/save" + Extension)
		if err != nil {
			return err
		}
	} else {
		SaveFile, err = os.Open(Directory + s.Name + "/save" + Extension)
		if err != nil {
			return err
		}
	}
	defer SaveFile.Close()

	if SaveHistory {
		//Create a folder with the save name with a history folder inside
		err = os.MkdirAll(Directory+s.Name+"/history", os.ModePerm)
		if err != nil {
			return err
		}
		if !saveIsNew {
			//If the save file exists, count the number of history files in the history folder
			count := 0
			//Walk the history folder
			err = filepath.WalkDir(Directory+s.Name+"/history", func(path string, info os.DirEntry, err error) error {
				//If the file is a json file, increment the count
				if filepath.Ext(path) == Extension+"history" {
					count++
				}
				return nil
			})
			if err != nil {
				return err
			}
			//Rename the save file to a history file
			newName := Directory + s.Name + "/history/save" + string(rune(count+1)) + Extension + "history"

			//get the current save files data and move it to a history file before saving the new save file
			oldSave := Save{}
			//Decode the save file into the oldSave struct
			err = json.NewDecoder(SaveFile).Decode(&oldSave)
			if err != nil {
				return err
			}

			var fileBytes []byte
			if SaveEncryption {
				//Convert oldSave into a byte array
				var oldSaveBytes []byte
				oldSaveBytes, err = json.Marshal(oldSave)
				if err != nil {
					return err
				}
				//Encrypt the byte array
				fileBytes, err = NFEncryption.Encrypt(oldSaveBytes, SaveEncryptionKey)
				if err != nil {
					return err
				}
			} else {
				//Just marshal the bytes with indentation
				fileBytes, err = json.MarshalIndent(oldSave, "", "    ")
				if err != nil {
					return err
				}
			}

			newHistoryFile, err := os.Create(newName)
			if err != nil {
				return err
			}

			//Write the file bytes to the new history file
			_, err = newHistoryFile.Write(fileBytes)
			if err != nil {
				return err
			}
		}
	}
	var saveBytes []byte
	if SaveEncryption {
		//Convert the save struct into a byte array
		saveBytes, err = json.Marshal(s)
		if err != nil {
			return err
		}
		//Encrypt the byte array
		saveBytes, err = NFEncryption.Encrypt(saveBytes, SaveEncryptionKey)
		if err != nil {
			return err
		}
	} else {
		//Just marshal the bytes with indentation
		saveBytes, err = json.MarshalIndent(s, "", "    ")
		if err != nil {
			return err
		}
	}
	//Write the save bytes to the save file
	_, err = SaveFile.Write(saveBytes)
	if err != nil {
		return err
	}
	return nil
}

// SetSaveName is used to set the save name
func (s *Save) SetSaveName(name string) {
	s.Name = name
}

// GetSaveName is used to get the save name
func (s *Save) GetSaveName() string {
	return s.Name
}

// SetScene is used to set the scene
func (s *Save) SetScene(scene string) {
	s.Scene = scene
}

// GetScene is used to get the scene
func (s *Save) GetScene() string {
	return s.Scene
}

// SetInt is used to set an int value in the save file
func (s *Save) SetInt(key string, value int) {
	s.IntData[key] = value
}

// SetFloat is used to set a float value in the save file
func (s *Save) SetFloat(key string, value float64) {
	s.FloatData[key] = value
}

// SetString is used to set a string value in the save file
func (s *Save) SetString(key string, value string) {
	s.StringData[key] = value
}

// SetBool is used to set a bool value in the save file
func (s *Save) SetBool(key string, value bool) {
	s.BoolData[key] = value
}

// GetInt is used to get an int value from the save file
func (s *Save) GetInt(key string) (int, error) {
	if value, ok := s.IntData[key]; ok {
		return value, nil
	}
	return 0, errors.New("key not found")
}

// GetFloat is used to get a float value from the save file
func (s *Save) GetFloat(key string) (float64, error) {
	if value, ok := s.FloatData[key]; ok {
		return value, nil
	}
	return 0, errors.New("key not found")
}

// GetString is used to get a string value from the save file
func (s *Save) GetString(key string) (string, error) {
	if value, ok := s.StringData[key]; ok {
		return value, nil
	}
	return "", errors.New("key not found")
}

// GetBool is used to get a bool value from the save file
func (s *Save) GetBool(key string) (bool, error) {
	if value, ok := s.BoolData[key]; ok {
		return value, nil
	}
	return false, errors.New("key not found")
}

// DeleteInt is used to delete an int value from the save file
func (s *Save) DeleteInt(key string) {
	delete(s.IntData, key)
}

// DeleteFloat is used to delete a float value from the save file
func (s *Save) DeleteFloat(key string) {
	delete(s.FloatData, key)
}

// DeleteString is used to delete a string value from the save file
func (s *Save) DeleteString(key string) {
	delete(s.StringData, key)
}

// DeleteBool is used to delete a bool value from the save file
func (s *Save) DeleteBool(key string) {
	delete(s.BoolData, key)
}

// SafeDeleteInt is used to delete an int value from the save file (This method will return an error if the key does not exist unlike DeleteInt)
func (s *Save) SafeDeleteInt(key string) error {
	if _, ok := s.IntData[key]; ok {
		delete(s.IntData, key)
		return nil
	}
	return errors.New("key not found")
}

// SafeDeleteFloat is used to delete a float value from the save file (This method will return an error if the key does not exist unlike DeleteFloat)
func (s *Save) SafeDeleteFloat(key string) error {
	if _, ok := s.FloatData[key]; ok {
		delete(s.FloatData, key)
		return nil
	}
	return errors.New("key not found")
}

// SafeDeleteString is used to delete a string value from the save file (This method will return an error if the key does not exist unlike DeleteString)
func (s *Save) SafeDeleteString(key string) error {
	if _, ok := s.StringData[key]; ok {
		delete(s.StringData, key)
		return nil
	}
	return errors.New("key not found")
}

// SafeDeleteBool is used to delete a bool value from the save file (This method will return an error if the key does not exist unlike DeleteBool)
func (s *Save) SafeDeleteBool(key string) error {
	if _, ok := s.BoolData[key]; ok {
		delete(s.BoolData, key)
		return nil
	}
	return errors.New("key not found")
}

// DeleteAll is used to delete all values from the save file
func (s *Save) DeleteAll() {
	s.IntData = map[string]int{}
	s.FloatData = map[string]float64{}
	s.StringData = map[string]string{}
	s.BoolData = map[string]bool{}
}

// DeleteAllInt is used to delete all int values from the save file
func (s *Save) DeleteAllInt() {
	s.IntData = map[string]int{}
}

// DeleteAllFloat is used to delete all float values from the save file
func (s *Save) DeleteAllFloat() {
	s.FloatData = map[string]float64{}
}

// DeleteAllString is used to delete all string values from the save file
func (s *Save) DeleteAllString() {
	s.StringData = map[string]string{}
}

// DeleteAllBool is used to delete all bool values from the save file
func (s *Save) DeleteAllBool() {
	s.BoolData = map[string]bool{}
}

// UpdateInt is used to update an int value in the save file (This method will return an error if the key does not exist unlike SetInt)
func (s *Save) UpdateInt(key string, value int) error {
	if _, ok := s.IntData[key]; ok {
		s.IntData[key] = value
		return nil
	}
	return errors.New("key not found")
}

// UpdateFloat is used to update a float value in the save file (This method will return an error if the key does not exist unlike SetFloat)
func (s *Save) UpdateFloat(key string, value float64) error {
	if _, ok := s.FloatData[key]; ok {
		s.FloatData[key] = value
		return nil
	}
	return errors.New("key not found")
}

// UpdateString is used to update a string value in the save file (This method will return an error if the key does not exist unlike SetString)
func (s *Save) UpdateString(key string, value string) error {
	if _, ok := s.StringData[key]; ok {
		s.StringData[key] = value
		return nil
	}
	return errors.New("key not found")
}

// UpdateBool is used to update a bool value in the save file (This method will return an error if the key does not exist unlike SetBool)
func (s *Save) UpdateBool(key string, value bool) error {
	if _, ok := s.BoolData[key]; ok {
		s.BoolData[key] = value
		return nil
	}
	return errors.New("key not found")
}

// GetInts is used to get all int values from the save file
func (s *Save) GetInts() map[string]int {
	return s.IntData
}

// GetFloats is used to get all float values from the save file
func (s *Save) GetFloats() map[string]float64 {
	return s.FloatData
}

// GetStrings is used to get all string values from the save file
func (s *Save) GetStrings() map[string]string {
	return s.StringData
}

// GetBools is used to get all bool values from the save file
func (s *Save) GetBools() map[string]bool {
	return s.BoolData
}

// GetIntKeys is used to get all int keys from the save file
func (s *Save) GetIntKeys() []string {
	var keys []string
	for key := range s.IntData {
		keys = append(keys, key)
	}
	return keys
}

// GetFloatKeys is used to get all float keys from the save file
func (s *Save) GetFloatKeys() []string {
	var keys []string
	for key := range s.FloatData {
		keys = append(keys, key)
	}
	return keys
}

// GetStringKeys is used to get all string keys from the save file
func (s *Save) GetStringKeys() []string {
	var keys []string
	for key := range s.StringData {
		keys = append(keys, key)
	}
	return keys
}

// GetBoolKeys is used to get all bool keys from the save file
func (s *Save) GetBoolKeys() []string {
	var keys []string
	for key := range s.BoolData {
		keys = append(keys, key)
	}
	return keys
}
