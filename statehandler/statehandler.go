package statehandler

import (
	"fmt"
	"time"

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
			updatedOrderMatrices := updateOrders(newOrderMatrices, orderMatrices)

			updatedOrderMatrices = replaceNewOrders(elevatorID, updatedOrderMatrices)
			fmt.Printf("modified matrix %v \n", updatedOrderMatrices)

			go sendAcceptedOrders(elevatorID, updatedOrderMatrices, acceptedOrderCh)
			go sendOrderUpdate(updatedOrderMatrices, orderUpdateCh, outgoingOrderCh)
			go setButtonLamps(updatedOrderMatrices, buttonLampCh)

			orderMatrices = updatedOrderMatrices

		case newOrders := <-newOrdersCh:
			updatedOrderMatrices := updateOrders(newOrders, orderMatrices)

			go sendOrderUpdate(updatedOrderMatrices, orderUpdateCh, outgoingOrderCh)
			fmt.Printf("new order %v \n", newOrders)
			orderMatrices = updatedOrderMatrices

		case newState := <-incomingStateCh:
			updatedStates := updateStates(elevatorID, newState, elevatorStates)

			go sendStateUpdate(updatedStates, stateUpdateCh)

			elevatorStates = updatedStates
		case newDriverStateUpdate := <-driverStateUpdateCh:
			updatedStates := updateOwnState(elevatorID, newDriverStateUpdate, elevatorStates)

			go sendOwnStateUpdate(newDriverStateUpdate, outgoingStateCh)
			go sendStateUpdate(updatedStates, stateUpdateCh)

			elevatorStates = updatedStates

		case completedOrderFloor := <-completedOrderFloorCh:
			updatedOrderMatrices := updateCompletedOrder(elevatorID, completedOrderFloor, orderMatrices)
			fmt.Printf("Completed orders at floor %v \n", completedOrderFloor)
			go sendOrderUpdate(updatedOrderMatrices, orderUpdateCh, outgoingOrderCh)
			fmt.Printf("compl %v \n", updatedOrderMatrices)
			orderMatrices = updatedOrderMatrices

		case disconnectingElevatorID := <-disconnectingElevatorIDCh:
			updatedStates := handleDisconnectingElevator(disconnectingElevatorID, elevatorStates)
			go sendStateUpdate(updatedStates, stateUpdateCh)

			go redirectOrders(disconnectingElevatorID, orderMatrices, redirectedOrderCh)

			updatedOrderMatrices := removeRedirectedOrders(disconnectingElevatorID, orderMatrices)
			go sendOrderUpdate(updatedOrderMatrices, orderUpdateCh, outgoingOrderCh)

			orderMatrices = updatedOrderMatrices
			fmt.Printf("disc %v \n", orderMatrices)
			elevatorStates = updatedStates
		}
	}
}

func sendOrderUpdate(newOrders [dt.ElevatorCount]dt.OrderMatrixType, orderUpdateCh chan<- [dt.ElevatorCount]dt.OrderMatrixType, outgoingOrderCh chan<- [dt.ElevatorCount]dt.OrderMatrixType) {
	orderUpdateCh <- newOrders
	outgoingOrderCh <- newOrders
	//go func() { orderUpdateCh <- newOrders }()
	//go func() { outgoingOrderCh <- newOrders }()
}

func sendStateUpdate(newStates [dt.ElevatorCount]dt.ElevatorState, stateUpdateCh chan<- [dt.ElevatorCount]dt.ElevatorState) {
	stateUpdateCh <- newStates
}

func sendOwnStateUpdate(state dt.ElevatorState, outgoingStateCh chan<- dt.ElevatorState) {
	outgoingStateCh <- state
}

