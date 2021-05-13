package golang

import (
	"math"
	"reflect"
)

type Type int8

const (
	Int    Type = iota // | (type + idMask+ dataSize)[1] | id | data |
	UInt               // | (type + idMask+ dataSize)[1] | id | data |
	Float              // | (type + idMask+ dataSize)[1] | id | data |
	Bytes              // | (type + idMask+ lenSize)[1]  | id | len | data |
	Array              // | (type + idMask+ lenSize)[1]  | id | len | data |
	Map                // | (type + idMask+ lenSize)[1]  | id | len | data |
	Server             // | (type + idMask+ lenSize)[1]  | id | len | data |
	Data               // | (type + idMask+ lenSize)[1]  | id | len | data |
)

type ToByteCall interface {
	ToBytes() []byte
}

// ToBytes /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func idToBytes(id uint32) ([]byte, int) {
	var temp = make([]byte, 8)
	i := 0
	for id >= 0x80 {
		temp[i] = byte(id) | 0x80
		id >>= 7
		i++
	}
	temp[i] = byte(id)
	return temp[0 : i+1], i + 1
}

func idFromBytes(bytes []byte) (uint32, int) {
	var id uint32
	var i = 0
	for _, val := range bytes {
		id |= uint32(val&0x7F) << (7 * i)
		i++
		if val < 0x80 {
			break
		}
	}
	return id, i
}

func intToBytes(id uint64) ([]byte, int) {
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

func joinBytes(t Type, id *uint32, call func() ([]byte, int)) []byte {
	if nil == id {
		valB, valL := call()
		var temp = make([]byte, 1+valL)
		temp[0] = byte(int(t)<<5 | valL)
		copy(temp[1:], valB)
		return temp
	} else {
		idB, idL := idToBytes(*id)
		valB, valL := call()
		var temp = make([]byte, 1+idL+valL)
		temp[0] = byte(int(t)<<5 | 1<<4 | valL)
		copy(temp[1:], idB)
		copy(temp[1+idL:], valB)
		return temp
	}
}

func JoinBytes(t Type, id uint32, call func() ([]byte, int)) []byte {
	return joinBytes(t, &id, call)
}

func ToBytes(typ Type, id uint32, val interface{}) []byte {
	return toBytes(&typ, &id, val)
}

func toBytes(typ *Type, id *uint32, val interface{}) []byte {
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
		t = Int
		call = func() ([]byte, int) {
			var temp int
			if val.(bool) {
				temp = 1
			} else {
				temp = 0
			}
			return intToBytes(uint64(temp))
		}

	case reflect.Int8:
		t = Int
		call = func() ([]byte, int) {
			return intToBytes(uint64(val.(int8)))
		}
	case reflect.Int16:
		t = Int
		call = func() ([]byte, int) {
			return intToBytes(uint64(val.(int16)))
		}
	case reflect.Int32:
		t = Int
		call = func() ([]byte, int) {
			return intToBytes(uint64(val.(int32)))
		}
	case reflect.Int:
		t = Int
		call = func() ([]byte, int) {
			return intToBytes(uint64(val.(int)))
		}
	case reflect.Int64:
		t = Int
		call = func() ([]byte, int) {
			return intToBytes(uint64(val.(int64)))
		}
	case reflect.Uint8:
		t = UInt
		call = func() ([]byte, int) {
			return intToBytes(uint64(val.(uint8)))
		}

	case reflect.Uint16:
		t = UInt
		call = func() ([]byte, int) {
			return intToBytes(uint64(val.(uint16)))
		}
	case reflect.Uint32:
		t = UInt
		call = func() ([]byte, int) {
			return intToBytes(uint64(val.(uint32)))
		}
	case reflect.Uint64:
		t = UInt
		call = func() ([]byte, int) {
			return intToBytes(val.(uint64))
		}
	case reflect.Float32:
		t = Float
		call = func() ([]byte, int) {
			return intToBytes(uint64(math.Float32bits(val.(float32))))
		}
	case reflect.Float64:
		t = Float
		call = func() ([]byte, int) {
			return intToBytes(math.Float64bits(val.(float64)))
		}
	case reflect.String:
		t = Bytes
		call = func() ([]byte, int) {
			valB := []byte(val.(string))
			valL := len(valB)
			lenB, lenL := intToBytes(uint64(valL))
			var temp = make([]byte, valL+lenL)
			copy(temp, valB)
			copy(temp[valL:], lenB)
			return temp, lenL
		}
	case reflect.Slice:
		t = Array
		v := reflect.ValueOf(val)
		le := v.Len()
		if 0 < le && reflect.Int8 == v.Index(0).Kind() {
			call = func() ([]byte, int) {
				valB := v.Interface().([]byte)
				valL := len(valB)
				lenB, lenL := intToBytes(uint64(valL))
				var temp = make([]byte, valL+lenL)
				copy(temp, valB)
				copy(temp[valL:], lenB)
				return temp, lenL
			}
		} else {
			call = func() ([]byte, int) {
				temp := []byte{}
				lenB, lenL := intToBytes(uint64(le))
				temp = append(temp, lenB...)
				var i uint32 = 0
				for ; i < uint32(le); i++ {
					temp = append(temp, toBytes(nil, &i, v.Index(int(i)).Interface())...)
				}
				return temp, lenL
			}
		}
	case reflect.Map:
		t = Map
		call = func() ([]byte, int) {
			temp := []byte{}
			v := reflect.ValueOf(val)
			le := v.Len()
			lenB, lenL := intToBytes(uint64(le))
			temp = append(temp, lenB...)
			for _, key := range v.MapKeys() {
				temp = append(temp, toBytes(nil, nil, key.Interface())...)
				temp = append(temp, toBytes(nil, nil, v.MapIndex(key).Interface())...)
			}
			return temp, lenL
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

//// FromBytes /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//
//func getType(b uint8) (Type, int, int) {
//	t := b >> 4
//	idLen := (b >> 2) >> 2
//	valLen := b
//	return Type(t), int(idLen), int(valLen)
//}
//
//func getUint64(buffer []byte, index int, l int) uint64 {
//	var ret uint64
//	for i := index + l - 1; i >= index; i-- {
//		index = index << 8
//	}
//	return ret
//}
//
//func FromBytes(buffer []byte, call func(typ *Type, id *int, data *interface{})) {
//	var i = 0
//	l := len(buffer)
//	for i < l {
//		ty, idLen, valLen := getType(buffer[i])
//		i++
//		var id int
//		if 0 < idLen {
//			id = int(getUint64(buffer, i, idLen))
//			i += idLen
//		}
//		var data interface{}
//		switch ty {
//		case Bool:
//			data = 1 == getUint64(buffer, i, valLen)
//		case Int8:
//			data = int8(getUint64(buffer, i, valLen))
//		case Int16:
//			data = int16(getUint64(buffer, i, valLen))
//		}
//		call(&ty, &id, &data)
//	}
//
//}
