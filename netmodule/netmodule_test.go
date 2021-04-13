package netmodule_test

import (
	"flag"
	"fmt"
	"strconv"
	"testing"
	"time"

	cf "../config"
	dt "../datatypes"
	"../netmodule"
)

var idString = flag.String("id", "int", "Id of the elevator")

func TestNetworkModule(*testing.T) {

	id1, err := strconv.Atoi(*idString)
	if err != nil {
		id1 = 0
	}

	fmt.Println("Testing Network Module")

	ports := netmodule.NetworkPorts{
		PeerTxPort:  16363,
		PeerRxPort:  16363,
		BcastRxPort: 26363,
		BcastTxPort: 26363,
	}

	outgoingStateCh := make(chan dt.ElevatorState)
	incomingStateCh := make(chan dt.ElevatorState)

	outgoingOrderCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType)
	incomingOrderCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType)

	fmt.Println("running module")

	disconnectCh := make(chan int)
	connectCh := make(chan int)
	netmodule.RunNetworkModule(id1, ports, outgoingStateCh, incomingStateCh, outgoingOrderCh, incomingOrderCh, disconnectCh, connectCh)

	var mockState dt.ElevatorState
	var mockOrders [cf.ElevatorCount]dt.OrderMatrixType

	mockState = dt.ElevatorState{ElevatorID: 0, MovingDirection: dt.MovingDown, Floor: 1, State: dt.Idle, IsFunctioning: true}
	mockOrders[2][1][3] = dt.New
	go func() {
		for {
			outgoingStateCh <- mockState
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

			fmt.Println(receivedState)

		case receivedOrder := <-incomingOrderCh:

			fmt.Println(receivedOrder[0])
			fmt.Println(receivedOrder[1])
			fmt.Println(receivedOrder[2])

		}
	}

}