func sendAcceptedOrders(elevatorID int, newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType,
	acceptedOrderCh chan<- dt.OrderType) {
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

func replaceNewOrders(elevatorID int, newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType) [dt.ElevatorCount]dt.OrderMatrixType {
	indexID := elevatorID - 1
	updatedOrderMatrices := newOrderMatrices
	for ID := range newOrderMatrices {
		if ID != indexID {
			updatedOrderMatrices[ID] = replaceExistingOrders(dt.New, dt.Acknowledged, updatedOrderMatrices[ID])
			updatedOrderMatrices[ID] = replaceExistingOrders(dt.Completed, dt.None, updatedOrderMatrices[ID])
		}
		if ID == indexID {
			updatedOrderMatrices[ID] = replaceExistingOrders(dt.Acknowledged, dt.Accepted, updatedOrderMatrices[ID])
		}
	}
	return updatedOrderMatrices
}

func replaceExistingOrders(existingOrderType dt.OrderStateType,
	newOrderType dt.OrderStateType, newOrderMatrix dt.OrderMatrixType) dt.OrderMatrixType {

	updatedOrderMatrix := newOrderMatrix
	for btn, row := range newOrderMatrix {
		for floor, newOrder := range row {
			if newOrder == existingOrderType {
				updatedOrderMatrix[btn][floor] = newOrderType
			}
		}
	}
	return updatedOrderMatrix

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

//Updates a single order based on the order update rules
func updateSingleOrder(newOrder dt.OrderStateType, oldOrder dt.OrderStateType) dt.OrderStateType {

	updatedOrder := oldOrder
	switch oldOrder {
	case dt.Unknown:
		updatedOrder = newOrder
	case dt.None:
		if newOrder == dt.New {
			updatedOrder = newOrder
		}
	case dt.New:
		if newOrder == dt.Acknowledged {
			updatedOrder = newOrder
		}
	case dt.Acknowledged:
		if newOrder == dt.Accepted || newOrder == dt.Completed {
			updatedOrder = newOrder
		}
	case dt.Accepted:
		if newOrder == dt.Completed {
			updatedOrder = newOrder
		}
	case dt.Completed:
		if newOrder == dt.None {
			updatedOrder = newOrder
		}
	}
	return updatedOrder
}

func updateCompletedOrder(elevatorID int, completedOrderFloor int, oldOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType) [dt.ElevatorCount]dt.OrderMatrixType {

	indexID := elevatorID - 1
	updatedOrderMatrices := oldOrderMatrices
	floor := completedOrderFloor

	for rowIndex := range oldOrderMatrices {
		oldOrder := oldOrderMatrices[indexID][rowIndex][floor]
		if oldOrder == dt.Accepted {
			updatedOrderMatrices[indexID][rowIndex][floor] = dt.Completed
		}
	}

	return updatedOrderMatrices
}

func updateStates(elevatorID int, newStateUpdate dt.ElevatorState, oldStates [dt.ElevatorCount]dt.ElevatorState) [dt.ElevatorCount]dt.ElevatorState {
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

func handleDisconnectingElevator(disconnectingElevatorID int, oldStates [dt.ElevatorCount]dt.ElevatorState) [dt.ElevatorCount]dt.ElevatorState {
	updatedStates := oldStates
	indexID := disconnectingElevatorID - 1
	updatedStates[indexID].IsFunctioning = false
	return oldStates
}

//Sends hall calls of the disconnecting elevator to the order scheduler
func redirectOrders(disconnectingElevatorID int, oldOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType, redirectedOrderCh chan<- dt.OrderType) {
	//Wait to make sure the state of the disconnected elevator has reached order scheduler
	time.Sleep(time.Millisecond * 10)
	indexID := disconnectingElevatorID - 1
	ownOrderMatrix := oldOrderMatrices[indexID]

	for rowIndex, row := range ownOrderMatrix {
		btn := dt.ButtonType(rowIndex)
		//We dont redistribute cab calls
		if btn == dt.BtnCab {
			continue
		}
		for floor, orderState := range row {
			//Dont redistribute non-existing orders
			if isOrderActive(orderState) {
				order := dt.OrderType{Button: btn, Floor: floor}
				redirectedOrderCh <- order
				//Wait a tiny bit to avoiding locking the order scheduler
				//time.Sleep(time.Millisecond * 1)
			}
		}
	}

}

func removeRedirectedOrders(disconnectingElevatorID int, oldOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType) [dt.ElevatorCount]dt.OrderMatrixType {
	updatedOrderMatrices := oldOrderMatrices
	indexID := disconnectingElevatorID - 1
	ownOrderMatrix := oldOrderMatrices[indexID]

	for rowIndex, row := range ownOrderMatrix {
		btn := dt.ButtonType(rowIndex)
		for floor, orderState := range row {
			newOrderState := dt.None
			//We dont remove cab calls, but set them as Acknowledged
			//So that when the elevator reconnects it will execute the orders if it restarted
			if btn == dt.BtnCab && isOrderActive(orderState) {
				newOrderState = dt.Acknowledged
			}
			oldOrder := &updatedOrderMatrices[indexID][rowIndex][floor]
			*oldOrder = newOrderState
		}
	}
	return updatedOrderMatrices
}

func isOrderActive(orderState dt.OrderStateType) bool {
	if orderState == dt.None || orderState == dt.Completed || orderState == dt.Unknown {
		return false
	}
	return true
}
