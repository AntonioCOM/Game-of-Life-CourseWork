package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

// This file contains game loop functions. RPC and others are in server.go

func doWorker(halo stubs.Halo, newBoard [][]bool, threads int, worker *worker, failChan chan<- bool, fragChan chan<- stubs.Fragment) {
	response := stubs.DoTurnResponse{}

	// Send the halo to the client, get the result
	err := worker.Client.Call(stubs.WorkerDoTurn,
		stubs.DoTurnRequest{Halo: halo, Threads: threads}, &response)
	if err != nil {
		println("Error getting fragment:", err.Error())
		disconnectWorker(worker)
		failChan <- true
		return
	}
	fragChan <- response.Frag
}

// Create a "halo" of cells containing only the cells required to calculat the next turn
// Take the whole board and return a halo which can be passed to a worker
func makeHalo(worker int, fragHeight int, numWorkers int, height, width int, board [][]bool) stubs.Halo {
	cells := make([][]bool, 0)

	start := worker * fragHeight
	end := (worker + 1) * fragHeight
	if worker == numWorkers-1 {
		end = height
	}

	downPtr := end % height // "max row + 1"
	upPtr := (start - 1)    // "min row - 1"
	if upPtr == -1 {
		upPtr = height - 1
	}
	workPtr := 0

	if upPtr != end-1 {
		cells = append(cells, board[upPtr])
		workPtr = 1
	}
	for row := start; row < end; row++ {
		cells = append(cells, board[row])
	}
	if downPtr != start {
		cells = append(cells, board[downPtr])
	}
	return stubs.Halo{
		BitBoard: stubs.BitBoardFromSlice(cells, len(cells), width),
		Offset:   workPtr,
		StartPtr: start,
		EndPtr:   end,
	}
}

func updateBoard(board [][]bool, newBoard [][]bool, height, width int, threads int) bool {
	// Create a WaitGroup so we only return when all workers have finished
	var wg sync.WaitGroup
	failChan := make(chan bool)
	workersMutex.Lock()

	// Calculate the number of rows each worker thread should use
	numWorkers := len(workers)

	if numWorkers == 0 {
		return false
	}
	fragHeight := height / numWorkers

	wg.Add(numWorkers)
	fragChan := make(chan stubs.Fragment, numWorkers)

	for w := 0; w < numWorkers; w++ {
		thisWorker := workers[w]
		go func(workerIdx int, worker *worker) {

			halo := makeHalo(workerIdx, fragHeight, numWorkers, height, width, board)
			// Send the fragment to the worker
			doWorker(halo, newBoard, threads, worker, failChan, fragChan)
		}(w, thisWorker)
	}

	workersMutex.Unlock()

	i := 0
	fail := false
	for i < numWorkers {
		select {
		case fail = <-failChan:
			i++
		case frag := <-fragChan:

			respCells := frag.BitBoard.ToSlice()
			for row := frag.StartRow; row < frag.EndRow; row++ {
				copy(newBoard[row], respCells[row-frag.StartRow])
			}
			i++
		}
	}

	if fail {
		// One or more of the workers have hit a problem
		return false
	}

	return true
}

