package NFEditor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"reflect"
	"sort"
	"strconv"
)

type CoupledObject interface {
	Get(key interface{}) (interface{}, bool)
	Set(key interface{}, value interface{})
	SetType(key interface{}, t ValueType)
	SetKey(key interface{}, newKey interface{}) bool
	Add() interface{}
	Delete(key interface{}, waitChan chan struct{}, window fyne.Window)
	Keys() []string
	Object() interface{}
}

type ValueType string

const (
	StringType   ValueType = "String"
	IntType      ValueType = "Int"
	FloatType    ValueType = "Float"
	BooleanType  ValueType = "Boolean"
	MapType      ValueType = "Object"
	SliceType    ValueType = "Slice"
	PropertyType ValueType = "Args"
	UnknownType  ValueType = "Unknown"
)

func (t ValueType) String() string {
	return string(t)
}

func GetTypes() []ValueType {
	return []ValueType{StringType, IntType, FloatType, BooleanType, MapType, SliceType, PropertyType, UnknownType}
}

func GetType(value string) ValueType {
	for _, t := range GetTypes() {
		if t.String() == value {
			return t
		}
	}
	return UnknownType
}

func GetTypesString() []string {
	types := GetTypes()
	strTypes := make([]string, len(types))
	for i, t := range types {
		strTypes[i] = t.String()
	}
	return strTypes
}

func GetValueTypeAndLabel(value interface{}) (ValueType, string) {
	valType := reflect.TypeOf(value)
	valKind := valType.Kind()
	//Check if the value is an nfInterfaceMap

	if valType == reflect.TypeOf(&NFData.NFInterfaceMap{}) {
		return PropertyType, "Args..."
	}

	if valKind == reflect.Ptr {
		valKind = valType.Elem().Kind()
	}
	switch valKind {
	case reflect.Slice, reflect.Array:
		return SliceType, "Array..."
	case reflect.Map:
		// Check if the map contains the "Data" key
		if dataMap, ok := value.(map[string]interface{}); ok {
			// Check if the map only contains the "Data" key
			if len(dataMap) == 1 {
				if _, ok := dataMap["Data"].(map[string]interface{}); ok {
					// If it does, return PropertyType
					return PropertyType, "Args..."
				}
			}
		}
		return MapType, "Object..."
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return IntType, strconv.Itoa(value.(int))
	case reflect.Float32, reflect.Float64:
		return FloatType, strconv.FormatFloat(value.(float64), 'f', -1, 64)
	case reflect.Bool:
		return BooleanType, strconv.FormatBool(value.(bool))
	case reflect.String:
		return StringType, value.(string)
	default:
		return UnknownType, "Unknown"
	}
}

func GetValueType(value interface{}) ValueType {
	if value == nil {
		return UnknownType
	}
	valType, _ := GetValueTypeAndLabel(value)
	return valType
}

func GetValueString(value interface{}) string {
	if value == nil {
		return "nil"
	}
	_, label := GetValueTypeAndLabel(value)
	return label
}

type TypedMap map[string]ValueType
type CustomMap map[string]interface{}

func NewTypedMap(data CustomMap) TypedMap {
	types := make(TypedMap)
	for key, value := range data {
		types[key] = GetValueType(value)
	}
	return types
}

type TypedSlice []ValueType
type CustomSlice []interface{}

func NewTypedSlice(slice CustomSlice) TypedSlice {
	types := make(TypedSlice, len(slice))
	for i, value := range slice {
		types[i] = GetValueType(value)
	}
	return types
}

type CoupledSlice struct {
	Slice CustomSlice
	types TypedSlice
}

func (d *CoupledSlice) SetKey(key interface{}, newKey interface{}) bool {
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
	if index < len(d.Slice) && index >= 0 {
		if newIndex < len(d.Slice) && newIndex >= 0 {
			d.Slice[index], d.Slice[newIndex] = d.Slice[newIndex], d.Slice[index]
			d.types[index], d.types[newIndex] = d.types[newIndex], d.types[index]
		} else if newIndex >= len(d.Slice) {
			//Copy the value before deleting it and then append it to the end of the slice
			value := d.Slice[index]
			valueType := d.types[index]
			d.Slice = append(d.Slice[:index], d.Slice[index+1:]...)
			d.types = append(d.types[:index], d.types[index+1:]...)
			d.Slice = append(d.Slice, value)
			d.types = append(d.types, valueType)
		} else if newIndex < 0 {
			//Copy the value before deleting it and then append it to the beginning of the slice
			value := d.Slice[index]
			valueType := d.types[index]
			d.Slice = append(d.Slice[:index], d.Slice[index+1:]...)
			d.types = append(d.types[:index], d.types[index+1:]...)
			d.Slice = append(CustomSlice{value}, d.Slice...)
			d.types = append(TypedSlice{valueType}, d.types...)
		}
		return true
	} else {
		return false
	}
}

