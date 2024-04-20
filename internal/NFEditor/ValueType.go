package NFEditor

import (
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"reflect"
	"strconv"
)

type CoupledObject interface {
	Set(key interface{}, value interface{})
	SetType(key interface{}, t ValueType)
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

func (d *CoupledSlice) Object() interface{} {
	return d.Slice
}

func NewCoupledSlice(args CustomSlice) *CoupledSlice {
	return &CoupledSlice{
		Slice: args,
		types: NewTypedSlice(args),
	}
}

func (d *CoupledSlice) Append(v interface{}) {
	d.Slice = append(d.Slice, v)
	d.types = append(d.types, GetValueType(v))
}

func (d *CoupledSlice) Prepend(v interface{}) {
	d.Slice = append(CustomSlice{v}, d.Slice...)
	d.types = append(TypedSlice{GetValueType(v)}, d.types...)
}

func (d *CoupledSlice) Set(i interface{}, v interface{}) {
	d.Slice[i.(int)] = v
	d.types[i.(int)] = GetValueType(v)
}

func (d *CoupledSlice) SetType(i interface{}, t ValueType) {
	d.types[i.(int)] = t
}

type CoupledMap struct {
	Map   CustomMap
	types TypedMap
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
}

type CoupledInterfaceMap struct {
	Args  *NFData.NFInterfaceMap
	types TypedMap
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
}
