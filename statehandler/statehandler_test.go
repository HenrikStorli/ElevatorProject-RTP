package statehandler_test

import (
	"flag"
	"fmt"
	"strconv"
	"testing"
	"time"

	cf "../config"
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

	outgoingOrderCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType)
	incomingOrderCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType)

	stateUpdateCh := make(chan [cf.ElevatorCount]dt.ElevatorState)
	orderUpdateCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType)

	driverStateUpdateCh := make(chan dt.ElevatorState)
	acceptedOrderCh := make(chan dt.OrderType)
	completedFloorOrderCh := make(chan int)

	disconnectCh := make(chan int)
	connectCh := make(chan int)
	scheduledOrdersCh := make(chan dt.OrderType)
	redirectedOrderCh := make(chan dt.OrderType)

	go statehandler.RunStateHandlerModule(id1, incomingOrderCh, outgoingOrderCh,
		incomingStateCh, outgoingStateCh,
		disconnectCh,
		connectCh,
		stateUpdateCh,
		orderUpdateCh,
		scheduledOrdersCh,
		redirectedOrderCh,
		driverStateUpdateCh,
		acceptedOrderCh,
		completedFloorOrderCh)

	fmt.Println("Module is running")

	var mockState dt.ElevatorState
	var mockOrders [cf.ElevatorCount]dt.OrderMatrixType
	mockOrders[0][0][0] = dt.NewOrder
	mockOrders[1][0][0] = dt.NewOrder
	mockOrders[1][1][0] = dt.NewOrder
	mockOrders[1][2][0] = dt.AcceptedOrder
	fmt.Println(mockOrders)
	go func() {
		for {

			mockState = dt.ElevatorState{ElevatorID: 0, MovingDirection: dt.MovingUp, Floor: 1, State: dt.InitState, IsFunctioning: true}
			driverState := dt.ElevatorState{ElevatorID: 1, MovingDirection: dt.MovingDown, Floor: 2, State: dt.DoorOpenState, IsFunctioning: true}

			incomingStateCh <- mockState
			fmt.Println("Sent state")
			time.Sleep(5 * time.Second)
			incomingOrderCh <- mockOrders
			fmt.Println("Sent order update")
			time.Sleep(5 * time.Second)
			driverStateUpdateCh <- driverState
			fmt.Println("Sent driver state")
			time.Sleep(5 * time.Second)
			completedFloorOrderCh <- 0
			fmt.Println("Sent completed order")
			time.Sleep(5 * time.Second)
			disconnectCh <- 2
			fmt.Println("Sent disconnect elevator 2")

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
		case redirectedOrder := <-redirectedOrderCh:
			fmt.Print("redirected order: ")
			fmt.Println(redirectedOrder)
		}

	}

}
