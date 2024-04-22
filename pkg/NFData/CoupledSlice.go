package NFData

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"sort"
	"strconv"
)

type CustomSlice []interface{}

func NewCustomSlice(customSlice CustomSlice) CustomSlice {
	return customSlice
}

func (c CustomSlice) Copy() Copyable {
	newSlice := make(CustomSlice, len(c))
	for i, v := range c {
		_, ok := v.(Copyable)
		if ok {
			newV := v.(Copyable).Copy()
			newSlice[i] = newV
		} else {
			newSlice[i] = v
		}
	}
	return newSlice
}

type CoupledSlice struct {
	Slice CustomSlice
	types TypedSlice
}

func NewCoupledSlice(args CustomSlice) *CoupledSlice {
	return &CoupledSlice{
		Slice: args,
		types: NewTypedSlice(args),
	}
}

func (cs *CoupledSlice) Copy(key interface{}) {
	index := -1
	if _, ok := key.(int); !ok {
		index, _ = strconv.Atoi(key.(string))
	} else {
		index = key.(int)
	}
	if index < 0 || index >= len(cs.Slice) {
		return
	}

	//Check if the value at the index is a copyable type
	_, ok := cs.Slice[index].(Copyable)
	if ok {
		newVal := cs.Slice[index].(Copyable).Copy()
		cs.Slice = append(cs.Slice, newVal)
		cs.types = append(cs.types, cs.types[index])
	} else {
		cs.Slice = append(cs.Slice, cs.Slice[index])
		cs.types = append(cs.types, cs.types[index])
	}
}

func (cs *CoupledSlice) ParseValue(key interface{}, val string) (interface{}, error) {
	//Check if the 'key' is a string or an int
	index := -1
	if _, ok := key.(int); !ok {
		var err error
		index, err = strconv.Atoi(key.(string))
		if err != nil {
			return nil, err
		}
	} else {
		index = key.(int)
	}
	//Check if the index is within the bounds of the slice
	if index < 0 || index >= len(cs.Slice) {
		return nil, errors.New("index out of bounds")
	}
	switch cs.types[index] {
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

func (cs *CoupledSlice) SetKey(key interface{}, newKey interface{}) bool {
	//Check if the 'key' is a string or an int
	index := -1
	if _, ok := key.(int); !ok {
		index, _ = strconv.Atoi(key.(string))
	} else {
		index = key.(int)
	}
	//Check if the 'newKey' is a string or an int
	newIndex := -1
	if _, ok := newKey.(int); !ok {
		newIndex, _ = strconv.Atoi(newKey.(string))
	} else {
		newIndex = newKey.(int)
	}

	//If they are the same index, return false
	if index == newIndex {
		return false
	}

	//If the index already exists, move the value to the new index swapping the values
	//If it doesn't exist and is greater than the length of the slice, append the value to the end of the slice
	//If it doesn't exist and is less than 0, insert the value at the beginning of the slice
	if index < len(cs.Slice) && index >= 0 {
		if newIndex < len(cs.Slice) && newIndex >= 0 {
			cs.Slice[index], cs.Slice[newIndex] = cs.Slice[newIndex], cs.Slice[index]
			cs.types[index], cs.types[newIndex] = cs.types[newIndex], cs.types[index]
		} else if newIndex >= len(cs.Slice) {
			//Copy the value before deleting it and then append it to the end of the slice
			value := cs.Slice[index]
			valueType := cs.types[index]
			cs.Slice = append(cs.Slice[:index], cs.Slice[index+1:]...)
			cs.types = append(cs.types[:index], cs.types[index+1:]...)
			cs.Slice = append(cs.Slice, value)
			cs.types = append(cs.types, valueType)
		} else if newIndex < 0 {
			//Copy the value before deleting it and then append it to the beginning of the slice
			value := cs.Slice[index]
			valueType := cs.types[index]
			cs.Slice = append(cs.Slice[:index], cs.Slice[index+1:]...)
			cs.types = append(cs.types[:index], cs.types[index+1:]...)
			cs.Slice = append(CustomSlice{value}, cs.Slice...)
			cs.types = append(TypedSlice{valueType}, cs.types...)
		}
		return true
	} else {
		return false
	}
}

func (cs *CoupledSlice) Add() interface{} {
	cs.Slice = append(cs.Slice, "")
	cs.types = append(cs.types, StringType)
	lastIndex := len(cs.Slice) - 1
	return lastIndex
}

func (cs *CoupledSlice) Get(key interface{}) (interface{}, bool) {
	//Check if the 'key' is a string or an int
	index := -1
	if _, ok := key.(int); !ok {
		index, _ = strconv.Atoi(key.(string))
	} else {
		index = key.(int)
	}
	if index < len(cs.Slice) && index >= 0 {
		return cs.Slice[index], true
	}
	return nil, false
}

func (cs *CoupledSlice) Delete(key interface{}, waitChan chan struct{}, window fyne.Window) {
	//Check if the 'key' is a string or an int
	index := -1
	if _, ok := key.(int); !ok {
		index, _ = strconv.Atoi(key.(string))
	} else {
		index = key.(int)
	}
	if index < len(cs.Slice) && index >= 0 {
		dialog.ShowConfirm("Delete Item", "Are you sure you want to delete this item?", func(b bool) {
			if b {
				cs.Slice = append(cs.Slice[:index], cs.Slice[index+1:]...)
				cs.types = append(cs.types[:index], cs.types[index+1:]...)
			}
			waitChan <- struct{}{}
		}, window)
	}
}

func (cs *CoupledSlice) Keys() []string {
	keys := make([]string, len(cs.Slice))
	for i := range cs.Slice {
		keys[i] = strconv.Itoa(i)
	}
	sort.Strings(keys)
	return keys
}

func (cs *CoupledSlice) Object() interface{} {
	return cs.Slice
}

func (cs *CoupledSlice) Set(i interface{}, v interface{}) {
	//Check if the 'i' is a string or an int
	index := -1
	if _, ok := i.(int); !ok {
		index, _ = strconv.Atoi(i.(string))
	} else {
		index = i.(int)
	}
	if index < 0 {
		return
	}
	//If it is greater than the length of the slice, append the value to the end of the slice
	if index >= len(cs.Slice) {
		cs.Slice = append(cs.Slice, v)
		cs.types = append(cs.types, GetValueType(v))
		return
	}

	cs.Slice[index] = v
	cs.types[index] = GetValueType(v)
}

func (cs *CoupledSlice) SetType(i interface{}, t ValueType) {
	index := -1
	if _, ok := i.(int); !ok {
		index, _ = strconv.Atoi(i.(string))
	} else {
		index = i.(int)
	}
	if index < 0 || index >= len(cs.types) {
		return
	}
	cs.types[index] = t
	switch t {
	case PropertyType:
		cs.Slice[index] = NewNFInterfaceMap()
	case MapType:
		cs.Slice[index] = make(CustomMap)
	case SliceType:
		cs.Slice[index] = make(CustomSlice, 0)
	case IntType:
		cs.Slice[index] = 0
	case FloatType:
		cs.Slice[index] = 0.0
	case BooleanType:
		cs.Slice[index] = false
	default:
		cs.Slice[index] = ""
	}
}
