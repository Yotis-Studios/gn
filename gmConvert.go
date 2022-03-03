package gn

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"sort"
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
		err = binary.Write(buf, binary.LittleEndian, data.(string))
	} else if typeName == "buffer" {
		// Buffer
		err = binary.Write(buf, binary.LittleEndian, data.([]byte))
	} else {
		// Number
		switch typeName {
		case "u8":
			err = binary.Write(buf, binary.LittleEndian, uint8(data.(int)))
			break
		case "u16":
			err = binary.Write(buf, binary.LittleEndian, uint16(data.(int)))
			break
		case "u32":
			err = binary.Write(buf, binary.LittleEndian, uint32(data.(int)))
			break
		case "s8":
			err = binary.Write(buf, binary.LittleEndian, int8(data.(int)))
			break
		case "s16":
			err = binary.Write(buf, binary.LittleEndian, int16(data.(int)))
			break
		case "s32":
			err = binary.Write(buf, binary.LittleEndian, int32(data.(int)))
			break
		case "f32":
			err = binary.Write(buf, binary.LittleEndian, float32(data.(float64)))
			break
		case "f64":
			err = binary.Write(buf, binary.LittleEndian, float64(data.(float64)))
			break
		case "undefined":
			break
		}
	}
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func DetermineType(data interface{}) int {
	switch data.(type) {
	// case undefined:
	// 	return sort.StringSlice(typeMap).Search("undefined")
	case string:
		return sort.StringSlice(typeMap).Search("string")
	case []byte:
		return sort.StringSlice(typeMap).Search("buffer")
	case int:
		val := data.(int)
		absVal := math.Abs(float64(val))
		if val < 0 {
			// signed
			if absVal <= 127 {
				return sort.StringSlice(typeMap).Search("s8")
			} else if absVal <= 32767 {
				return sort.StringSlice(typeMap).Search("s16")
			}
			return sort.StringSlice(typeMap).Search("s32")
		} else {
			// unsigned
			if val <= 255 {
				return sort.StringSlice(typeMap).Search("u8")
			} else if val <= 65535 {
				return sort.StringSlice(typeMap).Search("u16")
			}
			return sort.StringSlice(typeMap).Search("u32")
		}
	case float32:
		return sort.StringSlice(typeMap).Search("f32")
	case float64:
		return sort.StringSlice(typeMap).Search("f64")
	}

	return sort.StringSlice(typeMap).Search("undefined")
}

func Parse(r io.Reader, index int) (value interface{}, size int) {
	var typeIdx uint8
	err := binary.Read(r, binary.LittleEndian, &typeIdx)
	if err != nil {
		panic(err)
	}
	typeName := typeMap[int(typeIdx)]

	if typeName == "undefined" {
		return new(undefined), 0
	}

	switch typeName {
	case "u8":
		value = new(uint8)
		err = binary.Read(r, binary.LittleEndian, &value)
		break
	case "u16":
		value = new(uint16)
		err = binary.Read(r, binary.LittleEndian, &value)
		break
	case "u32":
		value = new(uint32)
		err = binary.Read(r, binary.LittleEndian, &value)
		break
	case "s8":
		value = new(int8)
		err = binary.Read(r, binary.LittleEndian, &value)
		break
	case "s16":
		value = new(int16)
		err = binary.Read(r, binary.LittleEndian, &value)
		break
	case "s32":
		value = new(int32)
		err = binary.Read(r, binary.LittleEndian, &value)
		break
	case "f32":
		value = new(float32)
		err = binary.Read(r, binary.LittleEndian, &value)
		break
	case "f64":
		value = new(float64)
		err = binary.Read(r, binary.LittleEndian, &value)
		break
	case "string":
		value = new(string)
		err = binary.Read(r, binary.LittleEndian, &value)
		break
	case "buffer":
		value = new([]byte)
		err = binary.Read(r, binary.LittleEndian, &value)
		break
	}
	if err != nil {
		panic(err)
	}

	size = 0
	if typeName == "string" {
		size = len(value.([]rune)) + 1
	} else if typeName == "buffer" {
		size = len(value.([]byte))
	} else {
		size = sizeMap[typeName]
	}
	if typeName == "f32" || typeName == "f64" {
		value = math.Round(value.(float64)*100) / 100
	}

	return value, size
}
