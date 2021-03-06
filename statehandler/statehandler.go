package statehandler

import (
	dt "../datatypes"
)

func RunStateHandlerModule(elevatorID int,
	//Interface towards both the network module and order scheduler
	outgoingOrderCh chan<- [dt.ElevatorCount]dt.OrderMatrixType,
	incomingOrderCh <-chan [dt.ElevatorCount]dt.OrderMatrixType,
	outgoingStateCh chan<- [dt.ElevatorCount]dt.ElevatorState,

	//Interface towards network module
	incomingStateCh <-chan [dt.ElevatorCount]dt.ElevatorState,
	disconnectingElevatorIDCh <-chan int,

	//Interface towards elevator driver
	driverStateUpdateCh <-chan dt.ElevatorState,
	acceptedOrderCh chan<- dt.OrderType,
	confirmedOrderCh <-chan dt.OrderType,

) {

	var orderMatrices [dt.ElevatorCount]dt.OrderMatrixType

	var elevatorStates [dt.ElevatorCount]dt.ElevatorState

	for {
		select {
		case newOrderMatrices := <-incomingOrderCh:
			updatedOrderMatrices := acceptNewOrders(newOrders, orderMatrices)
			orderMatrices = updatedOrderMatrices
		case newStates := <-incomingStateCh:

		case newDriverStateUpdate := <-driverStateUpdateCh:

		case confirmedOrder := <-confirmedOrderCh:

		}
	}
}

func acceptNewOrders(newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType, oldOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType) {

}

func updateOrders() {

}

func updateStates() {

}

func updateOwnState() {

}
