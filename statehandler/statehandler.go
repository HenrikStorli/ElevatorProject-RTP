package statehandler

import (
	dt "../datatypes"
)

//RunStateHandlerModule is...
func RunStateHandlerModule(elevatorID int,
	//Interface towards both the network module and order scheduler
	incomingOrderCh <-chan [dt.ElevatorCount]dt.OrderMatrixType,

	//Interface towards network module
	outgoingOrderCh chan<- [dt.ElevatorCount]dt.OrderMatrixType,
	incomingStateCh <-chan dt.ElevatorState,
	outgoingStateCh chan<- dt.ElevatorState,
	disconnectingElevatorIDCh <-chan int,

	//interface towards order scheduler
	stateUpdateCh chan<- [dt.ElevatorCount]dt.ElevatorState,
	orderUpdateCh chan<- [dt.ElevatorCount]dt.OrderMatrixType,

	//Interface towards elevator driver
	driverStateUpdateCh <-chan dt.ElevatorState,
	acceptedOrderCh chan<- dt.OrderType,
	completedOrderCh <-chan dt.OrderType,

) {

	var orderMatrices [dt.ElevatorCount]dt.OrderMatrixType
	var elevatorStates [dt.ElevatorCount]dt.ElevatorState

	for {
		select {
		case newOrderMatrices := <-incomingOrderCh:
			//TODO: add set lights
			updatedOrderMatrices := updateOrders(newOrderMatrices, orderMatrices)

			acceptedOrderMatrices := acceptNewOrders(elevatorID, updatedOrderMatrices)

			go sendAcceptedOrders(elevatorID, acceptedOrderMatrices, acceptedOrderCh)
			go sendOrderUpdate(acceptedOrderMatrices, orderUpdateCh, outgoingOrderCh)

			orderMatrices = acceptedOrderMatrices

		case newState := <-incomingStateCh:
			updatedStates := updateStates(elevatorID, newState, elevatorStates)

			go sendStateUpdate(updatedStates, stateUpdateCh)

			elevatorStates = updatedStates
		case newDriverStateUpdate := <-driverStateUpdateCh:
			updatedStates := updateOwnState(elevatorID, newDriverStateUpdate, elevatorStates)

			go sendOwnStateUpdate(newDriverStateUpdate, outgoingStateCh)
			go sendStateUpdate(updatedStates, stateUpdateCh)

			elevatorStates = updatedStates

		case completedOrder := <-completedOrderCh:
			updatedOrderMatrices := updateCompletedOrder(elevatorID, completedOrder, orderMatrices)

			go sendOrderUpdate(updatedOrderMatrices, orderUpdateCh, outgoingOrderCh)

			orderMatrices = updatedOrderMatrices

		case disconnectingElevatorID := <-disconnectingElevatorIDCh:
			updatedStates := handleDisconnectingElevator(disconnectingElevatorID, elevatorStates)
			elevatorStates = updatedStates
		}
	}
}

func sendOrderUpdate(newOrders [dt.ElevatorCount]dt.OrderMatrixType, orderUpdateCh chan<- [dt.ElevatorCount]dt.OrderMatrixType, outgoingOrderCh chan<- [dt.ElevatorCount]dt.OrderMatrixType) {
	go func() { orderUpdateCh <- newOrders }()
	go func() { outgoingOrderCh <- newOrders }()
}

func sendStateUpdate(newStates [dt.ElevatorCount]dt.ElevatorState, stateUpdateCh chan<- [dt.ElevatorCount]dt.ElevatorState) {
	stateUpdateCh <- newStates
}

func sendOwnStateUpdate(state dt.ElevatorState, outgoingStateCh chan<- dt.ElevatorState) {
	outgoingStateCh <- state
}

func sendAcceptedOrders(elevatorID int, newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType, acceptedOrderCh chan<- dt.OrderType) {
	indexID := elevatorID - 1
	newOwnOrderMatrix := newOrderMatrices[indexID]
	for rowIndex, row := range newOwnOrderMatrix {
		btn := dt.ButtonType(rowIndex)
		for floor, newOrder := range row {
			if newOrder == dt.Accepted {
				acceptedOrder := dt.OrderType{Button: btn, Floor: floor}
				acceptedOrderCh <- acceptedOrder
			}
		}
	}
}

func acceptNewOrders(elevatorID int, newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType) [dt.ElevatorCount]dt.OrderMatrixType {
	indexID := elevatorID - 1
	updatedOrderMatrices := newOrderMatrices
	newOwnOrderMatrix := newOrderMatrices[indexID]
	for btn, row := range newOwnOrderMatrix {
		for floor, newOrder := range row {
			if newOrder == dt.New {
				updatedOrderMatrices[indexID][btn][floor] = dt.Accepted
			}
		}

	}
	return updatedOrderMatrices
}

func updateOrders(newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType, oldOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType) [dt.ElevatorCount]dt.OrderMatrixType {

	updatedOrderMatrices := oldOrderMatrices
	for indexID, orderMatrix := range newOrderMatrices {
		for btn, row := range orderMatrix {
			for floor, newOrder := range row {
				oldOrder := &updatedOrderMatrices[indexID][btn][floor]
				*oldOrder = updateSingleOrder(newOrder, *oldOrder)
			}
		}
	}
	return updatedOrderMatrices
}

func updateSingleOrder(newOrder dt.OrderStateType, oldOrder dt.OrderStateType) dt.OrderStateType {

	updatedOrder := oldOrder
	switch oldOrder {
	case dt.Unknown:
		updatedOrder = newOrder
	case dt.New:
		if newOrder == dt.Accepted {
			updatedOrder = newOrder
		}
	case dt.Accepted:
		if newOrder == dt.Completed {
			updatedOrder = newOrder
		}
	case dt.Completed:
		if newOrder == dt.New {
			updatedOrder = newOrder
		}
	}

	return updatedOrder
}

func updateCompletedOrder(elevatorID int, completedOrder dt.OrderType, oldOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType) [dt.ElevatorCount]dt.OrderMatrixType {
	indexID := elevatorID - 1
	updatedOrderMatrices := oldOrderMatrices
	floor := completedOrder.Floor
	btn := completedOrder.Button

	oldOrder := &updatedOrderMatrices[indexID][btn][floor]
	*oldOrder = updateSingleOrder(dt.Completed, *oldOrder)

	return updatedOrderMatrices
}

func updateStates(elevatorID int, newStateUpdate dt.ElevatorState, oldStates [dt.ElevatorCount]dt.ElevatorState) [dt.ElevatorCount]dt.ElevatorState {
	updatedStates := oldStates

	if newStateUpdate.ElevatorID == elevatorID {
		return updatedStates
	}

	for id := range updatedStates {
		if id == newStateUpdate.ElevatorID {
			indexID := id - 1
			updatedStates[indexID] = newStateUpdate
		}
	}

	return updatedStates
}

func updateOwnState(elevatorID int, newState dt.ElevatorState, oldStates [dt.ElevatorCount]dt.ElevatorState) [dt.ElevatorCount]dt.ElevatorState {
	indexID := elevatorID - 1
	updatedStates := oldStates
	updatedStates[indexID] = newState

	return updatedStates
}

func handleDisconnectingElevator(disconnectingElevatorID int, oldStates [dt.ElevatorCount]dt.ElevatorState) [dt.ElevatorCount]dt.ElevatorState {
	updatedStates := oldStates
	indexID := disconnectingElevatorID - 1
	updatedStates[indexID].IsFunctioning = false
	return oldStates
}
