package netmodule_test

import (
	"flag"
	"fmt"
	"strconv"
	"testing"
	"time"

	dt "../datatypes"
	"../netmodule"
)

var idString = flag.String("id", "int", "Id of the elevator")

func TestNetworkModule(*testing.T) {

	id1, err := strconv.Atoi(*idString)
	if err != nil {
		id1 = 1
	}

	fmt.Println("Testing Network Module")

	ports := netmodule.NetworkPorts{
		PeerTxPort:  16363,
		PeerRxPort:  16363,
		BcastRxPort: 26363,
		BcastTxPort: 26363,
	}

	outgoingStateCh := make(chan [dt.ElevatorCount]dt.ElevatorState)
	incomingStateCh := make(chan [dt.ElevatorCount]dt.ElevatorState)

	outgoingOrderCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)
	incomingOrderCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)

	fmt.Println("running module")

	disconnectCh := make(chan int)
	connectCh := make(chan int)
	netmodule.RunNetworkModule(id1, ports, outgoingStateCh, incomingStateCh, outgoingOrderCh, incomingOrderCh, disconnectCh, connectCh)

	var mockStates [dt.ElevatorCount]dt.ElevatorState
	var mockOrders [dt.ElevatorCount]dt.OrderMatrixType

	mockStates[1] = dt.ElevatorState{ElevatorID: 1, MovingDirection: dt.MovingDown, Floor: 1, State: 1, IsFunctioning: true}
	mockOrders[2][1][3] = dt.New
	go func() {
		for {
			outgoingStateCh <- mockStates
			outgoingOrderCh <- mockOrders
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
		case receivedState := <-incomingStateCh:

			fmt.Println(receivedState[0])
			fmt.Println(receivedState[1])
			fmt.Println(receivedState[2])

		case receivedOrder := <-incomingOrderCh:

			fmt.Println(receivedOrder[0])
			fmt.Println(receivedOrder[1])
			fmt.Println(receivedOrder[2])

		}
	}

}
