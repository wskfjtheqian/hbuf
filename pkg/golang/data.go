package golang

import (
	"math"
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
	List               // | type[1] | id | len[uint] | data[(type[1] | id | data[8]) | (type[1] | id | len[uint] | data)] |
	Map                // | type[1] | id | len[uint] | key[(type[1] | data[8]) | (type[1] | len[uint] | data)] | val[1] | data[(type[1] | data[8]) | (type[1] | len[uint] | data)] |
	Nil
)

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

func toBytes(t Type, id *int, call func() ([]byte, int)) []byte {
	if nil == id {
		value, valLen := call()
		var temp = make([]byte, 1+len(value))
		temp[0] = byte(int(t)<<4 | valLen)
		copy(temp[1:], value)
		return temp
	} else {
		byteId, length := getIntBytes(uint64(*id))
		value, valLen := call()
		var temp = make([]byte, 1+length+len(value))
		temp[0] = byte(int(t)<<4 | length<<2 | valLen)
		copy(temp[1:], byteId)
		copy(temp[1+length:], value)
		return temp
	}
}

func BoolToBytes(id int, val *bool) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(Int8, &id, func() ([]byte, int) {
		var temp int
		if *val {
			temp = 1
		} else {
			temp = 0
		}
		return getIntBytes(uint64(temp))
	})
}

func Int8ToBytes(id int, val *int8) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(Int8, &id, func() ([]byte, int) {
		return getIntBytes(uint64(*val))
	})
}

func Int16ToBytes(id int, val *int16) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(Int16, &id, func() ([]byte, int) {
		return getIntBytes(uint64(*val))
	})
}

func Int32ToBytes(id int, val *int32) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(Int32, &id, func() ([]byte, int) {
		return getIntBytes(uint64(*val))
	})
}

func Int64ToBytes(id int, val *int64) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(Int64, &id, func() ([]byte, int) {
		return getIntBytes(uint64(*val))
	})
}

func UInt8ToBytes(id int, val *uint8) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(UInt8, &id, func() ([]byte, int) {
		return getIntBytes(uint64(*val))
	})
}

func UInt16ToBytes(id int, val *uint16) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(UInt16, &id, func() ([]byte, int) {
		return getIntBytes(uint64(*val))
	})
}

func UInt32ToBytes(id int, val *uint32) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(UInt32, &id, func() ([]byte, int) {
		return getIntBytes(uint64(*val))
	})
}

func UInt64ToBytes(id int, val *uint64) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(UInt64, &id, func() ([]byte, int) {
		return getIntBytes(uint64(*val))
	})
}

func FloatToBytes(id int, val *float32) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(Float, &id, func() ([]byte, int) {
		return getIntBytes(uint64(math.Float32bits(*val)))
	})
}

func DoubleToBytes(id int, val *float64) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(Double, &id, func() ([]byte, int) {
		return getIntBytes(uint64(math.Float64bits(*val)))
	})
}

func StringToBytes(id int, val *string) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(String, &id, func() ([]byte, int) {
		value := []byte(*val)
		valLen := len(value)
		lenByte, lenLen := getIntBytes(uint64(valLen))
		var temp = make([]byte, valLen+lenLen)
		copy(temp, lenByte)
		copy(temp[lenLen:], value)
		return temp, lenLen
	})
}

func ListToBytes(id int, val *[]interface{}) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(List, &id, func() ([]byte, int) {
		temp := []byte{}
		for index, value := range *val {
			temp = append(temp, ToByte(Nil, &index, value)...)
		}
		return temp, len(*val)
	})
}

func MapToBytes(id int, val *[]interface{}) []byte {
	if nil == val {
		return []byte{}
	}
	return toBytes(Map, &id, func() ([]byte, int) {
		temp := []byte{}
		for key, value := range *val {
			temp = append(temp, ToByte(Nil, nil, key)...)
			temp = append(temp, ToByte(Nil, nil, value)...)
		}
		return temp, len(*val)
	})
}

func ToByte(typ Type, id *int, val interface{}) []byte {
	if nil == val {
		return []byte{}
	}
	var t Type
	var call func() ([]byte, int)
	switch val.(type) {
	case bool:
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
	case *bool:
		t = Bool
		call = func() ([]byte, int) {
			var temp int
			if *val.(*bool) {
				temp = 1
			} else {
				temp = 0
			}
			return getIntBytes(uint64(temp))
		}
	case int8:
		t = Int8
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(int8)))
		}
	case *int8:
		t = Int8
		call = func() ([]byte, int) {
			return getIntBytes(uint64(*val.(*int8)))
		}
	case int16:
		t = Int16
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(int16)))
		}
	case *int16:
		t = Int16
		call = func() ([]byte, int) {
			return getIntBytes(uint64(*val.(*int16)))
		}
	case int32:
		t = Int32
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(int32)))
		}
	case *int32:
		t = Int32
		call = func() ([]byte, int) {
			return getIntBytes(uint64(*val.(*int32)))
		}
	case int64:
		t = Int64
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(int64)))
		}
	case *int64:
		t = Int64
		call = func() ([]byte, int) {
			return getIntBytes(uint64(*val.(*int64)))
		}
	case uint8:
		t = UInt8
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(uint8)))
		}
	case *uint8:
		t = UInt8
		call = func() ([]byte, int) {
			return getIntBytes(uint64(*val.(*uint8)))
		}
	case uint16:
		t = UInt16
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(uint16)))
		}
	case *uint16:
		t = UInt16
		call = func() ([]byte, int) {
			return getIntBytes(uint64(*val.(*uint16)))
		}
	case uint32:
		t = UInt32
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(uint32)))
		}
	case *uint32:
		t = UInt32
		call = func() ([]byte, int) {
			return getIntBytes(uint64(val.(uint32)))
		}
	case uint64:
		t = UInt16
		call = func() ([]byte, int) {
			return getIntBytes(val.(uint64))
		}
	case *uint64:
		t = UInt16
		call = func() ([]byte, int) {
			return getIntBytes(*val.(*uint64))
		}
	case float32:
		t = Float
		call = func() ([]byte, int) {
			return getIntBytes(uint64(math.Float32bits(val.(float32))))
		}
	case *float32:
		t = Float
		call = func() ([]byte, int) {
			return getIntBytes(uint64(math.Float32bits(*val.(*float32))))
		}
	case float64:
		t = Double
		call = func() ([]byte, int) {
			return getIntBytes(math.Float64bits(val.(float64)))
		}
	case *float64:
		t = Double
		call = func() ([]byte, int) {
			return getIntBytes(math.Float64bits(*val.(*float64)))
		}
	case string:
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
	case *string:
		t = String
		call = func() ([]byte, int) {
			value := []byte(*val.(*string))
			valLen := len(value)
			lenByte, lenLen := getIntBytes(uint64(valLen))
			var temp = make([]byte, valLen+lenLen)
			copy(temp, lenByte)
			copy(temp[lenLen:], value)
			return temp, lenLen
		}
	}
	if Nil == typ {
		typ = t
	}
	return toBytes(typ, id, call)
}
