package NFData

type KeyVal struct {
	Key   string
	Value interface{}
}

// NewKeyVal creates a new KeyVal struct
func NewKeyVal(key string, value interface{}) KeyVal {
	return KeyVal{
		Key:   key,
		Value: value,
	}
}
