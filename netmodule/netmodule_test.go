package netmodule_test

import (
	"flag"
	"fmt"
	"strconv"
	"testing"
	"time"

	"../netmodule"
)

var idString = flag.String("id", "int", "Id of the elevator")

func TestNetworkModule(*testing.T) {

	id1, err := strconv.Atoi(*idString)
	if err != nil {
		id1 = 1
	}

	fmt.Println("Testing Network Module")

	ports := netmodule.NetworkPorts{16363, 16363, 26363, 26363}

	sendStateCh := make(chan [netmodule.ElevatorCount]netmodule.ElevatorState)
	receiveStateCh := make(chan [netmodule.ElevatorCount]netmodule.ElevatorState)

	fmt.Println("running module")

	disconnectCh := make(chan int)
	connectCh := make(chan int)
	netmodule.RunNetworkModule(id1, ports, sendStateCh, receiveStateCh, disconnectCh, connectCh)

	mockStates := [netmodule.ElevatorCount]netmodule.ElevatorState{
		{1, "test"},
		{2, "test"},
		{3, "test"},
	}

	go func() {
		for {
			sendStateCh <- mockStates
			fmt.Printf("Sent package from %d\n", id1)
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		select {
		case ID := <-disconnectCh:
			fmt.Printf("Elevator %d disconnected\n", ID)
		case ID := <-connectCh:
			fmt.Printf("Elevator %d connected\n", ID)
		case receivedState := <-receiveStateCh:

			fmt.Println(receivedState[0].TestText)
			fmt.Println(receivedState[1].TestText)
			fmt.Println(receivedState[2].TestText)
		}
	}

}
