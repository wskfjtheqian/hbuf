package golang

import (
	"math"
	"reflect"
)

type Type int8

const (
	Server Type = iota // | type[1] | id | len[uint] | data |
	Data               // | type[1] | id | len[uint] | data |
	Bool               // | type[1] | id | data[1] |
	Int8               // | type[1] | id | data[1] |
	Int16              // | type[1] | id | data[2] |
	Int32              // | type[1] | id | data[4] |
	Int64              // | type[1] | id | data[8] |
	UInt8              // | type[1] | id | data[1] |
	UInt16             // | type[1] | id | data[2] |
	UInt32             // | type[1] | id | data[4] |
	UInt64             // | type[1] | id | data[8] |
	Float              // | type[1] | id | data[4] |
	Double             // | type[1] | id | data[8] |
	String             // | type[1] | id | len[uint] | data |
	Array              // | type[1] | id | len[uint] | data[(type[1] | id | data[8]) | (type[1] | id | len[uint] | data)] |
	Map                // | type[1] | id | len[uint] | key[(type[1] | data[8]) | (type[1] | len[uint] | data)] | val[1]   | data[(type[1] | data[8]) | (type[1] | len[uint] | data)] |
)

type ToByteCall interface {
	ToBytes() []byte
}

func getIntBytes(id uint64) ([]byte, int) {
	var temp = make([]byte, 8)
	if 0 == id {
		temp[0] = byte(0)
		return temp[0:1], 1
	}
	var i = 0
	for id > 0 {
		a := id & 0xFF
		temp[i] = byte(a)
		id = id >> 8
		i++
	}
	return temp[0:i], i
}

func joinBytes(t Type, id *int, call func() ([]byte, int)) []byte {
	if nil == id {
		value, valLen := call()
		var temp = make([]byte, 1+len(value))
		temp[0] = byte(int(t)<<4 | valLen)
		copy(temp[1:], value)
		return temp
	} else {
		byteId, idLen := getIntBytes(uint64(*id))
		value, valLen := call()
		var temp = make([]byte, 1+idLen+len(value))
		temp[0] = byte(int(t)<<4 | idLen<<2 | valLen)
		copy(temp[1:], byteId)
		copy(temp[1+idLen:], value)
		return temp
	}
}
func JoinBytes(t Type, id int32, call func() ([]byte, int)) []byte {
	intId := int(id)
	return joinBytes(t, &intId, call)
}

func ToBytes(typ Type, id int32, val interface{}) []byte {
	intId := int(id)
	return toBytes(&typ, &intId, val)
}

func toBytes(typ *Type, id *int, val interface{}) []byte {
	if nil == val {
		return []byte{}
	}
	var t Type
	var call func() ([]byte, int)
	tp := reflect.TypeOf(val)
	if reflect.Ptr == tp.Kind() {
		tp = tp.Elem()
		if reflect.Struct != tp.Kind() {
			val = reflect.ValueOf(val).Elem()
		}
	}
	switch tp.Kind() {
	case reflect.Bool:
		t = Bool
		call = func() ([]byte, int) {
			var temp int
			if val.(bool) {
				temp = 1
			} else {
				temp = 0
			}
			return getIntBytes(uint64(temp))
		}

	case reflect.Int8:
		t = Int8
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(int8)))
		}
	case reflect.Int16:
		t = Int16
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(int16)))
		}
	case reflect.Int32:
		t = Int32
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(int32)))
		}
	case reflect.Int:
		t = Int64
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(int)))
		}
	case reflect.Int64:
		t = Int64
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(int64)))
		}
	case reflect.Uint8:
		t = UInt8
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(uint8)))
		}

	case reflect.Uint16:
		t = UInt16
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(uint16)))
		}
	case reflect.Uint32:
		t = UInt32
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(uint32)))
		}
	case reflect.Uint64:
		t = UInt64
		call = func() ([]byte, int) {
			return getIntBytes(val.(uint64))
		}
	case reflect.Float32:
		t = Float
		call = func() ([]byte, int) {
			return getIntBytes(uint64(math.Float32bits(val.(float32))))
		}
	case reflect.Float64:
		t = Double
		call = func() ([]byte, int) {
			return getIntBytes(math.Float64bits(val.(float64)))
		}
	case reflect.String:
		t = String
		call = func() ([]byte, int) {
			value := []byte(val.(string))
			valLen := len(value)
			lenByte, lenLen := getIntBytes(uint64(valLen))
			var temp = make([]byte, valLen+lenLen)
			copy(temp, lenByte)
			copy(temp[lenLen:], value)
			return temp, lenLen
		}
	case reflect.Slice:
		t = Array
		call = func() ([]byte, int) {
			temp := []byte{}
			v := reflect.ValueOf(val)
			le := v.Len()
			for i := 0; i < le; i++ {
				temp = append(temp, toBytes(nil, &i, v.Index(i).Interface())...)
			}
			return temp, le
		}
	case reflect.Map:
		t = Map
		call = func() ([]byte, int) {
			temp := []byte{}
			v := reflect.ValueOf(val)
			le := v.Len()
			for _, key := range v.MapKeys() {
				temp = append(temp, toBytes(nil, nil, key.Interface())...)
				temp = append(temp, toBytes(nil, nil, v.MapIndex(key).Interface())...)
			}
			return temp, le
		}
	case reflect.Struct:
		var v = reflect.ValueOf(val)
		if call := v.MethodByName("ToBytes"); call.IsValid() {
			return call.Call(nil)[0].Bytes()
		}
		return []byte{}
	}
	if nil == typ {
		typ = &t
	}
	return joinBytes(*typ, id, call)
}
