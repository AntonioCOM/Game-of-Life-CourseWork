package stubs

// Fragment stores a section of cells in the board
// StartRow points to the row in the main board where this section starts
// EndRow points to the next row in the main board after this section ends (like an exclusive upper bound)
type Fragment struct {
	StartRow int
	EndRow   int
	BitBoard *BitBoard
}

// Halo is a subset of a board containing all the cells required to calculate the next turn cells
// between two parts of the board
// It stores the board state using a BitBoard, to save space
type Halo struct {
	BitBoard *BitBoard
	Offset   int
	StartPtr int
	EndPtr   int
}


// State represents a change in the state of execution.
type State int

const (
	Paused State = iota
	Executing
	Quitting
)

// String methods allow the different types of Events and States to be printed.
func (state State) String() string {
	switch state {
	case Paused:
		return "Paused"
	case Executing:
		return "Executing"
	case Quitting:
		return "Quitting"
	default:
		return "Incorrect State"
	}
}

//    RPC STRINGS

// Server RPC strings
var ServerStartGame = "Server.StartGame"
var ServerRegisterKeypress = "Server.RegisterKeypress"
var ServerConnectWorker = "Server.ConnectWorker"
var ServerPing = "Server.Ping"

// Controller RPC strings
var ControllerGameStateChange = "Controller.GameStateChange"
var ControllerTurnComplete = "Controller.TurnComplete"
var ControllerFinalTurnComplete = "Controller.FinalTurnComplete"
var ControllerSaveBoard = "Controller.SaveBoard"
var ControllerReportAliveCells = "Controller.ReportAliveCells"

// Worker RPC strings
var WorkerDoTurn = "Worker.DoTurn"
var WorkerShutdown = "Worker.Shutdown"

// ServerResponse contains a result from a standard server RPC call
// Success indicates if the call executed its desired function
// Message contains any additional information
type ServerResponse struct {
	Success bool
	Message string
}

// StartGameRequest contains all data required for a controller to connect to a server
// and start a game
// This will send the address of the controller, along with information about the board
// and the starting board state
type StartGameRequest struct {
	ControllerAddress string

	Height        int
	Width         int
	MaxTurns      int
	Threads       int
	VisualUpdates bool

	StartNew bool
	Board    *BitBoard
}

// KeypressRequest is used to send a keypress from a controller to be handled at the server
type KeypressRequest struct {
	Key rune
}

// WorkerConnectRequest is passed by a worker which wishes to connect to the server
// This contains the address of the worker so the server can establish a connection
type WorkerConnectRequest struct {
	WorkerAddress string
}

// StateChangeReport is passed to the controller to inform them of changes to game state
type StateChangeReport struct {
	Previous       State
	New            State
	CompletedTurns int
}

// TurnCompleteReport is passed to the controller every time a turn is completed
type TurnCompleteReport struct {
	CompletedTurns int
}

// BoardStateReport is passed to the controller to give them the state of the board
type BoardStateReport struct {
	CompletedTurns int

	Board *BitBoard
}

// AliveCellsReport is passed to the controller every 2 seconds to tell them how many
// cells are alive
type AliveCellsReport struct {
	CompletedTurns int
	NumAlive       int
}

// DoTurnRequest is passed to workers to ask them to calculate the next turn
// It sends the whole board along with fragment pointers for their portion to calculate
type DoTurnRequest struct {
	Halo    Halo
	Threads int
}

// DoTurnResponse is returned by workers to the server containing a fragment of the new board
type DoTurnResponse struct {
	Frag Fragment
}

// Empty is used when there is no information for an RPC function to return
type Empty struct{}

