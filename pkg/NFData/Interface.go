package NFData

import (
	"errors"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
	"log"
	"reflect"
	"strings"
)

// NFInterface is the struct that holds all the information about the properties of a NF object
//
// it has GADUS(Get, Add, Delete, Update, Set) functions for all the types
type NFInterface map[string]interface{}

type KeyVal struct {
	Key   string
	Value interface{}
}

// NewKeyVal creates a new KeyVal struct
func NewKeyVal(key string, value interface{}) KeyVal {
	return KeyVal{
		Key:   key,
		Value: value,
	}
}

// NewNFInterface creates a new NFInterface struct
func NewNFInterface(args ...KeyVal) NFInterface {
	nfi := make(NFInterface)
	for _, arg := range args {
		err := nfi.Add(arg.Key, arg.Value)
		if err != nil {
			log.Println(err)
		}
	}
	return nfi
}

// HasAllKeys compares a with b to see if all keys in b are in a
func (a *NFInterface) HasAllKeys(b NFInterface) (bool, []string) {
	missingArgs := make([]string, 0)
	for key := range b {
		if _, ok := (*a)[key]; !ok {
			missingArgs = append(missingArgs, key)
		}
	}
	if len(missingArgs) > 0 {
		return false, missingArgs
	}
	return true, missingArgs
}

// Export returns a map where the keys are the string representation of the type of the values in the NFInterface struct,
// and the values are slices of keys that correspond to that type.
func (a *NFInterface) Export() map[string][]string {
	types := make(map[string][]string)
	for key, value := range *a {
		typeName := reflect.TypeOf(value).String()
		types[typeName] = append(types[typeName], key)
	}
	return types
}

// Set sets a value in the interface, keys cannot contain a period as that is used for referencing other interfaces
func (a *NFInterface) Set(key string, value interface{}) error {
	if strings.Contains(key, ".") {
		return errors.New("key cannot contain a '.'")
	}
	(*a)[key] = value
	return nil
}

// SetMulti sets multiple values in the interface, keys cannot contain a period as that is used for referencing other interfaces
func (a *NFInterface) SetMulti(args ...KeyVal) {
	for _, arg := range args {
		err := a.Set(arg.Key, arg.Value)
		if err != nil {
			log.Println(err)
		}
	}
}

// Add adds a value to the interface by key, errors if the key already exists or if the key contains a period as that is used for referencing other interfaces
func (a *NFInterface) Add(key string, value interface{}) error {
	if strings.Contains(key, ".") {
		return errors.New("key cannot contain a '.'")
	}
	//Check if the key already exists
	if _, ok := (*a)[key]; ok {
		log.Println("Key already exists")
		return NFError.ErrKeyAlreadyExists
	}
	(*a)[key] = value
	return nil
}

// Update updates a value in the interface, errors if the key does not exist or if the type does not match
func (a *NFInterface) Update(key string, value interface{}) error {
	if _, ok := (*a)[key]; ok {
		//Check if the types match
		expectedType := reflect.TypeOf(value)
		actualType := reflect.TypeOf((*a)[key])
		if expectedType != actualType {
			log.Println("value is type: ", expectedType, "but value at key is type: ", actualType)
			return NFError.ErrTypeMismatch
		}
		(*a)[key] = value
	} else {
		return NFError.ErrKeyNotFound
	}
	return nil
}

// Delete deletes a value in the interface, errors if the key does not exist
func (a *NFInterface) Delete(key string) error {
	if _, ok := (*a)[key]; ok {
		delete(*a, key)
	} else {
		return NFError.ErrKeyNotFound
	}
	return nil
}

// Get gets an element from the interface by reference, it will return an error if the key does not exist or if the type does not match
func (a *NFInterface) Get(key string, ref interface{}) error {
	//If the key starts with global. or scene. then it is a reference to a different interface, and it should pull from that interface
	//Split the key by the first period
	keySplit := strings.SplitN(key, ".", 2)
	if len(keySplit) > 1 {
		//Check if the key is a reference to a different interface
		if keySplit[0] == "global" {
			//Get the value from the global interface
			return GlobalVars.Get(keySplit[1], ref)
		} else if keySplit[0] == "scene" {
			//Get the value from the scene interface
			return ActiveSceneData.Variables.Get(keySplit[1], ref)
		}
	}
	if value, ok := (*a)[key]; ok {
		valRef := reflect.ValueOf(ref)
		if valRef.Kind() != reflect.Ptr {
			return errors.New("ref is not a pointer")
		}
		valRef = valRef.Elem()
		if !valRef.CanSet() {
			return errors.New("ref cannot be set")
		}
		if valRef.Type() != reflect.TypeOf(value) {
			log.Println("Ref is type: ", reflect.TypeOf(value), "but type found for key is: ", valRef.Type())
			return NFError.ErrTypeMismatch
		}
		valRef.Set(reflect.ValueOf(value))
		return nil
	} else {
		return NFError.ErrKeyNotFound
	}
}
