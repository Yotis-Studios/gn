package gn

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
)

type undefined interface{}

//var stringOptions = {length: 99, zeroTerminated: true}
//var bufferOptions = {length: 99}
var typeMap = []string{"u8", "u16", "u32", "s8", "s16", "s32", "f16", "f32", "f64", "string", "buffer", "undefined"}
var sizeMap = map[string]int{"u8": 1, "u16": 2, "u32": 4, "s8": 1, "s16": 2, "s32": 4, "f16": 2, "f32": 4, "f64": 8, "undefined": 0}

func BytesFromData(data interface{}) []byte {
	var dataType = DetermineType(data)
	var typeName = typeMap[dataType]

	var buf = new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint8(dataType))
	if err != nil {
		panic(err)
	}

	if typeName == "string" {
		// String
		str := data.(string)
		strSize := len(str)
		err = binary.Write(buf, binary.LittleEndian, uint8(strSize))
		if err != nil {
			panic(err)
		}
		err = binary.Write(buf, binary.LittleEndian, []byte(str))
		if err != nil {
			panic(err)
		}
		// write null terminator
		//err = binary.Write(buf, binary.LittleEndian, uint8(0))
	} else if typeName == "buffer" {
		// Buffer array
		arr := data.([]byte)
		arrSize := len(arr)
		err = binary.Write(buf, binary.LittleEndian, uint8(arrSize))
		if err != nil {
			panic(err)
		}
		err = binary.Write(buf, binary.LittleEndian, arr)
	} else {
		// Number
		switch typeName {
		case "u8":
			u8, ok := data.(uint8)
			if !ok {
				u8 = uint8(data.(int))
			}
			err = binary.Write(buf, binary.LittleEndian, uint8(u8))
		case "u16":
			u16, ok := data.(uint16)
			if !ok {
				u16 = uint16(data.(int))
			}
			err = binary.Write(buf, binary.LittleEndian, uint16(u16))
		case "u32":
			u32, ok := data.(uint32)
			if !ok {
				u32 = uint32(data.(int))
			}
			err = binary.Write(buf, binary.LittleEndian, uint32(u32))
		case "s8":
			s8, ok := data.(int8)
			if !ok {
				s8 = int8(data.(int))
			}
			err = binary.Write(buf, binary.LittleEndian, int8(s8))
		case "s16":
			s16, ok := data.(int16)
			if !ok {
				s16 = int16(data.(int))
			}
			err = binary.Write(buf, binary.LittleEndian, int16(s16))
		case "s32":
			s32, ok := data.(int32)
			if !ok {
				s32 = int32(data.(int))
			}
			err = binary.Write(buf, binary.LittleEndian, int32(s32))
		case "f32":
			f32 := data.(float32)
			err = binary.Write(buf, binary.LittleEndian, float32(f32))
		case "f64":
			f64 := data.(float64)
			err = binary.Write(buf, binary.LittleEndian, float64(f64))
		case "undefined":
			// do nothing
		}
	}
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func DetermineType(data interface{}) int {
	switch data.(type) {
	case bool:
		return 0 // u8
	case uint8:
		return 0 // u8
	case uint16:
		return 1 // u16
	case uint32:
		return 2 // u32
	case int8:
		return 3 // s8
	case int16:
		return 4 // s16
	case int32:
		return 5 // s32
	case int:
		val := data.(int)
		absVal := math.Abs(float64(val))
		if val < 0 {
			// signed
			if absVal <= 127 {
				return 3 // s8
			} else if absVal <= 32767 {
				return 4 // s16
			}
			return 5 // s32
		} else {
			// unsigned
			if val <= 255 {
				return 0 // u8
			} else if val <= 65535 {
				return 1 // u16
			}
			return 2 // u32
		}
	// no f16 rip
	case float32:
		return 7 // f32
	case float64:
		return 8 // f64
	case string:
		return 9 // string
	case []byte:
		return 10 // buffer
	}
	return 11 // undefined
}

func Parse(r io.Reader) (value interface{}, size int) {
	var typeIdx uint8
	err := binary.Read(r, binary.LittleEndian, &typeIdx)
	if err != nil {
		panic(err)
	}
	typeName := typeMap[int(typeIdx)]

	if typeName == "undefined" {
		return *(new(undefined)), 0
	}

	switch typeName {
	case "u8":
		val := new(uint8)
		err = binary.Read(r, binary.LittleEndian, val)
		value = *val
	case "u16":
		val := new(uint16)
		err = binary.Read(r, binary.LittleEndian, val)
		value = *val
	case "u32":
		val := new(uint32)
		err = binary.Read(r, binary.LittleEndian, val)
		value = *val
	case "s8":
		val := new(int8)
		err = binary.Read(r, binary.LittleEndian, val)
		value = *val
	case "s16":
		val := new(int16)
		err = binary.Read(r, binary.LittleEndian, val)
		value = *val
	case "s32":
		val := new(int32)
		err = binary.Read(r, binary.LittleEndian, val)
		value = *val
	case "f32":
		val := new(float32)
		err = binary.Read(r, binary.LittleEndian, val)
		value = *val
	case "f64":
		val := new(float64)
		err = binary.Read(r, binary.LittleEndian, val)
		value = *val
	case "string":
		fallthrough
	case "buffer":
		bufSize := new(uint8)
		err = binary.Read(r, binary.LittleEndian, bufSize)
		if err != nil {
			panic(err)
		}
		val := make([]byte, *bufSize)
		err = binary.Read(r, binary.LittleEndian, &val)
		value = val
	}
	if err != nil {
		panic(err)
	}

	size = 0
	if typeName == "string" {
		str := string(value.([]byte))
		value = str
		size = len(str) + 2
	} else if typeName == "buffer" {
		size = len(value.([]byte)) + 1
	} else {
		size = sizeMap[typeName]
	}

	if typeName == "f32" {
		flt := value.(float32)
		value = math.Round(float64(flt)*100) / 100
	}

	if typeName == "f64" {
		flt := value.(float64)
		value = math.Round(flt*100) / 100
	}

	return value, size
}