func (d *CoupledSlice) Add() interface{} {
	d.Slice = append(d.Slice, "")
	d.types = append(d.types, UnknownType)
	lastIndex := len(d.Slice) - 1
	return lastIndex
}

func (d *CoupledSlice) Get(key interface{}) (interface{}, bool) {
	//Check if the 'key' is a string or an int
	index := -1
	if _, ok := key.(int); !ok {
		index, _ = strconv.Atoi(key.(string))
	} else {
		index = key.(int)
	}
	if index < len(d.Slice) && index >= 0 {
		return d.Slice[index], true
	}
	return nil, false
}

func (d *CoupledSlice) Delete(key interface{}, waitChan chan struct{}, window fyne.Window) {
	//Check if the 'key' is a string or an int
	index := -1
	if _, ok := key.(int); !ok {
		index, _ = strconv.Atoi(key.(string))
	} else {
		index = key.(int)
	}
	if index < len(d.Slice) && index >= 0 {
		dialog.ShowConfirm("Delete Item", "Are you sure you want to delete this item?", func(b bool) {
			if b {
				d.Slice = append(d.Slice[:index], d.Slice[index+1:]...)
				d.types = append(d.types[:index], d.types[index+1:]...)
			}
			waitChan <- struct{}{}
		}, window)
	}
}

func (d *CoupledSlice) Keys() []string {
	keys := make([]string, len(d.Slice))
	for i := range d.Slice {
		keys[i] = strconv.Itoa(i)
	}
	sort.Strings(keys)
	return keys
}

func (d *CoupledSlice) Object() interface{} {
	return d.Slice
}

func NewCoupledSlice(args CustomSlice) *CoupledSlice {
	return &CoupledSlice{
		Slice: args,
		types: NewTypedSlice(args),
	}
}

func (d *CoupledSlice) Set(i interface{}, v interface{}) {
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
	if index >= len(d.Slice) {
		d.Slice = append(d.Slice, v)
		d.types = append(d.types, GetValueType(v))
		return
	}

	d.Slice[index] = v
	d.types[index] = GetValueType(v)
}

func (d *CoupledSlice) SetType(i interface{}, t ValueType) {
	index := -1
	if _, ok := i.(int); !ok {
		index, _ = strconv.Atoi(i.(string))
	} else {
		index = i.(int)
	}
	if index < 0 || index >= len(d.types) {
		return
	}
	d.types[index] = t
	switch t {
	case PropertyType:
		if _, ok := d.Slice[index].(*NFData.NFInterfaceMap); !ok {
			d.Slice[index] = NFData.NewNFInterfaceMap()
		}
	case MapType:
		if _, ok := d.Slice[index].(map[string]interface{}); !ok {
			d.Slice[index] = make(map[string]interface{})
		}
	case SliceType:
		if _, ok := d.Slice[index].([]interface{}); !ok {
			d.Slice[index] = make([]interface{}, 0)
		}
	default:
		//If it cannot be converted to a string, set it to ""
		if _, ok := d.Slice[index].(string); !ok {
			d.Slice[index] = ""
		}
	}
}

type CoupledMap struct {
	Map   CustomMap
	types TypedMap
}

func (d *CoupledMap) SetKey(key interface{}, newKey interface{}) bool {
	//Make sure the key is a string
	index := key.(string)
	newIndex := newKey.(string)

	//Swap the values of the keys if they both exist
	if _, ok := d.Map[index]; ok {
		if _, ok := d.Map[newIndex]; ok {
			d.Map[index], d.Map[newIndex] = d.Map[newIndex], d.Map[index]
			d.types[index], d.types[newIndex] = d.types[newIndex], d.types[index]
		} else {
			//If the new key doesn't exist, copy the value before deleting it and then add it to the new key
			d.Map[newIndex] = d.Map[index]
			d.types[newIndex] = d.types[index]
			delete(d.Map, index)
			delete(d.types, index)
		}
		return true
	} else {
		return false
	}
}

func (d *CoupledMap) Get(key interface{}) (interface{}, bool) {
	if val, ok := d.Map[key.(string)]; ok {
		return val, true
	}
	return nil, false
}

func (d *CoupledMap) Add() interface{} {
	//Make sure the key doesn't already exist and if it does add a number to the end of it incrementing it until it doesn't exist
	newKey := "newKey"
	for i := 0; ; i++ {
		if _, ok := d.Map[newKey]; !ok {
			break
		}
		newKey = "newKey" + strconv.Itoa(i)
	}
	d.Map[newKey] = ""
	d.types[newKey] = UnknownType
	return newKey
}

func (d *CoupledMap) Delete(key interface{}, waitChan chan struct{}, window fyne.Window) {
	if _, ok := d.Map[key.(string)]; ok {
		dialog.ShowConfirm("Delete Item", "Are you sure you want to delete this item?", func(b bool) {
			if b {
				delete(d.Map, key.(string))
				delete(d.types, key.(string))
			}
			waitChan <- struct{}{}
		}, window)
	}
}

