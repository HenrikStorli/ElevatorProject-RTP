package statehandler

import (
	//"fmt"
	"fmt"

	cf "../config"
	dt "../datatypes"
)

type connectionState bool

const (
	Connected    connectionState = true
	Disconnected connectionState = false
)

func RunStateHandlerModule(elevatorID int,

	//Interface towards network module
	incomingOrderCh <-chan [cf.ElevatorCount]dt.OrderMatrixType,
	outgoingOrderCh chan<- [cf.ElevatorCount]dt.OrderMatrixType,
	incomingStateCh <-chan dt.ElevatorState,
	outgoingStateCh chan<- dt.ElevatorState,
	disconnectingElevatorIDCh <-chan int,
	connectingElevatorIDCh <-chan int,

	//interface towards order scheduler
	stateUpdateCh chan<- [cf.ElevatorCount]dt.ElevatorState,
	orderUpdateCh chan<- [cf.ElevatorCount]dt.OrderMatrixType,
	newScheduledOrderCh <-chan dt.OrderType,
	redirectedOrderCh chan<- dt.OrderType,

	//Interface towards elevator driver
	driverStateUpdateCh <-chan dt.ElevatorState,
	acceptedOrderCh chan<- dt.OrderType,
	completedOrderFloorCh <-chan int,
) {

	var orderMatrices [cf.ElevatorCount]dt.OrderMatrixType
	var elevatorStates [cf.ElevatorCount]dt.ElevatorState
	var connectedElevators [cf.ElevatorCount]connectionState

	for {
		select {

		// NewOrder order update coming from the other elevators
		case newOrderUpdate := <-incomingOrderCh:

			updatedOrderMatrices := updateIncomingOrders(newOrderUpdate, orderMatrices)

			updatedOrderMatrices = setNewOrdersToAck(elevatorID, updatedOrderMatrices, false)
			updatedOrderMatrices = setCompletedOrdersToNone(elevatorID, updatedOrderMatrices, false)

			if updatedOrderMatrices != orderMatrices {
				updatedOrderMatrices = acceptAndSendOrders(elevatorID, updatedOrderMatrices, acceptedOrderCh)
				outgoingOrderCh <- updatedOrderMatrices
			}
			orderMatrices = updatedOrderMatrices

		// NewOrder state update coming from the other elevators
		case newStateUpdate := <-incomingStateCh:

			updatedStates := updateIncomingStates(elevatorID, newStateUpdate, elevatorStates)

			stateUpdateCh <- updatedStates

			elevatorStates = updatedStates

		// Order scheduler has made a new scheduled order
		case newScheduledOrder := <-newScheduledOrderCh:

			updatedOrderMatrices := insertNewScheduledOrder(newScheduledOrder, orderMatrices)

			//If the elevator is single, skip the acknowlegdement step and accept new orders directly
			if isSingleElevator(elevatorID, connectedElevators) {
				updatedOrderMatrices = setNewOrdersToAck(elevatorID, updatedOrderMatrices, true)

				updatedOrderMatrices = acceptAndSendOrders(elevatorID, updatedOrderMatrices, acceptedOrderCh)
			}

			if updatedOrderMatrices != orderMatrices {
				outgoingOrderCh <- updatedOrderMatrices
			}

			orderMatrices = updatedOrderMatrices

		// NewOrder state coming from Elevator Driver
		case newDriverStateUpdate := <-driverStateUpdateCh:

			updatedStates := updateOwnState(elevatorID, newDriverStateUpdate, elevatorStates)

			outgoingStateCh <- newDriverStateUpdate

			stateUpdateCh <- updatedStates

			elevatorStates = updatedStates

		// Elevator drivers has completed orders on this floor
		case completedOrderFloor := <-completedOrderFloorCh:

			updatedOrderMatrices := completeOrders(elevatorID, completedOrderFloor, orderMatrices)

			if updatedOrderMatrices != orderMatrices {
				outgoingOrderCh <- updatedOrderMatrices
			}

			// Skip the complete -> none step when single elevator
			if isSingleElevator(elevatorID, connectedElevators) {
				updatedOrderMatrices = setCompletedOrdersToNone(elevatorID, updatedOrderMatrices, true)
				outgoingOrderCh <- updatedOrderMatrices
			}

			orderMatrices = updatedOrderMatrices

		case connectingElevatorID := <-connectingElevatorIDCh:
			if !isConnected(connectingElevatorID, connectedElevators) {

				fmt.Printf("Elevator %d connected \n", connectingElevatorID)

				updatedConnectedElevators := updateConnectedElevatorList(connectingElevatorID, Connected, connectedElevators)

				// Send own state and orders to all the elevators
				ownState := elevatorStates[elevatorID]

				outgoingStateCh <- ownState

				outgoingOrderCh <- orderMatrices

				connectedElevators = updatedConnectedElevators
			}

		case disconnectingElevatorID := <-disconnectingElevatorIDCh:
			if isConnected(disconnectingElevatorID, connectedElevators) {

				updatedConnectedElevators := updateConnectedElevatorList(disconnectingElevatorID, Disconnected, connectedElevators)
				updatedStates := elevatorStates
				updatedOrderMatrices := orderMatrices

				if disconnectingElevatorID == elevatorID {
					//Business as usual
				} else {

					//Remove existing hall calls
					updatedOrderMatrices = removeRedirectedOrders(disconnectingElevatorID, orderMatrices)
					//Send order update to order scheduler before redirecting orders
					orderUpdateCh <- updatedOrderMatrices

					//Send updated state to order scheduler before sending the redirected orders
					updatedStates = updateStateOfDisconnectingElevator(disconnectingElevatorID, elevatorStates)
					stateUpdateCh <- updatedStates

					//Sends redirected orders to orderscheduler after state and order update
					redirectOrders(disconnectingElevatorID, orderMatrices, redirectedOrderCh)

					fmt.Printf("Elevator %d disconnected \n", disconnectingElevatorID)
				}

				if updatedOrderMatrices != orderMatrices {
					outgoingOrderCh <- updatedOrderMatrices
				}

				connectedElevators = updatedConnectedElevators
				orderMatrices = updatedOrderMatrices
				elevatorStates = updatedStates
			}

		}
		orderUpdateCh <- orderMatrices
	}
}

