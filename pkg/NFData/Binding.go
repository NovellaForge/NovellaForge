package NFData

import (
	"fyne.io/fyne/v2/data/binding"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
	"reflect"
	"sync"
)

// NFBindingMap is a struct that holds all the bindings to allow dynamic UI updates
//
// # Bindings should not be deleted as they are used to update the UI, and they should only be used where required
//
// The bindings are stored in a map with the key being the name of the binding and are split into different types
type NFBindingMap struct {
	intBindings     map[string]binding.Int
	floatBindings   map[string]binding.Float
	boolBindings    map[string]binding.Bool
	stringBindings  map[string]binding.String
	untypedBindings map[string]binding.Untyped
	mu              sync.RWMutex
}

type SupportedBindingType string

const (
	BindingUntyped SupportedBindingType = "Untyped"
	BindingInt     SupportedBindingType = "Int"
	BindingFloat   SupportedBindingType = "Float"
	BindingBool    SupportedBindingType = "Bool"
	BindingString  SupportedBindingType = "String"
)

func (nfb *NFBindingMap) GetIntBinding(key string) (binding.Int, error) {
	var b *binding.Int
	err := nfb.GetBinding(key, &b)
	if err != nil {
		return nil, err
	}
	return *b, nil
}

func (nfb *NFBindingMap) GetFloatBinding(key string) (binding.Float, error) {
	var b *binding.Float
	err := nfb.GetBinding(key, &b)
	if err != nil {
		return nil, err
	}
	return *b, nil
}

func (nfb *NFBindingMap) GetBoolBinding(key string) (binding.Bool, error) {
	var b *binding.Bool
	err := nfb.GetBinding(key, &b)
	if err != nil {
		return nil, err
	}
	return *b, nil
}

func (nfb *NFBindingMap) GetStringBinding(key string) (binding.String, error) {
	var b *binding.String
	err := nfb.GetBinding(key, &b)
	if err != nil {
		return nil, err
	}
	return *b, nil
}

func (nfb *NFBindingMap) GetUntypedBinding(key string) (binding.Untyped, error) {
	var b *binding.Untyped
	err := nfb.GetBinding(key, &b)
	if err != nil {
		return nil, err
	}
	return *b, nil
}

func (nfb *NFBindingMap) GetBinding(key string, bindingRef interface{}) error {
	nfb.mu.RLock()
	defer nfb.mu.RUnlock()

	switch ref := bindingRef.(type) {
	case **binding.Int:
		if b, ok := nfb.intBindings[key]; ok {
			*ref = &b
			return nil
		}
	case **binding.Float:
		if b, ok := nfb.floatBindings[key]; ok {
			*ref = &b
			return nil
		}
	case **binding.Bool:
		if b, ok := nfb.boolBindings[key]; ok {
			*ref = &b
			return nil
		}
	case **binding.String:
		if b, ok := nfb.stringBindings[key]; ok {
			*ref = &b
			return nil
		}
	case **binding.Untyped:
		if b, ok := nfb.untypedBindings[key]; ok {
			*ref = &b
			return nil
		}
	default:
		return NFError.NewErrTypeMismatch(reflect.TypeOf(bindingRef).String(), "ref is not a pointer to a binding type")
	}

	return NFError.NewErrKeyNotFound(key)
}

// NewNFBindingMap creates a new NFBindingMap struct
func NewNFBindingMap(args ...KeyVal) *NFBindingMap {
	nfb := &NFBindingMap{
		intBindings:     make(map[string]binding.Int),
		floatBindings:   make(map[string]binding.Float),
		boolBindings:    make(map[string]binding.Bool),
		stringBindings:  make(map[string]binding.String),
		untypedBindings: make(map[string]binding.Untyped),
	}
	for _, arg := range args {
		err := nfb.CreateBinding(arg.Key, arg.Value)
		if err != nil {
			return nil
		}
	}
	return nfb
}

// CreateBinding creates a new binding based on the type of the value this DOES NOT update the value
func (nfb *NFBindingMap) CreateBinding(key string, value interface{}) error {
	nfb.mu.Lock()
	defer nfb.mu.Unlock()

	//Check if the ref is a pointer
	switch value.(type) {
	case int:
		nfb.intBindings[key] = binding.NewInt()
		err := nfb.intBindings[key].Set(value.(int))
		if err != nil {
			return err
		}
	case *int:
		nfb.intBindings[key] = binding.BindInt(value.(*int))
	case float64:
		nfb.floatBindings[key] = binding.NewFloat()
		err := nfb.floatBindings[key].Set(value.(float64))
		if err != nil {
			return err
		}
	case *float64:
		nfb.floatBindings[key] = binding.BindFloat(value.(*float64))
	case bool:
		nfb.boolBindings[key] = binding.NewBool()
		err := nfb.boolBindings[key].Set(value.(bool))
		if err != nil {
			return err
		}
	case *bool:
		nfb.boolBindings[key] = binding.BindBool(value.(*bool))
	case string:
		nfb.stringBindings[key] = binding.NewString()
		err := nfb.stringBindings[key].Set(value.(string))
		if err != nil {
			return err
		}
	case *string:
		nfb.stringBindings[key] = binding.BindString(value.(*string))
	default:
		//Check if the ref is a pointer
		if reflect.TypeOf(value).Kind() == reflect.Ptr {
			nfb.untypedBindings[key] = binding.BindUntyped(value)
		} else {
			nfb.untypedBindings[key] = binding.NewUntyped()
			err := nfb.untypedBindings[key].Set(value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
