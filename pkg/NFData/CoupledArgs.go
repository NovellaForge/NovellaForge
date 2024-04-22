package NFData

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"sort"
	"strconv"
)

type CoupledArgs struct {
	NfInterfaceMap *NFInterfaceMap
	types          map[string]ValueType
}

func NewCoupledInterfaceMap(args *NFInterfaceMap) *CoupledArgs {
	return &CoupledArgs{
		NfInterfaceMap: args,
		types:          NewTypedMap(args.Data),
	}
}

func (c *CoupledArgs) Keys() []string {
	c.NfInterfaceMap.Lock(true)
	defer c.NfInterfaceMap.Unlock(true)
	keys := make([]string, 0, len(c.NfInterfaceMap.Data))
	for key := range c.NfInterfaceMap.Data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (c *CoupledArgs) Add() interface{} {
	c.NfInterfaceMap.Lock(false)
	defer c.NfInterfaceMap.Unlock(false)
	newKey := "newKey"
	for i := 0; ; i++ {
		if _, ok := c.NfInterfaceMap.Data[newKey]; !ok {
			break
		}
		newKey = "newKey" + strconv.Itoa(i)
	}
	c.NfInterfaceMap.Data[newKey] = ""
	c.types[newKey] = UnknownType
	return newKey
}

func (c *CoupledArgs) Object() interface{} {
	return c.NfInterfaceMap
}

func (c *CoupledArgs) Get(key interface{}) (interface{}, bool) {
	c.NfInterfaceMap.Lock(true)
	defer c.NfInterfaceMap.Unlock(true)
	if val, ok := c.NfInterfaceMap.Data[key.(string)]; ok {
		return val, true
	}
	return nil, false
}

func (c *CoupledArgs) Set(key interface{}, value interface{}) {
	c.NfInterfaceMap.Set(key.(string), value)
	c.types[key.(string)] = GetValueType(value)
}

func (c *CoupledArgs) SetType(key interface{}, t ValueType) {
	c.NfInterfaceMap.Lock(false)
	defer c.NfInterfaceMap.Unlock(false)
	c.types[key.(string)] = t
	switch t {
	case PropertyType:
		c.NfInterfaceMap.Set(key.(string), NewNFInterfaceMap())
	case MapType:
		c.NfInterfaceMap.Set(key.(string), make(CustomMap))
	case SliceType:
		c.NfInterfaceMap.Set(key.(string), make(CustomSlice, 0))
	case IntType:
		c.NfInterfaceMap.Set(key.(string), 0)
	case FloatType:
		c.NfInterfaceMap.Set(key.(string), 0.0)
	case BooleanType:
		c.NfInterfaceMap.Set(key.(string), false)
	default:
		c.NfInterfaceMap.Set(key.(string), "")
	}
}

func (c *CoupledArgs) SetKey(key interface{}, newKey interface{}) bool {
	c.NfInterfaceMap.Lock(false)
	defer c.NfInterfaceMap.Unlock(false)
	if _, ok := c.NfInterfaceMap.Data[key.(string)]; ok {
		if _, ok := c.NfInterfaceMap.Data[newKey.(string)]; ok {
			c.NfInterfaceMap.Data[key.(string)], c.NfInterfaceMap.Data[newKey.(string)] = c.NfInterfaceMap.Data[newKey.(string)], c.NfInterfaceMap.Data[key.(string)]
			c.types[key.(string)], c.types[newKey.(string)] = c.types[newKey.(string)], c.types[key.(string)]
		} else {
			c.NfInterfaceMap.Data[newKey.(string)] = c.NfInterfaceMap.Data[key.(string)]
			c.types[newKey.(string)] = c.types[key.(string)]
			delete(c.NfInterfaceMap.Data, key.(string))
			delete(c.types, key.(string))
		}
		return true
	} else {
		return false
	}
}

func (c *CoupledArgs) ParseValue(key interface{}, val string) (interface{}, error) {
	switch c.types[key.(string)] {
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

func (c *CoupledArgs) Delete(key interface{}, waitChan chan struct{}, window fyne.Window) {
	if _, ok := c.NfInterfaceMap.Data[key.(string)]; ok {
		dialog.ShowConfirm("Delete Item", "Are you sure you want to delete this item?", func(b bool) {
			if b {
				err := c.NfInterfaceMap.Delete(key.(string))
				if err != nil {
					return
				}
				delete(c.types, key.(string))
			}
			waitChan <- struct{}{}
		}, window)
	}
}

func (c *CoupledArgs) Copy(key interface{}) {
	c.NfInterfaceMap.Lock(false)
	defer c.NfInterfaceMap.Unlock(false)
	index := key.(string)
	newIndex := index + "_copy"
	for i := 0; ; i++ {
		if _, ok := c.NfInterfaceMap.Data[newIndex]; !ok {
			break
		}
		newIndex = index + "_copy" + strconv.Itoa(i)
	}
	//Check if the value being copied is a copyable type
	_, ok := c.NfInterfaceMap.Data[index].(Copyable)
	if ok {
		c.NfInterfaceMap.Data[newIndex] = c.NfInterfaceMap.Data[index].(Copyable).Copy()
		c.types[newIndex] = c.types[index]
	} else {
		c.NfInterfaceMap.Data[newIndex] = c.NfInterfaceMap.Data[index]
		c.types[newIndex] = c.types[index]
	}
}
