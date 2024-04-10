package NFData

import (
	"errors"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
	"log"
	"reflect"
	"sync"
)

// NFInterfaceMap is the struct that holds all the information about the properties of a NF object.
//
// it has GADUS(Get, Add, Delete, Update, Set) functions to manipulate the data.
//
// Data should not be manipulated directly, but through the GADUS functions, as they handle the mutex locks.
type NFInterfaceMap struct {
	Data     map[string]interface{} `json:"Data"`
	mu       sync.RWMutex
	allowRef bool
}

// NewNFInterfaceMap creates a new NFInterfaceMap struct
func NewNFInterfaceMap(args ...NFKeyVal) *NFInterfaceMap {
	nfi := &NFInterfaceMap{
		Data:     make(map[string]interface{}),
		allowRef: true,
	}
	for _, arg := range args {
		err := nfi.Add(arg.Key, arg.Value)
		if err != nil {
			log.Println(err)
		}
	}
	return nfi
}

// NewRemoteNFInterfaceMap creates a new NFInterfaceMap struct that does not allow references
func NewRemoteNFInterfaceMap(args ...NFKeyVal) *NFInterfaceMap {
	nfi := &NFInterfaceMap{
		Data:     make(map[string]interface{}),
		allowRef: false,
	}
	for _, arg := range args {
		err := nfi.Add(arg.Key, arg.Value)
		if err != nil {
			log.Println(err)
		}
	}
	return nfi
}

// NewNFInterfaceFromMap creates a new NFInterfaceMap struct from a map
func NewNFInterfaceFromMap(args map[string]interface{}) *NFInterfaceMap {
	nfi := &NFInterfaceMap{
		Data: make(map[string]interface{}),
	}
	for key, value := range args {
		err := nfi.Add(key, value)
		if err != nil {
			log.Println(err)
		}
	}
	return nfi
}

// HasAllKeys compares a with b to see if all keys in b are in a
func (a *NFInterfaceMap) HasAllKeys(b *NFInterfaceMap) (bool, []string) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	missingArgs := make([]string, 0)
	for key := range b.Data {
		if _, ok := a.Data[key]; !ok {
			missingArgs = append(missingArgs, key)
		}
	}
	if len(missingArgs) > 0 {
		return false, missingArgs
	}
	return true, missingArgs
}

// ToMap returns the Data field of the NFInterfaceMap struct
func (a *NFInterfaceMap) ToMap() map[string]interface{} {
	return a.Data
}

// Export returns a map where the keys are the string representation of the type of the values in the NFInterfaceMap struct,
// and the values are slices of keys that correspond to that type.
func (a *NFInterfaceMap) Export() map[string][]string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	types := make(map[string][]string)
	for key, value := range a.Data {
		typeName := reflect.TypeOf(value).String()
		types[typeName] = append(types[typeName], key)
	}
	return types
}

// Get gets an element from the interface by reference, it will return an error if the Key does not exist or if the type does not match
func (a *NFInterfaceMap) Get(key string, ref interface{}) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if value, ok := a.Data[key]; ok {
		refVal := reflect.ValueOf(ref)
		valVal := reflect.ValueOf(value)
		if refVal.Kind() != reflect.Ptr {
			return errors.New("ref is not a pointer")
		}
		refVal = refVal.Elem()
		if !refVal.CanSet() {
			return errors.New("ref cannot be set")
		}
		if refVal.Type() != valVal.Type() {
			//If the expected type is a reference then get the value of the reference
			if valVal.Type() == reflect.TypeOf(NFReference{}) {
				if !a.allowRef {
					return NFError.NewErrInvalidArgument("reference", "references are not allowed in this context")
				}
				nfr := value.(NFReference)
				return nfr.Get(ref)
			}
			//If the expected type is this Interface then create a new one from the values' Data
			if refVal.Type() == reflect.TypeOf(&NFInterfaceMap{}) {
				//Check if the type of the value is a map that contains a Data field
				if valVal.Kind() != reflect.Map {
					return NFError.NewErrTypeMismatch(reflect.TypeOf(value).String(), refVal.Type().String())
				} else if _, ok := valVal.Interface().(map[string]interface{})["Data"]; !ok {
					return NFError.NewErrTypeMismatch(reflect.TypeOf(value).String(), refVal.Type().String())
				}
				nfi := NewNFInterfaceFromMap(value.(map[string]interface{})["Data"].(map[string]interface{}))
				refVal.Set(reflect.ValueOf(nfi))
				return nil
			}
			return NFError.NewErrTypeMismatch(reflect.TypeOf(value).String(), refVal.Type().String())
		}
		refVal.Set(reflect.ValueOf(value))
		return nil
	} else {
		return NFError.NewErrKeyNotFound(key)
	}
}

