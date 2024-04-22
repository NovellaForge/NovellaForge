package NFData

import (
	"fyne.io/fyne/v2"
)

type CoupledObject interface {
	Keys() []string
	Add() interface{}
	Object() interface{}
	Get(key interface{}) (interface{}, bool)
	Set(key interface{}, value interface{})
	SetType(key interface{}, t ValueType)
	SetKey(key interface{}, newKey interface{}) bool
	ParseValue(key interface{}, val string) (interface{}, error)
	Delete(key interface{}, waitChan chan struct{}, window fyne.Window)
	Copy(key interface{})
}

type TypedMap map[string]ValueType

func NewTypedMap(data CustomMap) TypedMap {
	types := make(TypedMap)
	for key, value := range data {
		types[key] = GetValueType(value)
	}
	return types
}

type TypedSlice []ValueType

func NewTypedSlice(slice CustomSlice) TypedSlice {
	types := make(TypedSlice, len(slice))
	for i, value := range slice {
		types[i] = GetValueType(value)
	}
	return types
}
