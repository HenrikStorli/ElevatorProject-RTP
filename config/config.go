package config

const (
	FloorCount    int = 4
	ElevatorCount     = 3
	ButtonCount       = 3
)

// Ports for network channels.
// The RX and TX ports should be equal
const (
	PeerTxPort  int = 16363
	PeerRxPort      = 16363
	BcastRxPort     = 26363
	BcastTxPort     = 26363
)

// The timeout duration [in seconds] after the elevator starts moving,
// or after the door opens
// The elevator will go into an error state when the time exceeds this timeout limit
const TimeoutStuckSec int = 10

const (
	// Estimated time [in seconds] that the elevator spends moving from one floor to another
	TravelTime int = 4
	// How long the door should remain open
	DoorOpenTime = 3
	// Number of max iterations in the elevator order simulator
	MaxTries = 5000
)

// Default hardware IP port
const DefaultIOPort int = 15657
