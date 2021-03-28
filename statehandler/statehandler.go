package statehandler

import (

	//"fmt"
	dt "../datatypes"
	"../iomodule"
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
	//Interface towards iomodule
	buttonLampCh chan<- iomodule.ButtonLampType,
) {

	var orderMatrices [dt.ElevatorCount]dt.OrderMatrixType
	var elevatorStates [dt.ElevatorCount]dt.ElevatorState

	for {
		select {
		case newOrderMatrices := <-incomingOrderCh:

			updatedOrderMatrices := updateIncomingOrders(newOrderMatrices, orderMatrices)

			updatedOrderMatrices = replaceNewOrders(elevatorID, updatedOrderMatrices)
			//fmt.Printf("modified matrix %v \n", updatedOrderMatrices)

			go sendAcceptedOrders(elevatorID, updatedOrderMatrices, acceptedOrderCh)
			go sendOrderUpdate(updatedOrderMatrices, orderUpdateCh, outgoingOrderCh)
			go setButtonLamps(updatedOrderMatrices, buttonLampCh)

			orderMatrices = updatedOrderMatrices

		case newOrders := <-newOrdersCh:

			updatedOrderMatrices := updateIncomingOrders(newOrders, orderMatrices)

			go sendOrderUpdate(updatedOrderMatrices, orderUpdateCh, outgoingOrderCh)

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

			go sendOrderUpdate(updatedOrderMatrices, orderUpdateCh, outgoingOrderCh)

			orderMatrices = updatedOrderMatrices

		case <-connectingElevatorIDCh:

			indexID := elevatorID - 1
			ownState := elevatorStates[indexID]

			go sendOwnStateUpdate(ownState, outgoingStateCh)

			go sendOrderUpdate(orderMatrices, orderUpdateCh, outgoingOrderCh)

		case disconnectingElevatorID := <-disconnectingElevatorIDCh:

			updatedStates := updateStateOfDisconnectingElevator(disconnectingElevatorID, elevatorStates)

			updatedOrderMatrices := removeRedirectedOrders(disconnectingElevatorID, orderMatrices)

			//Send state and orders to order scheduler before sending the redirected orders
			go sendStateUpdate(updatedStates, stateUpdateCh)

			go sendOrderUpdate(updatedOrderMatrices, orderUpdateCh, outgoingOrderCh)

			//Sends redirected orders to orderscheduler after state and order update
			go redirectOrders(disconnectingElevatorID, orderMatrices, redirectedOrderCh)

			orderMatrices = updatedOrderMatrices
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

func setButtonLamps(newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType, buttonLampCh chan<- iomodule.ButtonLampType) {

	for rowIndex, row := range newOrderMatrices[0] {
		btn := dt.ButtonType(rowIndex)
		for floor, _ := range row {
			lampStatus := false
			order := dt.OrderType{Button: btn, Floor: floor}
			for _, orderMatrix := range newOrderMatrices {
				if orderMatrix[rowIndex][floor] == dt.Accepted {
					lampStatus = true
				}
			}
			buttonLampCh <- iomodule.ButtonLampType{Order: order, TurnOn: lampStatus}
		}
	}

}
