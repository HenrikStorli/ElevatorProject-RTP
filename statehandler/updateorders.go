package statehandler

import (
	"time"
	//"fmt"
	cf "../config"
	dt "../datatypes"
)

func setCompletedOrdersToNone(elevatorID int, newOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType, singleElevator bool) [cf.ElevatorCount]dt.OrderMatrixType {

	updatedOrderMatrices := newOrderMatrices

	for indexID := range newOrderMatrices {
		if indexID != elevatorID || singleElevator {
			updatedOrderMatrices[indexID] = replaceExistingOrders(dt.CompletedOrder, dt.NoOrder, updatedOrderMatrices[indexID])
		}
	}
	return updatedOrderMatrices
}

func setNewOrdersToAck(elevatorID int, newOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType, singleElevator bool) [cf.ElevatorCount]dt.OrderMatrixType {

	updatedOrderMatrices := newOrderMatrices

	for indexID := range newOrderMatrices {
		if indexID != elevatorID || singleElevator {
			updatedOrderMatrices[indexID] = replaceExistingOrders(dt.NewOrder, dt.AckedOrder, updatedOrderMatrices[indexID])
		}
	}
	return updatedOrderMatrices
}

func acceptAndSendOrders(elevatorID int, newOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType, acceptedOrderCh chan<- dt.OrderType) [cf.ElevatorCount]dt.OrderMatrixType {

	updatedOrderMatrices := newOrderMatrices

	for btnIndex, row := range newOrderMatrices[elevatorID] {
		btn := dt.ButtonType(btnIndex)
		for floor, newOrder := range row {
			if newOrder == dt.AckedOrder {
				updatedOrderMatrices[elevatorID][btnIndex][floor] = dt.AcceptedOrder

				acceptedOrder := dt.OrderType{Button: btn, Floor: floor}
				acceptedOrderCh <- acceptedOrder
			}
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

func updateIncomingOrders(newOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType, oldOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType) [cf.ElevatorCount]dt.OrderMatrixType {

	updatedOrderMatrices := oldOrderMatrices
	for ID, orderMatrix := range newOrderMatrices {
		for btn, row := range orderMatrix {
			for floor, newOrder := range row {
				oldOrder := &updatedOrderMatrices[ID][btn][floor]
				*oldOrder = updateSingleOrderState(newOrder, *oldOrder)
			}
		}
	}
	return updatedOrderMatrices
}

func insertNewScheduledOrder(newScheduledOrder dt.OrderType, oldOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType) [cf.ElevatorCount]dt.OrderMatrixType {

	updatedOrderMatrices := oldOrderMatrices
	btnIndex := int(newScheduledOrder.Button)
	floor := newScheduledOrder.Floor
	elevatorID := newScheduledOrder.ElevatorID

	oldOrder := &updatedOrderMatrices[elevatorID][btnIndex][floor]
	*oldOrder = updateSingleOrderState(dt.NewOrder, *oldOrder)

	return updatedOrderMatrices
}

//Updates a single order based on the order update rules
func updateSingleOrderState(newOrder dt.OrderStateType, oldOrder dt.OrderStateType) dt.OrderStateType {

	updatedOrder := oldOrder
	switch oldOrder {
	case dt.UnknownOrder:
		updatedOrder = newOrder
	case dt.NoOrder:
		if newOrder == dt.NewOrder {
			updatedOrder = newOrder
		}
	case dt.NewOrder:
		if newOrder == dt.AckedOrder {
			updatedOrder = newOrder
		}
	case dt.AckedOrder:
		if newOrder == dt.AcceptedOrder || newOrder == dt.CompletedOrder {
			updatedOrder = newOrder
		}
	case dt.AcceptedOrder:
		if newOrder == dt.CompletedOrder {
			updatedOrder = newOrder
		}
	case dt.CompletedOrder:
		if newOrder == dt.NoOrder {
			updatedOrder = newOrder
		}
	}
	return updatedOrder
}

func completeOrders(elevatorID int, completedOrderFloor int, oldOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType) [cf.ElevatorCount]dt.OrderMatrixType {

	updatedOrderMatrices := oldOrderMatrices
	floor := completedOrderFloor

	for btnIndex := range oldOrderMatrices {
		oldOrder := oldOrderMatrices[elevatorID][btnIndex][floor]
		if oldOrder == dt.AcceptedOrder {
			updatedOrderMatrices[elevatorID][btnIndex][floor] = dt.CompletedOrder
		}
	}

	return updatedOrderMatrices
}

//Sends hall calls of the disconnecting elevator to the order scheduler
func redirectOrders(disconnectingElevatorID int, oldOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType, redirectedOrderCh chan<- dt.OrderType) {
	//Wait to make sure the state of the disconnected elevator has reached order scheduler
	time.Sleep(time.Millisecond * 10)

	ownOrderMatrix := oldOrderMatrices[disconnectingElevatorID]

	for btnIndex, row := range ownOrderMatrix {
		btn := dt.ButtonType(btnIndex)
		//We dont redistribute cab calls
		if btn == dt.ButtonCab {
			continue
		}
		for floor, orderState := range row {
			//Dont redistribute non-existing orders
			if isOrderActive(orderState) {
				order := dt.OrderType{Button: btn, Floor: floor}
				redirectedOrderCh <- order
			}
		}
	}

}

func removeRedirectedOrders(disconnectingElevatorID int, oldOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType) [cf.ElevatorCount]dt.OrderMatrixType {
	updatedOrderMatrices := oldOrderMatrices

	ownOrderMatrix := oldOrderMatrices[disconnectingElevatorID]

	for btnIndex, row := range ownOrderMatrix {
		btn := dt.ButtonType(btnIndex)
		for floor, orderState := range row {
			newOrderState := dt.NoOrder
			//We dont remove cab calls, but set them as AckedOrder
			//So that when the elevator reconnects it will execute the orders if it restarted
			if btn == dt.ButtonCab && isOrderActive(orderState) {
				newOrderState = dt.AckedOrder
			}
			oldOrder := &updatedOrderMatrices[disconnectingElevatorID][btnIndex][floor]
			*oldOrder = newOrderState
		}
	}
	return updatedOrderMatrices
}

func isOrderActive(orderState dt.OrderStateType) bool {
	if orderState == dt.NoOrder || orderState == dt.CompletedOrder || orderState == dt.UnknownOrder {
		return false
	}
	return true
}

func isSingleElevator(elevatorID int, connectedElevators [cf.ElevatorCount]connectionState) bool {

	for ID, state := range connectedElevators {
		if elevatorID != ID {
			if state == Connected {
				return false
			}
		}
	}
	return true
}

func isConnected(disconnectingElevatorID int, connectedElevators [cf.ElevatorCount]connectionState) bool {

	if connectedElevators[disconnectingElevatorID] == Connected {
		return true
	} else {
		return false
	}
}
