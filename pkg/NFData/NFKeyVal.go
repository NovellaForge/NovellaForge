package NFData

type NFKeyVal struct {
	Key   string
	Value interface{}
}

// NewKeyVal creates a new NFKeyVal struct
func NewKeyVal(key string, value interface{}) NFKeyVal {
	return NFKeyVal{
		Key:   key,
		Value: value,
	}
}
