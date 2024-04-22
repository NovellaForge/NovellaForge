package NFData

import (
	"fyne.io/fyne/v2"
	"log"
	"reflect"
	"strconv"
)

type ValueType string

const (
	StringType   ValueType = "String"
	IntType      ValueType = "Int"
	FloatType    ValueType = "Float"
	BooleanType  ValueType = "Bool"
	MapType      ValueType = "Map"
	SliceType    ValueType = "Array"
	PropertyType ValueType = "Args"
	UnknownType  ValueType = "Unknown"
)

func (vt ValueType) String() string {
	return string(vt)
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

func GetValueType(value interface{}) ValueType {
	valType := reflect.TypeOf(value)
	valKind := valType.Kind()
	//Check if the value is an nfInterfaceMap

	if valType == reflect.TypeOf(&NFInterfaceMap{}) {
		log.Println("PropertyType")
		return PropertyType
	}

	if valKind == reflect.Ptr {
		valKind = valType.Elem().Kind()
	}
	switch valKind {
	case reflect.Slice, reflect.Array:
		return SliceType
	case reflect.Map:
		if dataMap, ok := value.(map[string]interface{}); ok {
			if len(dataMap) == 1 {
				if _, ok := dataMap["Data"].(map[string]interface{}); ok {
					return PropertyType
				}
			}
		}
		return MapType
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return IntType
	case reflect.Float32, reflect.Float64:
		return FloatType
	case reflect.Bool:
		return BooleanType
	case reflect.String:
		return StringType
	default:
		return UnknownType
	}
}

func (vt ValueType) Validator() fyne.StringValidator {
	switch vt {
	case IntType:
		return func(s string) error {
			_, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			return nil
		}
	case FloatType:
		return func(s string) error {
			_, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return err
			}
			return nil
		}
	case BooleanType:
		return func(s string) error {
			_, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			return nil
		}
	case StringType:
		return nil
	default:
		return nil
	}

}
