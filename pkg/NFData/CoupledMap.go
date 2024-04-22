package NFData

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"sort"
	"strconv"
)

type CustomMap map[string]interface{}

func NewCustomMap(customMap CustomMap) CustomMap {
	return customMap
}

func (c CustomMap) Copy() Copyable {
	newMap := make(CustomMap)
	for k, v := range c {
		_, ok := v.(Copyable)
		if ok {
			newV := v.(Copyable).Copy()
			newMap[k] = newV
		} else {
			newMap[k] = v
		}
	}
	return newMap
}

type CoupledMap struct {
	Map   CustomMap
	types TypedMap
}

func NewCoupledMap(args CustomMap) *CoupledMap {
	return &CoupledMap{
		Map:   args,
		types: NewTypedMap(args),
	}
}

func (cm *CoupledMap) Copy(key interface{}) {
	index := key.(string)
	newIndex := index + "_copy"
	//Loop until the new key doesn't exist adding a number to the end of it incrementing it each time
	for i := 0; ; i++ {
		if _, ok := cm.Map[newIndex]; !ok {
			break
		}
		newIndex = index + "_copy" + strconv.Itoa(i)
	}
	_, ok := cm.Map[index].(Copyable)
	if ok {
		newVal := cm.Map[index].(Copyable).Copy()
		cm.Map[newIndex] = newVal
	} else {
		cm.Map[newIndex] = cm.Map[index]
		cm.types[newIndex] = cm.types[index]
	}
}

func (cm *CoupledMap) ParseValue(key interface{}, val string) (interface{}, error) {
	index, ok := key.(string)
	if !ok {
		return nil, errors.New("key is not a string")
	}
	switch cm.types[index] {
	case IntType:
		return strconv.Atoi(val)
	case FloatType:
		return strconv.ParseFloat(val, 64)
	case BooleanType:
		return strconv.ParseBool(val)
	case StringType:
		return val, nil
	default:
		return nil, errors.New("unknown type")
	}
}

func (cm *CoupledMap) SetKey(key interface{}, newKey interface{}) bool {
	//Make sure the key is a string
	index := key.(string)
	newIndex := newKey.(string)

	//Swap the values of the keys if they both exist
	if _, ok := cm.Map[index]; ok {
		if _, ok := cm.Map[newIndex]; ok {
			cm.Map[index], cm.Map[newIndex] = cm.Map[newIndex], cm.Map[index]
			cm.types[index], cm.types[newIndex] = cm.types[newIndex], cm.types[index]
		} else {
			//If the new key doesn't exist, copy the value before deleting it and then add it to the new key
			cm.Map[newIndex] = cm.Map[index]
			cm.types[newIndex] = cm.types[index]
			delete(cm.Map, index)
			delete(cm.types, index)
		}
		return true
	} else {
		return false
	}
}

func (cm *CoupledMap) Get(key interface{}) (interface{}, bool) {
	if val, ok := cm.Map[key.(string)]; ok {
		return val, true
	}
	return nil, false
}

func (cm *CoupledMap) Add() interface{} {
	//Make sure the key doesn't already exist and if it does add a number to the end of it incrementing it until it doesn't exist
	newKey := "newKey"
	for i := 0; ; i++ {
		if _, ok := cm.Map[newKey]; !ok {
			break
		}
		newKey = "newKey" + strconv.Itoa(i)
	}
	cm.Map[newKey] = ""
	cm.types[newKey] = StringType
	return newKey
}

func (cm *CoupledMap) Delete(key interface{}, waitChan chan struct{}, window fyne.Window) {
	if _, ok := cm.Map[key.(string)]; ok {
		dialog.ShowConfirm("Delete Item", "Are you sure you want to delete this item?", func(b bool) {
			if b {
				delete(cm.Map, key.(string))
				delete(cm.types, key.(string))
			}
			waitChan <- struct{}{}
		}, window)
	}
}

func (cm *CoupledMap) Keys() []string {
	keys := make([]string, 0, len(cm.Map))
	for key := range cm.Map {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (cm *CoupledMap) Object() interface{} {
	return cm.Map
}

func (cm *CoupledMap) Set(key interface{}, v interface{}) {
	cm.Map[key.(string)] = v
	cm.types[key.(string)] = GetValueType(v)
}

func (cm *CoupledMap) SetType(key interface{}, t ValueType) {
	cm.types[key.(string)] = t
	switch t {
	case PropertyType:
		cm.Map[key.(string)] = NewNFInterfaceMap()
	case MapType:
		cm.Map[key.(string)] = make(CustomMap)
	case SliceType:
		cm.Map[key.(string)] = make(CustomSlice, 0)
	case IntType:
		cm.Map[key.(string)] = 0
	case FloatType:
		cm.Map[key.(string)] = 0.0
	case BooleanType:
		cm.Map[key.(string)] = false
	default:
		cm.Map[key.(string)] = ""
	}
}