func controllerLoop(board [][]bool, startTurn, height, width, maxTurns, threads int, visualUpdates bool) {

	defer func() {

		controllerMutex.Lock()
		controller.Close()
		controller = nil
		controllerMutex.Unlock()
		println("Disconnected Controller")
	}()

	ticker := time.NewTicker(2 * time.Second)

	turn := startTurn

	newBoard := make([][]bool, height)
	for row := 0; row < height; row++ {
		newBoard[row] = make([]bool, width)
	}
	println("Max turns: ", maxTurns)

	if visualUpdates {
		controller.Call(stubs.ControllerTurnComplete,
			stubs.BoardStateReport{CompletedTurns: turn, Board: stubs.BitBoardFromSlice(board, height, width)}, &stubs.Empty{})
	}

	for turn < maxTurns {
		select {

		case key := <-keypresses:
			println("Received keypress: ", key)
			quit := handleKeypress(key, turn, board, height, width)
			if quit {
				return
			}

		case <-ticker.C:
			println("Telling controller number of cells alive")

			err := controller.Call(stubs.ControllerReportAliveCells,
				stubs.AliveCellsReport{CompletedTurns: turn, NumAlive: len(util.GetAliveCells(board))}, &stubs.Empty{})

			if err != nil {
				fmt.Println("Error sending num alive ", err)
				return
			}

		default:

			success := updateBoard(board, newBoard, height, width, threads)

			if success {

				for row := 0; row < height; row++ {
					copy(board[row], newBoard[row])
				}
				if visualUpdates {

					controller.Call(stubs.ControllerTurnComplete,
						stubs.BoardStateReport{CompletedTurns: turn, Board: stubs.BitBoardFromSlice(board, height, width)}, &stubs.Empty{})
				}
				turn++

				lastBoardState = board
				lastTurn = turn
			} else {
				if len(workers) == 0 {
					return
				}

				println("Encountered a problem handling turn", turn)
				println("Retrying this turn")
			}
		}

	}

	println("All turns done, send final turn complete")

	err := controller.Call(stubs.ControllerFinalTurnComplete,
		stubs.BoardStateReport{
			CompletedTurns: maxTurns,
			Board:          stubs.BitBoardFromSlice(board, height, width),
		},
		&stubs.Empty{})
	if err != nil {
		fmt.Println("Error sending final turn complete ", err)
	}
	// End the game
	return
}

func randomiseBoard(board [][]bool, height, width int) {
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			// Get a random number from 0.0-1.0
			r := rand.Float32()

			ratio := float32(0.2)
			if r < ratio {
				board[row][col] = true
			} else {
				board[row][col] = false
			}
		}
	}
}

// Cleanly disconnect a worker and remove it from the workers slice
func disconnectWorker(worker *worker) {
	// Lock the workers slice to get exclusive access
	workersMutex.Lock()
	defer workersMutex.Unlock()

	for w := 0; w < len(workers); w++ {
		if workers[w].Address == worker.Address {
			// Try and close the RPC connection
			worker.Client.Close()

			workers = append(workers[:w], workers[w+1:]...)
			println("Worker", worker.Address, "disconnected")
			return
		}
	}

	println("We aren't connected to worker", worker.Address)
}

// Handle keypress sent from the client
func handleKeypress(key rune, turn int, board [][]bool, height, width int) bool {
	switch key {
	case 'q':

		controller.Call(stubs.ControllerGameStateChange,
			stubs.StateChangeReport{Previous: stubs.Executing, New: stubs.Quitting, CompletedTurns: turn}, &stubs.Empty{})
		println("Closing controller")
		return true
	case 'p':
		// Pause: pause execution and wait for another P
		println("Pausing execution")

		controller.Call(stubs.ControllerGameStateChange,
			stubs.StateChangeReport{Previous: stubs.Executing, New: stubs.Paused, CompletedTurns: turn}, &stubs.Empty{})
		// Wait for another P
		for <-keypresses != 'p' {
		}
		// Tell the controller we're resuming
		controller.Call(stubs.ControllerGameStateChange,
			stubs.StateChangeReport{Previous: stubs.Paused, New: stubs.Executing, CompletedTurns: turn}, &stubs.Empty{})
		println("Resuming execution")
	case 's':

		println("Telling controller to save board")

		controller.Call(stubs.ControllerSaveBoard,
			stubs.BoardStateReport{CompletedTurns: turn, Board: stubs.BitBoardFromSlice(board, height, width)}, &stubs.Empty{})
	case 'k':

		println("Controller wants to close everything")

		for w := 0; w < len(workers); w++ {
			println("Disconnecting worker", w)

			workers[w].Client.Call(stubs.WorkerShutdown, stubs.Empty{}, &stubs.Empty{})
			workers[w].Client.Close()
		}

		controller.Call(stubs.ControllerFinalTurnComplete,
			stubs.BoardStateReport{
				CompletedTurns: turn,
				Board:          stubs.BitBoardFromSlice(board, height, width),
			},
			&stubs.Empty{})

		// Closing our listener will close our RPC serfver
		listener.Close()
		return true

	case 'r':

		println("Randomising Board")
		randomiseBoard(board, height, width)
	}
	return false
}
