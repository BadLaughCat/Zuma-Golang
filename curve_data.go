package main

import (
	"bytes"
	"io"
	"os"
)

type PathPoint struct {
	X, Y     float32
	Priority uint8
	InTunnel bool
}

const INV_SUBPIXEL_MULT float32 = 0.01

func LoadCurveData(filePath string) []PathPoint {
	raw, _ := os.ReadFile(filePath)
	reader := bytes.NewReader(raw)
	reader.Seek(12, io.SeekStart)
	buffer_size := ReadData[uint32](reader)
	reader.Seek(int64(buffer_size), io.SeekCurrent)

	size := ReadData[uint32](reader)
	point_list := make([]PathPoint, 0, size)
	if size > 0 {
		start_point := PathPoint{}
		start_point.X = ReadData[float32](reader)
		start_point.Y = ReadData[float32](reader)
		start_point.InTunnel = ReadData[uint8](reader) != 0
		start_point.Priority = ReadData[uint8](reader)
		point_list = append(point_list, start_point)

		ox, oy := start_point.X, start_point.Y
		for range size - 1 {
			point := PathPoint{}
			dx, dy := ReadData[int8](reader), ReadData[int8](reader)
			point.X = float32(dx)*INV_SUBPIXEL_MULT + ox
			point.Y = float32(dy)*INV_SUBPIXEL_MULT + oy
			point.InTunnel = ReadData[uint8](reader) != 0
			point.Priority = ReadData[uint8](reader)
			point_list = append(point_list, point)
			ox, oy = point.X, point.Y
		}
	}
	return point_list
}
