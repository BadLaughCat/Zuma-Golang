package main

import (
	"bytes"
	"encoding/binary"
	"maps"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func rect(x, y, width, height int32) rl.Rectangle {
	return rl.NewRectangle(float32(x), float32(y), float32(width), float32(height))
}

func vec2(x, y int32) rl.Vector2 {
	return rl.NewVector2(float32(x), float32(y))
}

func randomly_get_map_key(theMap map[int32]int32) int32 {
	rd := int32(rand.Intn(len(theMap)))
	var index, value int32
	for value = range maps.Keys(theMap) {
		if index == rd {
			break
		}
		index++
	}
	return value
}

func widen_for_index[T int32 | float32](array *[]T, index int) {
	if index >= len(*array) {
		for range index - len(*array) + 1 {
			*array = append(*array, 0)
		}
	}
}

func ReadData[T int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64 | float32 | float64](reader *bytes.Reader) T {
	var tmp T
	binary.Read(reader, binary.LittleEndian, &tmp)
	return tmp
}
