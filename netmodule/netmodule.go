package netmodule

import (
	"strconv"
	"time"

	cf "../config"
	dt "../datatypes"
	"./network/bcast"
	"./network/peers"
)

type networkPackage struct {
	SenderID         int
	NewState         dt.ElevatorState
	NewOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType
}

type NetworkPorts struct {
	PeerTxPort  int
	PeerRxPort  int
	BcastRxPort int
	BcastTxPort int
}

type networkChannelsType struct {
	PeerTxEnable chan bool
	PeerUpdateCh chan peers.PeerUpdate
	ReceiveCh    chan networkPackage
	TransmitCh   chan networkPackage
}

const (
	resendCount    int = 10
	resendInterval     = 15
)

//RunNetworkModule is
func RunNetworkModule(elevatorID int, networkPorts NetworkPorts,
	outgoingStateCh <-chan dt.ElevatorState,
	incomingStateCh chan<- dt.ElevatorState,
	outgoingOrderCh <-chan [cf.ElevatorCount]dt.OrderMatrixType,
	incomingOrderCh chan<- [cf.ElevatorCount]dt.OrderMatrixType,
	disconnectingElevatorIDCh chan<- int,
	connectingElevatorIDCh chan<- int) {

	networkChannels := networkChannelsType{
		PeerTxEnable: make(chan bool),
		PeerUpdateCh: make(chan peers.PeerUpdate),
		ReceiveCh:    make(chan networkPackage),
		TransmitCh:   make(chan networkPackage),
	}

	initNetworkConnections(elevatorID, networkPorts, networkChannels)

	go sendNetworkPackage(elevatorID, networkChannels, outgoingStateCh, outgoingOrderCh, resendCount, resendInterval)

	go receiveNetworkPackage(elevatorID, networkChannels, incomingStateCh, incomingOrderCh, true, true)

	go checkForPeerUpdates(elevatorID, networkChannels, disconnectingElevatorIDCh, connectingElevatorIDCh)
}

func initNetworkConnections(elevatorID int, networkPorts NetworkPorts, networkChannels networkChannelsType) {

	//TODO: try to find existing elevators and gain id from them?
	//This would probably not work, I think elevators should have fixed ids
	elevatorIDString := strconv.Itoa(elevatorID)
	go peers.Receiver(networkPorts.PeerRxPort, networkChannels.PeerUpdateCh)
	go peers.Transmitter(networkPorts.PeerTxPort, elevatorIDString, networkChannels.PeerTxEnable)

	go bcast.Transmitter(networkPorts.BcastTxPort, networkChannels.TransmitCh)
	go bcast.Receiver(networkPorts.BcastRxPort, networkChannels.ReceiveCh)
}

func sendNetworkPackage(elevatorID int, networkChannels networkChannelsType, outgoingStateCh <-chan dt.ElevatorState, outgoingOrderCh <-chan [cf.ElevatorCount]dt.OrderMatrixType, resendCount int, resendInterval int) {

	intervalMillis := time.Duration(resendInterval) * time.Millisecond
	var nilOrders [cf.ElevatorCount]dt.OrderMatrixType
	var nilState dt.ElevatorState
	for {
		select {
		case newState := <-outgoingStateCh:

			newPackage := networkPackage{elevatorID, newState, nilOrders}

			for i := 0; i < resendCount; i++ {
				//Queue the package for sending
				networkChannels.TransmitCh <- newPackage

				//Wait for next resend
				time.Sleep(intervalMillis)
			}

		case newOrders := <-outgoingOrderCh:

			newPackage := networkPackage{elevatorID, nilState, newOrders}

			for i := 0; i < resendCount; i++ {
				//Queue the package for sending
				networkChannels.TransmitCh <- newPackage

				//Wait for next resend
				time.Sleep(intervalMillis)
			}

		}
	}
}

func receiveNetworkPackage(elevatorID int, networkChannels networkChannelsType,
	incomingStateCh chan<- dt.ElevatorState,
	incomingOrderCh chan<- [cf.ElevatorCount]dt.OrderMatrixType,
	discardOwnPackages bool,
	discardRepeatingPackages bool) {

	var nilOrders [cf.ElevatorCount]dt.OrderMatrixType
	var nilState dt.ElevatorState

	var lastPackageReceived networkPackage

	for {
		select {
		case newPackage := <-networkChannels.ReceiveCh:

			//discard repeating packages
			if discardRepeatingPackages && newPackage == lastPackageReceived {
				continue
			}

			//Discard any packages that was sent by this elevator
			if discardOwnPackages && newPackage.SenderID == elevatorID {
				continue
			}
			//Check if senders ID is correct
			if IsValidID(newPackage.SenderID) {
				if newPackage.NewState != nilState {

					if IsValidID(newPackage.NewState.ElevatorID) {
						incomingStateCh <- newPackage.NewState
					}
				}
				if newPackage.NewOrderMatrices != nilOrders {
					incomingOrderCh <- newPackage.NewOrderMatrices
				}
			}
			lastPackageReceived = newPackage
		}
	}
}

func checkForPeerUpdates(elevatorID int, networkChannels networkChannelsType, disconnectingElevatorIDCh chan<- int, connectingElevatorIDCh chan<- int) {

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

				if IsValidID(newID) {
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
				if lostID != elevatorID && IsValidID(lostID) {
					disconnectingElevatorIDCh <- lostID
				}
			}
		}
	}
}

func IsValidID(elevatorID int) bool {
	if elevatorID < 0 && elevatorID > cf.ElevatorCount-1 {
		return false
	}
	return true
}