func sendAcceptedOrders(elevatorID int, newOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType, acceptedOrderCh chan<- dt.OrderType) {

	newOwnOrderMatrix := newOrderMatrices[elevatorID]

	for btnIndex, row := range newOwnOrderMatrix {
		btn := dt.ButtonType(btnIndex)
		for floor, newOrder := range row {
			if newOrder == dt.AcceptedOrder {
				acceptedOrder := dt.OrderType{Button: btn, Floor: floor}

				acceptedOrderCh <- acceptedOrder
			}
		}
	}
}

func updateIncomingStates(elevatorID int, newStateUpdate dt.ElevatorState, oldStates [cf.ElevatorCount]dt.ElevatorState) [cf.ElevatorCount]dt.ElevatorState {
	updatedStates := oldStates

	if newStateUpdate.ElevatorID == elevatorID {
		return updatedStates
	}

	for ID := range updatedStates {
		if ID == newStateUpdate.ElevatorID {
			updatedStates[ID] = newStateUpdate
		}
	}

	return updatedStates
}

func updateOwnState(elevatorID int, newState dt.ElevatorState, oldStates [cf.ElevatorCount]dt.ElevatorState) [cf.ElevatorCount]dt.ElevatorState {

	updatedStates := oldStates
	updatedStates[elevatorID] = newState

	return updatedStates
}

func updateStateOfDisconnectingElevator(disconnectingElevatorID int, oldStates [cf.ElevatorCount]dt.ElevatorState) [cf.ElevatorCount]dt.ElevatorState {
	updatedStates := oldStates
	updatedStates[disconnectingElevatorID].IsFunctioning = false
	return updatedStates
}

func updateConnectedElevatorList(elevatorID int, newConnectionState connectionState, connectedElevators [cf.ElevatorCount]connectionState) [cf.ElevatorCount]connectionState {
	updatedList := connectedElevators

	updatedList[elevatorID] = newConnectionState

	return updatedList
}
