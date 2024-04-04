package NFData

import (
	"go.novellaforge.dev/novellaforge/pkg/NFError"
)

type Type string

const (
	NFRefScene  Type = "Scene"
	NFRefGlobal Type = "Global"
)

type NFReference struct {
	Location    Type                 `json:"Location"`
	Key         string               `json:"Key"`
	IsBinding   bool                 `json:"IsBinding"`
	BindingType SupportedBindingType `json:"BindingType"`
}

// NewRef creates a new NFReference struct
func NewRef(t Type, key string) NFReference {
	return NFReference{
		Location: t,
		Key:      key,
	}
}

// NewRefWithBinding creates a new NFReference struct with a binding
func NewRefWithBinding(t Type, key string, bindingType SupportedBindingType) NFReference {
	return NFReference{
		Location:    t,
		Key:         key,
		IsBinding:   true,
		BindingType: bindingType,
	}
}

// Get gets the value of the reference
func (r *NFReference) Get(ref interface{}) error {
	switch r.Location {
	case NFRefScene:
		return ActiveSceneData.Variables.Get(r.Key, ref)
	case NFRefGlobal:
		return GlobalVars.Get(r.Key, ref)
	default:
		return NFError.NewErrInvalidArgument("reference", "type not found")
	}
}

func (r *NFReference) GetBinding() (interface{}, error) {
	var bindings *NFBindingMap
	switch r.Location {
	case NFRefScene:
		bindings = ActiveSceneData.Bindings
	case NFRefGlobal:
		bindings = GlobalBindings
	default:
		return nil, NFError.NewErrInvalidArgument("reference", "type not found")
	}
	switch r.BindingType {
	case BindingUntyped:
		return bindings.GetUntypedBinding(r.Key)
	case BindingInt:
		return bindings.GetIntBinding(r.Key)
	case BindingFloat:
		return bindings.GetFloatBinding(r.Key)
	case BindingBool:
		return bindings.GetBoolBinding(r.Key)
	case BindingString:
		return bindings.GetStringBinding(r.Key)
	default:
		return nil, NFError.NewErrInvalidArgument("reference", "binding type not found")
	}
}

// Add adds the reference to the Location
func (r *NFReference) Add(ref interface{}) error {
	switch r.Location {
	case NFRefScene:
		return ActiveSceneData.Variables.Add(r.Key, ref)
	case NFRefGlobal:
		return GlobalVars.Add(r.Key, ref)
	default:
		return NFError.NewErrInvalidArgument("reference", "type not found")
	}
}

// CreateBinding creates a new binding for the reference
func (r *NFReference) CreateBinding(ref interface{}) error {
	switch r.Location {
	case NFRefScene:
		return ActiveSceneData.Bindings.CreateBinding(r.Key, ref)
	case NFRefGlobal:
		return GlobalBindings.CreateBinding(r.Key, ref)
	default:
		return NFError.NewErrInvalidArgument("reference", "type not found")
	}
}

// Delete deletes the reference from the Location
func (r *NFReference) Delete() error {
	switch r.Location {
	case NFRefScene:
		return ActiveSceneData.Variables.Delete(r.Key)
	case NFRefGlobal:
		return GlobalVars.Delete(r.Key)
	default:
		return NFError.NewErrInvalidArgument("reference", "type not found")
	}
}

// Set sets the value of the reference
func (r *NFReference) Set(ref interface{}) error {
	switch r.Location {
	case NFRefScene:
		ActiveSceneData.Variables.Set(r.Key, ref)
		return nil
	case NFRefGlobal:
		GlobalVars.Set(r.Key, ref)
		return nil
	default:
		return NFError.NewErrInvalidArgument("reference", "type not found")
	}
}
