package gol

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns         int
	Threads       int
	ImageWidth    int
	ImageHeight   int
	ServerAddress string
	Port          string
	OurIP         string
	VisualUpdates bool
	ResumeGame    bool
}

// Find the server address as an env variable
func getServerAddressFromEnvs() string {
	return "localhost:8020"
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {
	if p.OurIP == "" {
		p.OurIP = "localhost"
	}
	println("Our IP Address: ", p.OurIP)
	if p.Port == "" {
		p.Port = "8050"
	}
	if p.ServerAddress == "" {
		p.ServerAddress = getServerAddressFromEnvs()
	}

	ioCommand := make(chan ioCommand)
	ioIdle := make(chan bool)
	ioFilename := make(chan string)
	ioImageInput := make(chan uint8)
	ioImageOutput := make(chan uint8)

	controllerChannels := controllerChannels{
		events,
		ioCommand,
		ioIdle,
		ioFilename,
		ioImageInput,
		ioImageOutput,
		keyPresses,
	}
	go controller(p, controllerChannels)

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: ioFilename,
		output:   ioImageOutput,
		input:    ioImageInput,
	}
	go startIo(p, ioChannels)

	distributorChannels := distributorChannels{
		events:     events,
		ioCommand:  ioCommand,
		ioIdle:     ioIdle,
		ioFilename: ioFilename,
		ioOutput:   nil,
		ioInput:    nil,
	}
	distributor(p, distributorChannels)
}
