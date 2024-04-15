package EditorWidgets

import (
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"reflect"
	"strconv"
)

type ValueType string

const (
	StringType   ValueType = "String"
	IntType      ValueType = "Int"
	FloatType    ValueType = "Float"
	BooleanType  ValueType = "Boolean"
	MapType      ValueType = "Object"
	SliceType    ValueType = "Slice"
	StructType   ValueType = "Struct"
	PropertyType ValueType = "Args"
	UnknownType  ValueType = "Unknown"
)

func DetectValueType(value interface{}) (ValueType, string) {
	//Check if the value can be converted to a NFInterfaceMap
	if _, ok := value.(*NFData.NFInterfaceMap); ok {
		return PropertyType, "Properties..."
	}
	valType := reflect.TypeOf(value)
	if valType == nil {
		return UnknownType, "Unknown"
	}
	valKind := valType.Kind()
	switch valKind {
	case reflect.Slice, reflect.Array:
		return SliceType, "Array..."
	case reflect.Map:
		return MapType, "Object..."
	case reflect.Struct:
		return StructType, "Struct..."
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
