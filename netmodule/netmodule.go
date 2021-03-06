package netmodule

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"./network/bcast"
	"./network/localip"
	"./network/peers"
)

type MoveDirectionType int

const (
	MovingDown    MoveDirectionType = -1
	MovingStopped                   = 0
	MovingUp                        = 1
)

type MachineStateType int

type OrderStateType int

const (
	Unknown OrderStateType = iota
	New
	Accepted
	Completed
)

const (
	FloorCount    int = 4
	ElevatorCount     = 3
	ButtonCount       = 3
)

type OrderMatrix [ButtonCount][FloorCount]OrderStateType

//ElevatorState ...
type ElevatorState struct {
	ElevatorID int
	//MovingDirection MoveDirectionType
	//Floor           int
	//State           MachineStateType
	//IsFunctioning   bool
	//OrderMatrix     OrderMatrix
	TestText string
}

//NetworkPackage is...
type NetworkPackage struct {
	SenderID  int
	NewStates [ElevatorCount]ElevatorState
}

//NetworkPorts is ...
type NetworkPorts struct {
	PeerTxPort  int
	PeerRxPort  int
	BcastRxPort int
	BcastTxPort int
}

//NetworkChannels is ...
type NetworkChannels struct {
	PeerTxEnable chan bool
	PeerUpdateCh chan peers.PeerUpdate
	ReceiveCh    chan NetworkPackage
	TransmitCh   chan NetworkPackage
}

const (
	resendCount    int = 10
	resendInterval     = 15
)

//RunNetworkModule is
func RunNetworkModule(elevatorID int, networkPorts NetworkPorts,
	sendStateCh <-chan [ElevatorCount]ElevatorState,
	receivedStateCh chan<- [ElevatorCount]ElevatorState,
	disconnectingElevatorIDCh chan<- int,
	connectingElevatorIDCh chan<- int) {

	networkChannels := NetworkChannels{
		PeerTxEnable: make(chan bool),
		PeerUpdateCh: make(chan peers.PeerUpdate),
		ReceiveCh:    make(chan NetworkPackage),
		TransmitCh:   make(chan NetworkPackage),
	}

	initNetworkConnections(elevatorID, networkPorts, networkChannels)

	go sendStateUpdate(elevatorID, networkChannels, sendStateCh, resendCount, resendInterval)

	go receiveStateUpdates(elevatorID, networkChannels, receivedStateCh, true)

	go checkForPeerUpdates(elevatorID, networkChannels, disconnectingElevatorIDCh, connectingElevatorIDCh)
}

func initNetworkConnections(elevatorID int, networkPorts NetworkPorts, networkChannels NetworkChannels) {

	//TODO: try to find existing elevators and gain id from them?
	//This would probably not work, I think elevators should have fixed ids
	elevatorIDString := strconv.Itoa(elevatorID)
	go peers.Receiver(networkPorts.PeerRxPort, networkChannels.PeerUpdateCh)
	go peers.Transmitter(networkPorts.PeerTxPort, elevatorIDString, networkChannels.PeerTxEnable)

	go bcast.Transmitter(networkPorts.BcastTxPort, networkChannels.TransmitCh)
	go bcast.Receiver(networkPorts.BcastRxPort, networkChannels.ReceiveCh)
}

func sendStateUpdate(elevatorID int, networkChannels NetworkChannels, sendStateCh <-chan [ElevatorCount]ElevatorState, resendCount int, resendInterval int) {

	intervalMillis := time.Duration(resendInterval) * time.Millisecond
	for {
		select {
		case newStates := <-sendStateCh:

			newStatePackage := NetworkPackage{elevatorID, newStates}

			for i := 0; i < resendCount; i++ {
				//Queue the package for sending
				networkChannels.TransmitCh <- newStatePackage

				//Wait for next resend
				time.Sleep(intervalMillis)
			}

		}
	}
}

func receiveStateUpdates(elevatorID int, networkChannels NetworkChannels, receivedStateCh chan<- [ElevatorCount]ElevatorState, discardOwnPackages bool) {
	//TODO: Receive new states on StateReceiveCh
	//Update the current state based on whom sent it and what it contains
	for {
		select {
		case newPackage := <-networkChannels.ReceiveCh:

			//Discard any packages that was sent by this elevator
			if discardOwnPackages && newPackage.SenderID == elevatorID {
				continue
			}

			//TODO: Handle received data
			receivedStateCh <- newPackage.NewStates
		}
	}
}

//CheckForPeerUpdates is
//TODO: add interface for updating list of connected elevators
//TODO: how to sense that a newly connected peer is the missing one?
//TODO: handle peer update if the disconnected elevator is THIS elevator
func checkForPeerUpdates(elevatorID int, networkChannels NetworkChannels, disconnectingElevatorIDCh chan<- int, connectingElevatorIDCh chan<- int) {

	for {
		select {
		case peerUpdate := <-networkChannels.PeerUpdateCh:
			//peerUpdate.Lost will contain this ID if it is disconnected from the network.
			// fmt.Printf("Peer update:\n")
			// fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
			// fmt.Printf("  New:      %q\n", peerUpdate.New)
			// fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)

			//TODO: add handling of newly connected peers
			if peerUpdate.New != "" {
				newID, _ := strconv.Atoi(peerUpdate.New)
				if newID != elevatorID {
					connectingElevatorIDCh <- newID
				}
			}
			//Not sure if we need to handle that the disconnected elevator is THIS elevator here
			//In any case, send that this elevator disconnected before the others to avoid uneccessary redistribution.
			for _, lostPeer := range peerUpdate.Lost {
				lostID, _ := strconv.Atoi(lostPeer)
				if lostID == elevatorID {
					disconnectingElevatorIDCh <- lostID
					break
				}
			}
			for _, lostPeer := range peerUpdate.Lost {
				lostID, _ := strconv.Atoi(lostPeer)
				if lostID != elevatorID {
					disconnectingElevatorIDCh <- lostID
				}
			}
		}
	}
}

//GenerateID is ...
func GenerateID() string {

	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}
	id := fmt.Sprintf("elevator-%s-%d", localIP, os.Getpid())

	return id
}
