package EditorWidgets

import "strconv"

type ValueType string

const (
	StringType  ValueType = "String"
	IntType     ValueType = "Int"
	FloatType   ValueType = "Float"
	BooleanType ValueType = "Boolean"
	UnknownType ValueType = "Unknown"
)

func DetectValueType(value interface{}) (ValueType, string) {
	switch V := value.(type) {
	case string:
		return StringType, V
	case int:
		return IntType, strconv.Itoa(V)
	case float64:
		return FloatType, strconv.FormatFloat(V, 'f', -1, 64)
	case bool:
		return BooleanType, strconv.FormatBool(V)
	default:
		return UnknownType, "Object..."
	}
}
