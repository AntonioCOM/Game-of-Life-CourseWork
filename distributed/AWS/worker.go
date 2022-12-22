package main

import (
	"flag"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

// Global variables
var (
	server        *rpc.Client
	serverAddress string
	ourAddress    string
)

// Worker is the struct for our RPC server
type Worker struct{}

// DoTurn is called by the server when it wants to calculate a new turn
// It will pass the board and fragment pointers
func (w *Worker) DoTurn(req stubs.DoTurnRequest, res *stubs.DoTurnResponse) (err error) {

	frag := doTurn(req.Halo, req.Threads)
	res.Frag = frag
	return
}

// Shutdown is called by the server to disconnect and close the worker
func (w *Worker) Shutdown(req stubs.Empty, res *stubs.Empty) (err error) {
	println("Received shutdown request")
	server.Close()
	os.Exit(0)
	return
}

// Main worker loop
func main() {
	defer println("Closing worker")

	portPtr := flag.String("p", "8020", "port to listen on")

	localhost := flag.Bool("localhost", false, "set to true if we want to use localhost")

	serverAddressPtr := flag.String("s", "localhost:8030", "server address")

	flag.Parse()

	// If the localhost flag is set, run on localhost
	if *localhost {
		ourAddress = "localhost:" + *portPtr
	} else {

		ourAddress = util.GetPublicIP() + ":" + *portPtr
	}
	serverAddress = *serverAddressPtr
	println("Starting worker (" + ourAddress + ")")

	rpc.Register(&Worker{})

	// Create a listener to handle rpc requests
	listener, _ := net.Listen("tcp", ":"+*portPtr)
	defer listener.Close()

	go rpc.Accept(listener)

	// Try and connect to the server for the first time
	connectToServer()

	pingTicker := time.NewTicker(10 * time.Second)
	for {
		select {
		// Ping the server at an interval
		case <-pingTicker.C:

			if server != nil {

				err := server.Call(stubs.ServerPing, stubs.Empty{}, &stubs.Empty{})

				//If there is an error in pinging them, we have lost connection
				if err != nil {
					println("Error pinging server:", err.Error())

					server.Close()
					server = nil
					println("Disconnected")
				}
			} else {

				connectToServer()
			}

		}
	}
}

// Attempt to connect to the server
// Returns true if we successfully connected
// This will also set the server global variable
func connectToServer() bool {
	println("Attempting to connect to server ", serverAddress)

	newServer, err := rpc.Dial("tcp", serverAddress)

	if err != nil {
		println("Cannot find server:", err.Error())
		return false
	}
	server = newServer
	response := new(stubs.ServerResponse)

	err = server.Call(stubs.ServerConnectWorker,
		stubs.WorkerConnectRequest{WorkerAddress: ourAddress}, response)
	if err != nil {
		println("Connection error", err.Error())
		return false
	} else if response.Success == false {
		println("Server error", response.Message)
		return false
	}

	// No errors, connection successful!
	println("Connected!")
	return true
}

// GAME LOGIC BELOW

// Calculate the next turn, given pointers to the start and end to operate over
// Return a fragment of the board with the next turn's cells
func doTurn(halo stubs.Halo, threads int) (boardFragment stubs.Fragment) {
	width := halo.BitBoard.RowLength
	board := halo.BitBoard.Bytes.Decode()
	height := halo.EndPtr - halo.StartPtr
	newBoard := make([][]bool, height)

	if threads > height {
		threads = height
	}
	var wg sync.WaitGroup

	fragHeight := height / threads
	for i := 0; i < threads; i++ {

		start := i * fragHeight
		end := (i + 1) * fragHeight
		if i == threads-1 {
			end = height
		}

		wg.Add(1)

		go updateRegion(start, end, halo, newBoard, width, board, &wg)
	}

	// Wait for all threads to finish
	wg.Wait()

	boardFragment = stubs.Fragment{
		StartRow: halo.StartPtr,
		EndRow:   halo.EndPtr,
		BitBoard: stubs.BitBoardFromSlice(newBoard, halo.EndPtr-halo.StartPtr, width),
	}
	return boardFragment
}
