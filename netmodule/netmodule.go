package netmodule

import (
	"fmt"
	"os"
	"project-gruppe-63/netmodule/network/bcast"
	"project-gruppe-63/netmodule/network/localip"
	"project-gruppe-63/netmodule/network/peers"
	"time"
)

//Dummy struct until real is made
type ElevatorState struct {
	ElevatorId string
	TestText   string
}

type NetworkPackage struct {
	SenderId string
	NewState ElevatorState
}

type NetworkPorts struct {
	PeerTxPort  int
	PeerRxPort  int
	BcastRxPort int
	BcastTxPort int
}

type NetworkChannels struct {
	PeerTxEnable chan bool
	PeerUpdateCh chan peers.PeerUpdate
	ReceiveCh    chan NetworkPackage
	TransmitCh   chan NetworkPackage
}

func InitNetworkConnections(elevatorId string, networkPorts NetworkPorts, networkChannels NetworkChannels) {

	//TODO: try to find existing elevators and gain id from them?

	go peers.Receiver(networkPorts.PeerRxPort, networkChannels.PeerUpdateCh)
	go peers.Transmitter(networkPorts.PeerTxPort, elevatorId, networkChannels.PeerTxEnable)

	go bcast.Transmitter(networkPorts.BcastTxPort, networkChannels.TransmitCh)
	go bcast.Receiver(networkPorts.BcastRxPort, networkChannels.ReceiveCh)
}

func SendStateUpdate(elevatorId string, networkChannels NetworkChannels, sendStateCh <-chan ElevatorState, resendCount int, resendInterval int) {

	intervalMillis := time.Duration(resendInterval) * time.Millisecond
	for {
		select {
		case newState := <-sendStateCh:

			newStatePackage := NetworkPackage{elevatorId, newState}

			for i := 0; i < resendCount; i++ {
				//Queue the package for sending
				networkChannels.TransmitCh <- newStatePackage

				//Wait for next resend
				time.Sleep(intervalMillis)
			}

		}
	}
}

func ReceiveStateUpdates(elevatorId string, networkChannels NetworkChannels, receivedStateCh chan<- ElevatorState, discardOwnPackages bool) {
	//TODO: Receive new states on StateReceiveCh
	//Update the current state based on whom sent it and what it contains
	for {
		select {
		case newPackage := <-networkChannels.ReceiveCh:

			//Discard any packages that was sent by this elevator
			if discardOwnPackages && newPackage.SenderId == elevatorId {
				continue
			}

			//TODO: Handle received data
			receivedStateCh <- newPackage.NewState
		}
	}
}

//TODO: add interface for updating list of connected elevators
//TODO: how to sense that a newly connected peer is the missing one?
func CheckForPeerUpdates(elevatorId string, networkChannels NetworkChannels) {

	for {
		select {
		case peerUpdate := <-networkChannels.PeerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
			fmt.Printf("  New:      %q\n", peerUpdate.New)
			fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)

			//TODO: add handling of newly connected peers

			//TODO: add handling for disconnected peers

		}
	}
}

func GenerateId() string {

	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}
	id := fmt.Sprintf("elevator-%s-%d", localIP, os.Getpid())

	return id
}
