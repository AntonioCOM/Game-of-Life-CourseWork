package main

import (
	"sync"
	"uk.ac.bris.cs/gameoflife/stubs"
)

// Calculate the next cell state for all cells within bounds
func updateRegion(start, end int, halo stubs.Halo, newBoard [][]bool, width int, board []byte, wg *sync.WaitGroup) {

	for row := start; row < end; row++ {
		newBoard[row] = make([]bool, width)
		for col := 0; col < width; col++ {

			newCell := nextCellState(col, row+halo.Offset, board, halo.BitBoard.NumRows, halo.BitBoard.RowLength)

			newBoard[row][col] = newCell
		}
	}
	wg.Done()
}

// Calculate the next cell state according to Game Of Life rules
// Returns a bool with the next state of the cell
func nextCellState(x int, y int, board []byte, bHeight, bWidth int) bool {

	adj := countAliveNeighbours(x, y, board, bHeight, bWidth)

	newState := false

	if stubs.GetBitArrayCell(board, bHeight, bWidth, y, x) == true {
		if adj == 2 || adj == 3 {

			newState = true
		}
	} else {
		if adj == 3 {

			newState = true
		}
	}
	return newState
}

// Count how many alive neighbours a cell has
// This will correctly wrap around edges
func countAliveNeighbours(x int, y int, board []byte, height, width int) int {
	numNeighbours := 0

	for _x := -1; _x < 2; _x++ {
		for _y := -1; _y < 2; _y++ {

			if _x == 0 && _y == 0 {
				continue
			}

			wrapX := (x + _x) % width

			if wrapX == -1 {
				wrapX = width - 1
			}

			wrapY := (y + _y) % height
			if wrapY == -1 {
				wrapY = height - 1
			}

			v := stubs.GetBitArrayCell(board, height, width, wrapY, wrapX)
			if v == true {
				numNeighbours++
			}
		}
	}

	return numNeighbours
}
