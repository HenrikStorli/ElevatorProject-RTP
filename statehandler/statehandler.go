package statehandler

import (
	dt "../datatypes"
)

//RunStateHandlerModule is...
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
	completedOrderCh <-chan dt.OrderType,

) {

	var orderMatrices [dt.ElevatorCount]dt.OrderMatrixType
	var elevatorStates [dt.ElevatorCount]dt.ElevatorState

	for {
		select {
		case newOrderMatrices := <-incomingOrderCh:
			//TODO: add set lights
			updatedOrderMatrices := updateOrders(elevatorID, newOrderMatrices, orderMatrices)
			orderMatrices = updatedOrderMatrices

			acceptedOrderMatrices := acceptNewOrders(elevatorID, orderMatrices)
			orderMatrices = acceptedOrderMatrices

			go sendAcceptedOrders(elevatorID, acceptedOrderCh, orderMatrices)

		case newStates := <-incomingStateCh:
			updatedStates := updateStates(elevatorID, newStates, elevatorStates)
			elevatorStates = updatedStates

		case newDriverStateUpdate := <-driverStateUpdateCh:
			updatedStates := updateOwnState(elevatorID, newDriverStateUpdate, elevatorStates)
			elevatorStates = updatedStates

		case completedOrder := <-completedOrderCh:
			updatedOrderMatrices := updateCompletedOrder(elevatorID,completedOrder, orderMatrices)
			orderMatrices = updatedOrderMatrices

		case disconnectingElevatorID := <-disconnectingElevatorIDCh:
			updatedStates := handleDisconnectingElevator(disconnectingElevatorID,elevatorStates)
			elevatorStates = updatedStates

		}
	}
}

func acceptNewOrders(elevatorID int, newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType) [dt.ElevatorCount]dt.OrderMatrixType {
	updatedOrderMatrices := newOrderMatrices
	newOwnOrderMatrix := newOrderMatrices[elevatorID]
	for btn, row := range newOwnOrderMatrix {
		for floor, newOrder := range row {
			if newOrder == dt.New {
				updatedOrderMatrices[btn][floor] = dt.Accepted
			}
		}
	}
	return updatedOrderMatrices
}

func sendAcceptedOrders(elevatorID int, acceptedOrderCh chan<- dt.OrderType, newOrderMatrix dt.OrderMatrixType) {
	newOwnOrderMatrix := newOrderMatrices[elevatorID]
	for btn, row := range newOwnOrderMatrix {
		for floor, newOrder := range row {
			if newOrder == dt.Accepted {
				acceptedOrder := dt.OrderType{btn, floor}
				acceptedOrderCh <- acceptedOrder
			}
		}
	}
}

func updateOrders(elevatorID int, newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType, oldOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType) [dt.ElevatorCount]dt.OrderMatrixType {
	updatedOrderMatrices := oldOrderMatrices
	for elevator, orderMatrix := range newOrderMatrices {
		for btn, row := range orderMatrix {
			for floor, newOrder := range row {
				oldOrder := updatedOrderMatrices[elevatorID][btn][floor]
				oldOrder = updateSingleOrder(newOrder, oldOrder)
			}
		}
	}
	return updatedOrderMatrices
}

func updateSingleOrder(newOrder dt.OrderStateType, oldOrder dt.OrderStateType) dt.OrderStateType {

	updatedOrder := oldOrder
	select {
	case oldOrder == dt.Unknown:
		updatedOrder = newOrder
	case oldOrder == dt.New:
		if newOrder == dt.Accepted {
			updatedOrder = newOrder
		}
	case oldOrder == dt.Accepted:
		if newOrder == dt.Completed {
			updatedOrder = newOrder
		}
	case oldOrder == dt.Completed:
		if newOrder == dt.New {
			updateOrder = newOrder
		}
	}

	return updatedOrder
}

func updateCompletedOrder(elevatorID, completedOrder dt.OrderType, oldOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType) [dt.ElevatorCount]dt.OrderMatrixType {
	updatedOrderMatrices := oldOrderMatrices
	floor := completedOrder.Floor
	btn := completedOrder.Button
	
	oldOrder := updatedOrderMatrices[elevatorID][btn][floor] 
	oldOrder = updateSingleOrder(dt.Completed, oldOrder)

	return updatedOrderMatrices
}

func updateStates(elevatorID int, newStateUpdates [dt.ElevatorCount]ElevatorState, oldStates [dt.ElevatorCount]ElevatorState) [dt.ElevatorCount]ElevatorState {
	updatedStates := oldStates
	for id, state := range newStateUpdates {
		if id != elevatorID {
			updatedStates[id] = state
		}
	}
	return updatedStates
}

func updateOwnState(elevatorID int, newState ElevatorState, oldStates [dt.ElevatorCount]ElevatorState) [dt.ElevatorCount]ElevatorState {
	updatedStates := oldStates
	updatedStates[elevatorID] = newState

	return updatedStates
}

func handleDisconnectingElevator(disconnectingElevatorID int, oldStates [dt.ElevatorCount]ElevatorState) [dt.ElevatorCount]ElevatorState){
	updatedStates := oldStates
	updatedStates[disconnectingElevatorID].IsFunctioning = false
	return oldStates
}