func (d *CoupledMap) Keys() []string {
	keys := make([]string, 0, len(d.Map))
	for key := range d.Map {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (d *CoupledMap) Object() interface{} {
	return d.Map
}

func NewCoupledMap(args CustomMap) *CoupledMap {
	return &CoupledMap{
		Map:   args,
		types: NewTypedMap(args),
	}
}

func (d *CoupledMap) Set(key interface{}, v interface{}) {
	d.Map[key.(string)] = v
	d.types[key.(string)] = GetValueType(v)
}

func (d *CoupledMap) SetType(key interface{}, t ValueType) {
	d.types[key.(string)] = t
	switch t {
	case PropertyType:
		if _, ok := d.Map[key.(string)].(*NFData.NFInterfaceMap); !ok {
			d.Map[key.(string)] = NFData.NewNFInterfaceMap()
		}
	case MapType:
		if _, ok := d.Map[key.(string)].(map[string]interface{}); !ok {
			d.Map[key.(string)] = make(map[string]interface{})
		}
	case SliceType:
		if _, ok := d.Map[key.(string)].([]interface{}); !ok {
			d.Map[key.(string)] = make([]interface{}, 0)
		}
	default:
		//If it cannot be converted to a string, set it to ""
		if _, ok := d.Map[key.(string)].(string); !ok {
			d.Map[key.(string)] = ""
		}
	}
}

type CoupledInterfaceMap struct {
	Args  *NFData.NFInterfaceMap
	types TypedMap
}

func (d *CoupledInterfaceMap) SetKey(key interface{}, newKey interface{}) bool {
	//Check if the newKey already exists if it does not, just simply copy the current value to the new key before deleting it
	if _, ok := d.Args.Data[newKey.(string)]; !ok {
		d.Args.Set(newKey.(string), d.Args.Data[key.(string)])
		d.types[newKey.(string)] = d.types[key.(string)]
		_ = d.Args.Delete(key.(string))
		delete(d.types, key.(string))
		return true
	} else {
		return false
	}
}

func (d *CoupledInterfaceMap) Get(key interface{}) (interface{}, bool) {
	if val, ok := d.Args.UnTypedGet(key.(string)); ok {
		return val, true
	}
	return nil, false
}

func (d *CoupledInterfaceMap) Add() interface{} {
	newKey := "newKey"
	for i := 0; ; i++ {
		if _, ok := d.Args.Data[newKey]; !ok {
			break
		}
		newKey = "newKey" + strconv.Itoa(i)
	}
	err := d.Args.Add(newKey, "")
	if err != nil {
		panic("Could not add new key to the interface map")
	}
	d.types[newKey] = UnknownType
	return newKey
}

func (d *CoupledInterfaceMap) Delete(key interface{}, waitChan chan struct{}, window fyne.Window) {
	if _, ok := d.Args.Data[key.(string)]; ok {
		dialog.ShowConfirm("Delete Item", "Are you sure you want to delete this item?", func(b bool) {
			if b {
				_ = d.Args.Delete(key.(string))
				delete(d.types, key.(string))
			}
			waitChan <- struct{}{}
		}, window)
	}
}

func (d *CoupledInterfaceMap) Keys() []string {
	keys := make([]string, 0, len(d.Args.Data))
	for key := range d.Args.Data {
		keys = append(keys, key)
	}
	//Sort the keys
	sort.Strings(keys)
	return keys
}

func (d *CoupledInterfaceMap) Object() interface{} {
	return d.Args
}

func NewCoupledInterfaceMap(args *NFData.NFInterfaceMap) *CoupledInterfaceMap {
	return &CoupledInterfaceMap{
		Args:  args,
		types: NewTypedMap(args.Data),
	}
}

func (d *CoupledInterfaceMap) Set(key interface{}, v interface{}) {
	d.Args.Set(key.(string), v)
	d.types[key.(string)] = GetValueType(v)
}

func (d *CoupledInterfaceMap) SetType(key interface{}, t ValueType) {
	d.types[key.(string)] = t
	switch t {
	case PropertyType:
		if _, ok := d.Args.Data[key.(string)].(*NFData.NFInterfaceMap); !ok {
			d.Args.Set(key.(string), NFData.NewNFInterfaceMap())
		}
	case MapType:
		if _, ok := d.Args.Data[key.(string)].(map[string]interface{}); !ok {
			d.Args.Set(key.(string), make(map[string]interface{}))
		}
	case SliceType:
		if _, ok := d.Args.Data[key.(string)].([]interface{}); !ok {
			d.Args.Set(key.(string), make([]interface{}, 0))
		}
	default:
		//If it cannot be converted to a string, set it to ""
		if _, ok := d.Args.Data[key.(string)].(string); !ok {
			d.Args.Set(key.(string), "")
		}
	}
}
