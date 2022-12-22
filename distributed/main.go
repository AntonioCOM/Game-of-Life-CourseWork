package main

import (
	"flag"
	"fmt"
	"runtime"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/sdl"
)

// main is the function called when starting Game of Life with 'go run .'
func main() {
	runtime.LockOSThread()
	var params gol.Params

	flag.IntVar(
		&params.Threads,
		"t",
		8,
		"Specify the number of worker threads to use. Defaults to 8.")

	flag.IntVar(
		&params.ImageWidth,
		"w",
		512,
		"Specify the width of the image. Defaults to 512.")

	flag.IntVar(
		&params.ImageHeight,
		"h",
		512,
		"Specify the height of the image. Defaults to 512.")

	flag.IntVar(
		&params.Turns,
		"turns",
		10000000000,
		"Specify the number of turns to process. Defaults to 10000000000.")

	// Get the server address from the commandline
	flag.StringVar(
		&params.ServerAddress,
		"server",
		"localhost:8030",
		"Specify the address of the server. Defaults to localhost:8020")
	// Get our RPC port from the commandline
	flag.StringVar(
		&params.Port,
		"port",
		"8030",
		"Specify our port. Defaults to 8030")

	flag.BoolVar(&params.VisualUpdates,
		"sdl",
		true,
		"Specify whether or not to use SDL")

	flag.BoolVar(&params.ResumeGame,
		"resume",
		false,
		"Disables the SDL window, so there is no visualisation during the tests.")

	flag.Parse()

	fmt.Println("Threads:", params.Threads)
	fmt.Println("Width:", params.ImageWidth)
	fmt.Println("Height:", params.ImageHeight)
	fmt.Println("Server:", params.ServerAddress)
	fmt.Println("RPC Port:", params.Port)

	keyPresses := make(chan rune, 10)
	events := make(chan gol.Event, 1000)

	go gol.Run(params, events, keyPresses)

	sdl.Run(params, events, keyPresses)
}
