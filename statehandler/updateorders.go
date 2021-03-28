package statehandler

import (
	"time"
	//"fmt"
	dt "../datatypes"
)

func replaceNewOrders(elevatorID int, newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType, singleElevator bool) [dt.ElevatorCount]dt.OrderMatrixType {
	indexID := elevatorID - 1
	updatedOrderMatrices := newOrderMatrices

	for ID := range newOrderMatrices {
		if ID != indexID || singleElevator {
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

func updateIncomingOrders(newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType, oldOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType) [dt.ElevatorCount]dt.OrderMatrixType {

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

func completeOrders(elevatorID int, completedOrderFloor int, oldOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType) [dt.ElevatorCount]dt.OrderMatrixType {

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

func isSingleElevator(elevatorID int, connectedElevators [dt.ElevatorCount]connectionState) bool {
	ownIndexID := elevatorID - 1
	for indexID, state := range connectedElevators {
		if ownIndexID != indexID {
			if state == Connected {
				return false
			}
		}
	}
	return true
}
