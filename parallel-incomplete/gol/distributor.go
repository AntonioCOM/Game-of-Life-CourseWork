package gol

import (
	"fmt"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// Done: Create a 2D slice to store the world.
	// call IOInput ie read image, send filname and recieve byte by byte the image and store in world

	c.ioCommand <- ioInput
	c.ioFilename <- fmt.Sprintf("%dx%d", p.ImageWidth, p.ImageHeight)
	world := newWorld(p.ImageHeight, p.ImageWidth)
	turn := 0

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			world[y][x] = <-c.ioInput
		}
	}

	// TODO: Execute all turns of the Game of Life.

	if p.Threads == 1 {
		for i := 0; i < p.Turns; i++ {
			world = calculateNextState(world, 0, p.ImageWidth)
			turn += 1
		}
	} else {
		workerHeight := p.ImageHeight / p.Threads
		r := p.ImageHeight % p.Threads
		out := make([]chan [][]uint8, p.Threads)
		worldB := newWorld(0, 0)
		for i := range out {
			out[i] = make(chan [][]uint8)
		}

		for i := 0; i < p.Turns; i++ {
			for j := 0; j < p.Threads; j++ {
				if (p.Threads - j) <= r {
					go worker(world, j*workerHeight, (1+j)*workerHeight+1, out[j]) /// fix!!!!
				} else {
					go worker(world, j*workerHeight, (1+j)*workerHeight, out[j])
				}
			}
			for t := 0; t < p.Threads; t++ {
				part := <-out[t]
				worldB = append(worldB, part...)
			}
			fmt.Println("Len:", len(worldB))
			world = worldB
			turn += 1
		}
	}

	// TODO: Report the final state using FinalTurnCompleteEvent.
	alive := make([]util.Cell, 0)
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if int(world[y][x]) > 0 {
				cell := util.Cell{x, y}
				alive = append(alive, cell)
			}
		}
	}

	FTCevent := FinalTurnComplete{turn, alive}
	c.events <- FTCevent
	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func calculateNextState(world [][]uint8, startY, endY int) [][]uint8 {
	sumAlive := 0
	IH := len(world)
	IW := len(world[0])
	worldB := newWorld(IH, IW)
	for y := startY; y < endY; y++ {
		for x := 0; x < IW; x++ {
			sumAlive = 0
			sumAlive = int(world[(y+IH+1)%IH][(x+IW-1)%IW]) + int(world[(y+IH+1)%IH][(x+IW)%IW]) +
				int(world[(y+IH+1)%IH][(x+IW+1)%IW]) + int(world[(y+IH)%IH][(x+IW-1)%IW]) +
				int(world[(y+IH)%IH][(x+IW+1)%IW]) + int(world[(y+IH-1)%IH][(x+IW-1)%IW]) +
				int(world[(y+IH-1)%IH][(x+IW)%IW]) + int(world[(y+IH-1)%IH][(x+IW+1)%IW])
			if int(world[y][x]) > 0 {
				if sumAlive < 510 {
					worldB[y][x] = 0
				} else if sumAlive == 510 || sumAlive == 765 {
					worldB[y][x] = 255
				} else {
					worldB[y][x] = 0
				}
			} else {
				if sumAlive == 765 {
					worldB[y][x] = 255
				} else {
					worldB[y][x] = 0
				}
			}
		}
	}
	worldB = worldB[startY:endY]
	return worldB
}
func newWorld(imageHeight, imageWidth int) [][]uint8 {
	world := make([][]uint8, imageHeight)
	for i := 0; i < imageHeight; i++ {
		world[i] = make([]uint8, imageWidth)
	}
	return world
}

func worker(world [][]uint8, startY, endY int, out chan<- [][]uint8) {
	part := calculateNextState(world, startY, endY)
	out <- part
}
