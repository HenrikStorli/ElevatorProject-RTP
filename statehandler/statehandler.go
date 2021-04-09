package statehandler

import (
	//"fmt"
	dt "../datatypes"
)

type connectionState bool

const (
	Connected    connectionState = true
	Disconnected connectionState = false
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
	connectingElevatorIDCh <-chan int,

	//interface towards order scheduler
	stateUpdateCh chan<- [dt.ElevatorCount]dt.ElevatorState,
	orderUpdateCh chan<- [dt.ElevatorCount]dt.OrderMatrixType,
	newOrdersCh <-chan [dt.ElevatorCount]dt.OrderMatrixType,
	redirectedOrderCh chan<- dt.OrderType,

	//Interface towards elevator driver
	driverStateUpdateCh <-chan dt.ElevatorState,
	acceptedOrderCh chan<- dt.OrderType,
	completedOrderFloorCh <-chan int,
) {

	var orderMatrices [dt.ElevatorCount]dt.OrderMatrixType
	var elevatorStates [dt.ElevatorCount]dt.ElevatorState
	var connectedElevators [dt.ElevatorCount]connectionState

	var timeoutCh chan bool = make(chan bool)

	for {
		select {
		case newOrderMatrices := <-incomingOrderCh:

			updatedOrderMatrices := updateIncomingOrders(newOrderMatrices, orderMatrices)

			updatedOrderMatrices = ackNewOrders(elevatorID, updatedOrderMatrices, false)

			if updatedOrderMatrices != orderMatrices {
				updatedOrderMatrices = acceptAndSendOrders(elevatorID, updatedOrderMatrices, acceptedOrderCh)
				outgoingOrderCh <- updatedOrderMatrices
			}
			orderMatrices = updatedOrderMatrices

		case newOrders := <-newOrdersCh:

			updatedOrderMatrices := updateIncomingOrders(newOrders, orderMatrices)

			//If the elevator is single, skip the acknowlegdement step and accept new orders directly
			if isSingleElevator(elevatorID, connectedElevators) {
				updatedOrderMatrices = ackNewOrders(elevatorID, updatedOrderMatrices, true)

				acceptAndSendOrders(elevatorID, updatedOrderMatrices, acceptedOrderCh)

			}

			if updatedOrderMatrices != orderMatrices {
				outgoingOrderCh <- updatedOrderMatrices
			}

			orderMatrices = updatedOrderMatrices

		case newState := <-incomingStateCh:

			updatedStates := updateIncomingStates(elevatorID, newState, elevatorStates)

			go sendStateUpdate(updatedStates, stateUpdateCh)

			elevatorStates = updatedStates

		case newDriverStateUpdate := <-driverStateUpdateCh:

			updatedStates := updateOwnState(elevatorID, newDriverStateUpdate, elevatorStates)

			go sendOwnStateUpdate(newDriverStateUpdate, outgoingStateCh)

			go sendStateUpdate(updatedStates, stateUpdateCh)

			elevatorStates = updatedStates

		case completedOrderFloor := <-completedOrderFloorCh:

			updatedOrderMatrices := completeOrders(elevatorID, completedOrderFloor, orderMatrices)

			if updatedOrderMatrices != orderMatrices {
				outgoingOrderCh <- updatedOrderMatrices
			}

			orderMatrices = updatedOrderMatrices

		case connectingElevatorID := <-connectingElevatorIDCh:
			if !isConnected(connectingElevatorID, connectedElevators) {
				updatedConnectedElevators := updateConnectedElevatorList(connectingElevatorID, Connected, connectedElevators)

				indexID := elevatorID - 1
				ownState := elevatorStates[indexID]

				go sendOwnStateUpdate(ownState, outgoingStateCh)

				outgoingOrderCh <- orderMatrices

				connectedElevators = updatedConnectedElevators
			}

		case disconnectingElevatorID := <-disconnectingElevatorIDCh:
			if isConnected(disconnectingElevatorID, connectedElevators) {

				updatedConnectedElevators := updateConnectedElevatorList(disconnectingElevatorID, Disconnected, connectedElevators)
				updatedStates := elevatorStates
				//Remove existing hall calls
				updatedOrderMatrices := removeRedirectedOrders(disconnectingElevatorID, orderMatrices)
				orderUpdateCh <- orderMatrices

				if disconnectingElevatorID == elevatorID {
					//Business as usual
				} else {

					updatedStates = updateStateOfDisconnectingElevator(disconnectingElevatorID, elevatorStates)

					//Send state and orders to order scheduler before sending the redirected orders
					go sendStateUpdate(updatedStates, stateUpdateCh)

					//Sends redirected orders to orderscheduler after state and order update
					go redirectOrders(disconnectingElevatorID, orderMatrices, redirectedOrderCh)

				}

				if updatedOrderMatrices != orderMatrices {
					outgoingOrderCh <- updatedOrderMatrices
				}

				connectedElevators = updatedConnectedElevators
				orderMatrices = updatedOrderMatrices
				elevatorStates = updatedStates
			}

		case timeout := <-timeoutCh:
			if timeout {

			} else {
				// All good, pass
			}
		}
		orderUpdateCh <- orderMatrices
	}
}

func sendOrderUpdate(newOrders [dt.ElevatorCount]dt.OrderMatrixType, orderUpdateCh chan<- [dt.ElevatorCount]dt.OrderMatrixType, outgoingOrderCh chan<- [dt.ElevatorCount]dt.OrderMatrixType) {
	outgoingOrderCh <- newOrders
}

func sendStateUpdate(newStates [dt.ElevatorCount]dt.ElevatorState, stateUpdateCh chan<- [dt.ElevatorCount]dt.ElevatorState) {
	stateUpdateCh <- newStates
}

func sendOwnStateUpdate(state dt.ElevatorState, outgoingStateCh chan<- dt.ElevatorState) {
	outgoingStateCh <- state
}

func sendAcceptedOrders(elevatorID int, newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType, acceptedOrderCh chan<- dt.OrderType) {
	//TODO: add timeout timer for accepted orders
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

func updateIncomingStates(elevatorID int, newStateUpdate dt.ElevatorState, oldStates [dt.ElevatorCount]dt.ElevatorState) [dt.ElevatorCount]dt.ElevatorState {
	updatedStates := oldStates

	if newStateUpdate.ElevatorID == elevatorID {
		return updatedStates
	}

	for indexID := range updatedStates {
		id := indexID + 1
		if id == newStateUpdate.ElevatorID {
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

func updateStateOfDisconnectingElevator(disconnectingElevatorID int, oldStates [dt.ElevatorCount]dt.ElevatorState) [dt.ElevatorCount]dt.ElevatorState {
	updatedStates := oldStates
	indexID := disconnectingElevatorID - 1
	updatedStates[indexID].IsFunctioning = false
	return updatedStates
}

func updateConnectedElevatorList(elevatorID int, newConnectionState connectionState, connectedElevators [dt.ElevatorCount]connectionState) [dt.ElevatorCount]connectionState {
	updatedList := connectedElevators

	indexID := elevatorID - 1
	updatedList[indexID] = newConnectionState

	return updatedList
}
