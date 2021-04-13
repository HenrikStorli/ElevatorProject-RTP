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

//RunStateHandlerModule is...
func RunStateHandlerModule(elevatorID int,
	//Interface towards both the network module and order scheduler
	incomingOrderCh <-chan [cf.ElevatorCount]dt.OrderMatrixType,

	//Interface towards network module
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
		case newOrderMatrices := <-incomingOrderCh:

			updatedOrderMatrices := updateIncomingOrders(newOrderMatrices, orderMatrices)

			updatedOrderMatrices = ackNewOrders(elevatorID, updatedOrderMatrices, false)

			if updatedOrderMatrices != orderMatrices {
				updatedOrderMatrices = acceptAndSendOrders(elevatorID, updatedOrderMatrices, acceptedOrderCh)
				outgoingOrderCh <- updatedOrderMatrices
			}
			orderMatrices = updatedOrderMatrices

		case newScheduledOrder := <-newScheduledOrderCh:

			updatedOrderMatrices := insertNewScheduledOrder(newScheduledOrder, orderMatrices)

			//If the elevator is single, skip the acknowlegdement step and accept new orders directly
			if isSingleElevator(elevatorID, connectedElevators) {
				updatedOrderMatrices = ackNewOrders(elevatorID, updatedOrderMatrices, true)

				updatedOrderMatrices = acceptAndSendOrders(elevatorID, updatedOrderMatrices, acceptedOrderCh)

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

				ownState := elevatorStates[elevatorID]

				go sendOwnStateUpdate(ownState, outgoingStateCh)

				outgoingOrderCh <- orderMatrices

				connectedElevators = updatedConnectedElevators

				fmt.Printf("Elevator %d connected \n", connectingElevatorID)
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
					go sendStateUpdate(updatedStates, stateUpdateCh)

					//Sends redirected orders to orderscheduler after state and order update
					go redirectOrders(disconnectingElevatorID, orderMatrices, redirectedOrderCh)

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

func sendOrderUpdate(newOrders [cf.ElevatorCount]dt.OrderMatrixType, orderUpdateCh chan<- [cf.ElevatorCount]dt.OrderMatrixType, outgoingOrderCh chan<- [cf.ElevatorCount]dt.OrderMatrixType) {
	outgoingOrderCh <- newOrders
}

func sendStateUpdate(newStates [cf.ElevatorCount]dt.ElevatorState, stateUpdateCh chan<- [cf.ElevatorCount]dt.ElevatorState) {
	stateUpdateCh <- newStates
}

func sendOwnStateUpdate(state dt.ElevatorState, outgoingStateCh chan<- dt.ElevatorState) {
	outgoingStateCh <- state
}

func sendAcceptedOrders(elevatorID int, newOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType, acceptedOrderCh chan<- dt.OrderType) {

	newOwnOrderMatrix := newOrderMatrices[elevatorID]

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
