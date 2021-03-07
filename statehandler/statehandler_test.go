package statehandler_test

import (
	"flag"
	"fmt"
	"strconv"
	"testing"
	"time"

	dt "../datatypes"
	"../statehandler"
)

var idString = flag.String("id", "int", "Id of the elevator")

func TestNetworkModule(*testing.T) {

	id1, err := strconv.Atoi(*idString)
	if err != nil {
		id1 = 1
	}

	fmt.Println("Testing Statehandler Module")

	//Create channels
	outgoingStateCh := make(chan dt.ElevatorState)
	incomingStateCh := make(chan dt.ElevatorState)

	outgoingOrderCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)
	incomingOrderCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)

	stateUpdateCh := make(chan [dt.ElevatorCount]dt.ElevatorState)
	orderUpdateCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)

	driverStateUpdateCh := make(chan dt.ElevatorState)

	disconnectCh := make(chan int)

	acceptedOrderCh := make(chan dt.OrderType)

	completedOrderCh := make(chan dt.OrderType)

	go statehandler.RunStateHandlerModule(id1, incomingOrderCh, outgoingOrderCh,
		incomingStateCh, outgoingStateCh,
		disconnectCh,
		stateUpdateCh,
		orderUpdateCh,
		driverStateUpdateCh,
		acceptedOrderCh,
		completedOrderCh)

	fmt.Println("Module is running")

	var mockState dt.ElevatorState
	var mockOrders [dt.ElevatorCount]dt.OrderMatrixType
	mockOrders[0][0][0] = dt.New
	fmt.Println(mockOrders)
	go func() {
		for {

			mockState = dt.ElevatorState{ElevatorID: 2, MovingDirection: dt.MovingDown, Floor: 1, State: 1, IsFunctioning: true}
			driverState := dt.ElevatorState{ElevatorID: 1, MovingDirection: dt.MovingDown, Floor: 2, State: 4, IsFunctioning: true}

			incomingStateCh <- mockState
			fmt.Println("Sent state")
			time.Sleep(5 * time.Second)
			incomingOrderCh <- mockOrders
			fmt.Println("Sent order update")
			time.Sleep(5 * time.Second)
			driverStateUpdateCh <- driverState
			fmt.Println("Sent driver state")
			time.Sleep(5 * time.Second)
			completedOrderCh <- dt.OrderType{Button: dt.BtnHallUp, Floor: 0}
			fmt.Println("Sent completed order")
			time.Sleep(5 * time.Second)

		}
	}()
	fmt.Println("Sending mockup data")

	for {
		select {
		case ID := <-disconnectCh:
			fmt.Printf("Elevator %d disconnected\n", ID)
		case state := <-outgoingStateCh:
			fmt.Print("Outgoing state: ")
			fmt.Println(state)
		case orders := <-outgoingOrderCh:
			fmt.Print("Outgoing order: ")
			fmt.Println(orders)
		case order := <-acceptedOrderCh:
			fmt.Print("accepted order: ")
			fmt.Println(order)
		case stateUpdate := <-stateUpdateCh:
			fmt.Print("state update: ")
			fmt.Println(stateUpdate)
		case orderUpdate := <-orderUpdateCh:
			fmt.Print("order update: ")
			fmt.Println(orderUpdate)
		}
	}

}