// Add adds a value to the interface by Key, errors if the Key already exists
func (a *NFInterfaceMap) Add(key string, value interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	//Check if the Key already exists
	if _, ok := a.Data[key]; ok {
		return NFError.NewErrKeyAlreadyExists(key)
	}
	a.Data[key] = value
	return nil
}

// AddMulti adds multiple values to the interface, errors if a Key already exists
func (a *NFInterfaceMap) AddMulti(args ...NFKeyVal) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	var err error
	for _, arg := range args {
		if _, ok := a.Data[arg.Key]; ok {
			errors.Join(err, NFError.NewErrKeyAlreadyExists(arg.Key))
			continue
		}
		a.Data[arg.Key] = arg.Value
	}
	return err
}

// Delete deletes a value in the interface, errors if the Key does not exist
func (a *NFInterfaceMap) Delete(key string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.Data[key]; ok {
		delete(a.Data, key)
	} else {
		return NFError.NewErrKeyNotFound(key)
	}
	return nil
}

// DeleteMulti deletes multiple values in the interface, errors if the Key does not exist
func (a *NFInterfaceMap) DeleteMulti(keys ...string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	var err error
	for _, key := range keys {
		if _, ok := a.Data[key]; !ok {
			errors.Join(err, NFError.NewErrKeyNotFound(key))
			continue
		}
		delete(a.Data, key)
	}
	return err
}

// Update updates a value in the interface, errors if the Key does not exist or if the type does not match
func (a *NFInterfaceMap) Update(key string, value interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.Data[key]; ok {
		//Check if the types match
		expectedType := reflect.TypeOf(value)
		actualType := reflect.TypeOf(a.Data[key])
		if expectedType != actualType {
			return NFError.NewErrTypeMismatch(expectedType.String(), actualType.String())
		}
		a.Data[key] = value
	} else {
		return NFError.NewErrKeyNotFound(key)
	}
	return nil
}

// UpdateMulti updates multiple values in the interface, errors if the Key does not exist or if the type does not match
func (a *NFInterfaceMap) UpdateMulti(args ...NFKeyVal) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	var err error
	for _, arg := range args {
		if _, ok := a.Data[arg.Key]; !ok {
			errors.Join(err, NFError.NewErrKeyNotFound(arg.Key))
			continue
		}
		//Check if the types match
		expectedType := reflect.TypeOf(arg.Value)
		actualType := reflect.TypeOf(a.Data[arg.Key])
		if expectedType != actualType {
			errors.Join(err, NFError.NewErrTypeMismatch(expectedType.String(), actualType.String()))
			continue
		}
		a.Data[arg.Key] = arg.Value
	}
	return err
}

// Set sets a value in the interface does not care about the type
func (a *NFInterfaceMap) Set(key string, value interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Data[key] = value
}

// SetMulti sets multiple values in the interface,
func (a *NFInterfaceMap) SetMulti(args ...NFKeyVal) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, arg := range args {
		a.Data[arg.Key] = arg.Value
	}
}
