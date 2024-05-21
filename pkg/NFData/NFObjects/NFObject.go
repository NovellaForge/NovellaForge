package NFObjects

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"github.com/google/uuid"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"os"
	"path/filepath"
)

type NFObject interface {
	GetArgs() *NFData.NFInterfaceMap        // GetArgs All NFObjects should have an NFData.NFInterfaceMap for their arguments
	SetArgs(newArgs *NFData.NFInterfaceMap) // SetArgs All NFObjects should have an NFData.NFInterfaceMap for their arguments and this function allows resetting the arguments
	CheckArgs() error                       // CheckArgs Validates the object's arguments
	GetID() uuid.UUID                       // GetID All NFObjects should have a UUID(Universally Unique Identifier), this function returns the ID
	FetchIDs() []uuid.UUID                  // FetchIDs will fetch all the IDs of the object and call FetchIDs on all its children and functions
	GetName() string                        // GetName All NFObjects should have a name for customizing the object display in editors
	SetName(newName string)                 // SetName Allows changing the name of the object
	GetType() string                        // GetType All NFObjects should specify a type that is used to parse the object
	SetType(newType string)                 // SetType Allows changing the type of object
	MakeId()                                // MakeId Generates a new UUID for the object and calls MakeId on all its children and functions
}

//TODO fetch functions needs to be reconsidered so that there will be a way to handle the gaps of parents without functions

type NFRendered interface {
	NFObject                                                                                                                  // All NFRendered objects should implement the NFObject interface
	AddChild(newChild NFRendered)                                                                                             // AddChild Adds a child to the object
	AddFunction(newFunction NFObject)                                                                                         // AddFunction Adds an action based function to the object
	DeleteByID(ID uuid.UUID, search bool) error                                                                               // DeleteByID Removes an object based on ID, if search is true it will call DeleteByID with the same arguments on its children and functions
	DeleteChild(childID uuid.UUID, search bool) error                                                                         // DeleteChild Removes a child from the object, if search is true it will call DeleteChild with the same arguments on its children
	DeleteFunction(functionID uuid.UUID, search bool) error                                                                   // DeleteFunction Removes a function from the object based on ID, if search is true it will call DeleteFunction with the same arguments on its children
	GetFunctions() []NFObject                                                                                                 // GetFunctions Returns all functions associated with the object
	GetChildByID(childID uuid.UUID) NFObject                                                                                  // GetChildByID Returns the first found child of the object based on the ID if none are found it calls GetChildByID on its children and returns the result
	GetFunctionByID(functionID uuid.UUID) NFObject                                                                            // GetFunctionByID returns the first found function of the object based on the ID if none are found it calls GetFunctionByID on its children and returns the result
	GetByID(ID uuid.UUID) NFObject                                                                                            // GetByID returns the object based on the ID if none are found it calls GetByID on its children and returns the result
	FetchChildrenAndFunctions(childrenAndFunctions map[uuid.UUID][]NFObject) int                                              // FetchChildrenAndFunctions Collects all the children and functions associated with the object and its children. Maps are reference types, so the passed in map will be modified and not returned, the return value is the number of children and functions found
	FetchChildren(children map[uuid.UUID][]NFObject) int                                                                      // FetchChildren Collects all the children associated with the object and its children. Maps are reference types, so the passed in map will be modified and not returned, the return value is the number of children found
	FetchFunctions(functions map[uuid.UUID][]NFObject) int                                                                    // FetchFunctions Collects all the functions associated with the object and its children. Maps are reference types, so the passed in map will be modified and not returned, the return value is the number of functions found
	RunAllActions(action string, window fyne.Window, newValues *NFData.NFInterfaceMap) (int, []*NFData.NFInterfaceMap, error) // RunAllActions Runs all the functions associated with the action on the object
	RunAction(action string, window fyne.Window, newValues *NFData.NFInterfaceMap) (*NFData.NFInterfaceMap, error)            // RunAction Runs the first found function associated with the action on the object
}

type NFRoot interface {
	NFRendered                                          // All NFRoot objects should implement the NFRendered interface
	Validate() error                                    // Validate Should be called on the highest level object in the hierarchy to validate the object and its children
	FetchAll() (map[uuid.UUID][]NFObject, int)          // FetchAll Should be called on the highest level object in the hierarchy to fetch all the objects and their children and functions all the way down the hierarchy nesting them according to their parent
	FetchAllChildren() (map[uuid.UUID][]NFObject, int)  // FetchAllChildren Should be called on the highest level object in the hierarchy to fetch all the children of the object all the way down the hierarchy nesting them according to their parent
	FetchAllFunctions() (map[uuid.UUID][]NFObject, int) // FetchAllFunctions Should be called on the highest level object in the hierarchy to fetch all the functions of the object all the way down the hierarchy nesting them according to their parent
}

type AssetProperties struct {
	Type             string              `json:"Type"`
	SupportedActions []string            `json:"SupportedActions"`
	RequiredArgs     map[string][]string `json:"RequiredArgs"`
	OptionalArgs     map[string][]string `json:"OptionalArgs"`
}

func (a *AssetProperties) Load(path string) error {
	//Check if the path is a file
	if filepath.Ext(path) == "" {
		return errors.New("invalid file type")
	}
	//Read the file
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	//Unmarshal the json file
	err = json.Unmarshal(file, a)
	if err != nil {
		return err
	}
	return nil
}
