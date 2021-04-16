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
			updatedOrderMatrices[indexID] = replaceExistingOrders(dt.Completed, dt.None, updatedOrderMatrices[indexID])
		}
	}
	return updatedOrderMatrices
}

func setNewOrdersToAck(elevatorID int, newOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType, singleElevator bool) [cf.ElevatorCount]dt.OrderMatrixType {

	updatedOrderMatrices := newOrderMatrices

	for indexID := range newOrderMatrices {
		if indexID != elevatorID || singleElevator {
			updatedOrderMatrices[indexID] = replaceExistingOrders(dt.New, dt.Acknowledged, updatedOrderMatrices[indexID])
		}
	}
	return updatedOrderMatrices
}

func acceptAndSendOrders(elevatorID int, newOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType, acceptedOrderCh chan<- dt.OrderType) [cf.ElevatorCount]dt.OrderMatrixType {

	updatedOrderMatrices := newOrderMatrices

	for btnIndex, row := range newOrderMatrices[elevatorID] {
		btn := dt.ButtonType(btnIndex)
		for floor, newOrder := range row {
			if newOrder == dt.Acknowledged {
				updatedOrderMatrices[elevatorID][btnIndex][floor] = dt.Accepted

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
	*oldOrder = updateSingleOrderState(dt.New, *oldOrder)

	return updatedOrderMatrices
}

//Updates a single order based on the order update rules
func updateSingleOrderState(newOrder dt.OrderStateType, oldOrder dt.OrderStateType) dt.OrderStateType {

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

func completeOrders(elevatorID int, completedOrderFloor int, oldOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType) [cf.ElevatorCount]dt.OrderMatrixType {

	updatedOrderMatrices := oldOrderMatrices
	floor := completedOrderFloor

	for btnIndex := range oldOrderMatrices {
		oldOrder := oldOrderMatrices[elevatorID][btnIndex][floor]
		if oldOrder == dt.Accepted {
			updatedOrderMatrices[elevatorID][btnIndex][floor] = dt.Completed
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
		if btn == dt.BtnCab {
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
			newOrderState := dt.None
			//We dont remove cab calls, but set them as Acknowledged
			//So that when the elevator reconnects it will execute the orders if it restarted
			if btn == dt.BtnCab && isOrderActive(orderState) {
				newOrderState = dt.Acknowledged
			}
			oldOrder := &updatedOrderMatrices[disconnectingElevatorID][btnIndex][floor]
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
