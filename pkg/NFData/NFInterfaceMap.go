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
	Data     CustomMap `json:"Data"`
	mu       sync.RWMutex
	allowRef bool
}

// Lock locks the NFInterfaceMap struct with an optional read lock
func (a *NFInterfaceMap) Lock(read bool) {
	if read {
		a.mu.RLock()
	} else {
		a.mu.Lock()
	}
}

// Unlock unlocks the NFInterfaceMap struct with an optional read lock
func (a *NFInterfaceMap) Unlock(read bool) {
	if read {
		a.mu.RUnlock()
	} else {
		a.mu.Unlock()
	}
}

// Copy returns a new NFInterfaceMap with the same data as a Copyable
func (a *NFInterfaceMap) Copy() Copyable {
	a.mu.RLock()
	defer a.mu.RUnlock()
	newMap := make(CustomMap, len(a.Data))
	for key, value := range a.Data {
		if copyable, ok := value.(Copyable); ok {
			newMap[key] = copyable.Copy()
		} else {
			newMap[key] = value
		}
	}
	return NewNFInterfaceFromMap(newMap)
}

// NewNFInterfaceMap creates a new NFInterfaceMap struct
func NewNFInterfaceMap(args ...NFKeyVal) *NFInterfaceMap {
	nfi := &NFInterfaceMap{
		Data:     make(CustomMap),
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
		Data:     make(CustomMap),
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
func NewNFInterfaceFromMap(args CustomMap) *NFInterfaceMap {
	nfi := &NFInterfaceMap{
		Data: make(CustomMap),
	}

	if len(args) == 1 {
		if dataMap, ok := args["Data"].(map[string]interface{}); ok {
			return NewNFInterfaceFromMap(dataMap)
		} else if dataMap, ok = args["Data"].(CustomMap); ok {
			return NewNFInterfaceFromMap(dataMap)
		}
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
func (a *NFInterfaceMap) ToMap() CustomMap {
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

// UnTypedGet gets an element from the interface if it exists
func (a *NFInterfaceMap) UnTypedGet(key string) (interface{}, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if value, ok := a.Data[key]; ok {
		return value, true
	}
	return nil, false
}

// Get gets an element from the interface by reference, it will return an error if the Key does not exist or if the type does not match
func (a *NFInterfaceMap) Get(key string, ref interface{}) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if value, ok := a.Data[key]; ok {
		//Make sure the ref is a pointer
		refType := reflect.TypeOf(ref)
		if refType.Kind() != reflect.Ptr {
			return errors.New("ref must be a pointer")
		}

		//Check if the types match
		expectedType := reflect.TypeOf(value)
		actualType := reflect.TypeOf(ref).Elem()
		if expectedType != actualType {
			//Check if the expected type is an NFInterfaceMap
			if expectedType == reflect.TypeOf(&NFInterfaceMap{}) {
				//Check if the actual type is a map with only the Data field
				var customMap CustomMap
				simpleMap, ok := value.(map[string]interface{})
				if !ok {
					customMap, ok = value.(CustomMap)
					if !ok {
						return NFError.NewErrTypeMismatch(expectedType.String(), actualType.String())
					}
				}
				customMap = simpleMap
				//Create a new NFInterfaceMap from the map
				newMap := NewNFInterfaceFromMap(customMap)
				//Set the ref to the new NFInterfaceMap
				reflect.ValueOf(ref).Elem().Set(reflect.ValueOf(newMap))
			} else {
				return NFError.NewErrTypeMismatch(expectedType.String(), actualType.String())
			}

		}
		//Set the ref to the value
		reflect.ValueOf(ref).Elem().Set(reflect.ValueOf(value))
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
