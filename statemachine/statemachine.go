package statemachine

import (
	dt "../datatypes"
)

func RunStateMachine(elevatorID int,
	//Interface towards statehandler
	driverStateUpdateCh chan<- dt.ElevatorState,
	acceptedOrderCh <-chan dt.OrderType,
	completedOrderCh chan<- dt.OrderType) {

	//var currentState dt.MachineStateType
	go listenForAcceptedOrders(acceptedOrderCh)

	//run state machine here
}

func listenForAcceptedOrders(acceptedOrderCh <-chan dt.OrderType) {
	for {
		select {
		case <-acceptedOrderCh:

		}
	}
}
