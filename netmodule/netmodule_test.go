package netmodule_test

import (
	"fmt"
	"project-gruppe-63/netmodule"
	"project-gruppe-63/netmodule/network/peers"
	"testing"
	"time"
)

func TestNetworkModule(*testing.T) {

	fmt.Println("Testing Network Module")

	ports := netmodule.NetworkPorts{16363, 16363, 26363, 26363}

	networkChs := netmodule.NetworkChannels{
		PeerTxEnable: make(chan bool),
		PeerUpdateCh: make(chan peers.PeerUpdate),
		ReceiveCh:    make(chan netmodule.NetworkPackage),
		TransmitCh:   make(chan netmodule.NetworkPackage),
	}

	id1 := netmodule.GenerateId()
	fmt.Println("Intializing connections")
	netmodule.InitNetworkConnections(id1, ports, networkChs)

	fmt.Println("Suceeded intializing connections")

	sendStateCh := make(chan netmodule.ElevatorState)
	receiveStateCh := make(chan netmodule.ElevatorState)

	go netmodule.CheckForPeerUpdates(id1, networkChs)

	go netmodule.SendStateUpdate(id1, networkChs, sendStateCh, 1, 10)

	go netmodule.ReceiveStateUpdates(id1, networkChs, receiveStateCh, true)

	go func() {
		for {
			sendStateCh <- netmodule.ElevatorState{"NaN", "This is a test"}
			fmt.Printf("Sent package from %s\n", id1)
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		select {
		case receivedState := <-receiveStateCh:
			fmt.Printf("Received new state for elevator %s\n", receivedState.ElevatorId)
			fmt.Println(receivedState.TestText)

		}
	}

}
