package util

import (
	"io/ioutil"
	"strconv"
	"strings"
)

// Cell is used as the return type for the testing framework.
type Cell struct {
	X, Y int
}

// GetAliveCells returns all the alive cells in a board
func GetAliveCells(board [][]bool) []Cell {
	height := len(board)
	width := len(board[0])
	aliveCells := make([]Cell, 0)
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			if board[row][col] == true {
				aliveCells = append(aliveCells, Cell{X: col, Y: row})
			}
		}
	}
	return aliveCells
}

func ReadAliveCells(path string, width, height int) []Cell {
	//data, ioError := ioutil.ReadFile("check/images/" + fmt.Sprintf("%vx%vx%v.pgm", width, height, turns))
	data, ioError := ioutil.ReadFile(path)
	Check(ioError)

	fields := strings.Fields(string(data))

	if fields[0] != "P5" {
		panic("Not a pgm file")
	}

	imageWidth, _ := strconv.Atoi(fields[1])
	if imageWidth != width {
		panic("Incorrect width")
	}

	imageHeight, _ := strconv.Atoi(fields[2])
	if imageHeight != height {
		panic("Incorrect height")
	}

	maxval, _ := strconv.Atoi(fields[3])
	if maxval != 255 {
		panic("Incorrect maxval/bit depth")
	}

	image := []byte(fields[4])

	var cells []Cell
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := image[0]
			if cell != 0 {
				cells = append(cells, Cell{
					X: x,
					Y: y,
				})
			}
			image = image[1:]
		}
	}
	return cells
}
