package netmodule

import (
	"fmt"
	"os"
	"strconv"
	"time"

	dt "../datatypes"
	"./network/bcast"
	"./network/localip"
	"./network/peers"
)

type networkPackage struct {
	SenderID         int
	NewStates        [dt.ElevatorCount]dt.ElevatorState
	NewOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType
}

//NetworkPorts is ...
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
	outgoingStateCh <-chan [dt.ElevatorCount]dt.ElevatorState,
	incomingStateCh chan<- [dt.ElevatorCount]dt.ElevatorState,
	outgoingOrderCh <-chan [dt.ElevatorCount]dt.OrderMatrixType,
	incomingOrderCh chan<- [dt.ElevatorCount]dt.OrderMatrixType,
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

func sendNetworkPackage(elevatorID int, networkChannels networkChannelsType, outgoingStateCh <-chan [dt.ElevatorCount]dt.ElevatorState, outgoingOrderCh <-chan [dt.ElevatorCount]dt.OrderMatrixType, resendCount int, resendInterval int) {

	intervalMillis := time.Duration(resendInterval) * time.Millisecond
	var nilOrders [dt.ElevatorCount]dt.OrderMatrixType
	var nilStates [dt.ElevatorCount]dt.ElevatorState
	for {
		select {
		case newStates := <-outgoingStateCh:

			newPackage := networkPackage{elevatorID, newStates, nilOrders}

			for i := 0; i < resendCount; i++ {
				//Queue the package for sending
				networkChannels.TransmitCh <- newPackage

				//Wait for next resend
				time.Sleep(intervalMillis)
			}

		case newOrders := <-outgoingOrderCh:

			newPackage := networkPackage{elevatorID, nilStates, newOrders}

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
	incomingStateCh chan<- [dt.ElevatorCount]dt.ElevatorState,
	incomingOrderCh chan<- [dt.ElevatorCount]dt.OrderMatrixType,
	discardOwnPackages bool,
	discardRepeatingPackages bool) {

	var nilOrders [dt.ElevatorCount]dt.OrderMatrixType
	var nilStates [dt.ElevatorCount]dt.ElevatorState

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

			if newPackage.NewStates != nilStates {
				incomingStateCh <- newPackage.NewStates
			}
			if newPackage.NewOrderMatrices != nilOrders {
				incomingOrderCh <- newPackage.NewOrderMatrices
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
