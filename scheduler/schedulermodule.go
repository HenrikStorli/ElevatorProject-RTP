package scheduler

import (
	"fmt"

	dt "../datatypes"
)

func RunOrdersScheduler(
	elevatorID int,
	newOrderCh <-chan dt.OrderType,
	elevatorStatesCh <-chan [dt.ElevatorCount]dt.ElevatorState,
	orderMatricesCh <-chan [dt.ElevatorCount]dt.OrderMatrixType,
	updateOrderMatricesCh chan<- [dt.ElevatorCount]dt.OrderMatrixType,

) {
	var elevatorStatesCopy [dt.ElevatorCount]dt.ElevatorState
	var orderMatricesCopy [dt.ElevatorCount]dt.OrderMatrixType

	for {
		select {
		case newOrder := <-newOrderCh:
			if orderIsNew(newOrder, orderMatricesCopy) {
				updatedOrderMatrices := placeOrder(elevatorID, newOrder, elevatorStatesCopy, orderMatricesCopy)
				updateOrderMatricesCh <- updatedOrderMatrices
			}

		case elevatorStatesUpdate := <-elevatorStatesCh:
			elevatorStatesCopy = elevatorStatesUpdate
		case orderMatricesUpdate := <-orderMatricesCh:
			orderMatricesCopy = orderMatricesUpdate
		}
	}
}

func placeOrder(
	elevatorID int,
	newOrder dt.OrderType,
	elevatorStates [dt.ElevatorCount]dt.ElevatorState,
	orderMatrices [dt.ElevatorCount]dt.OrderMatrixType,
) [dt.ElevatorCount]dt.OrderMatrixType {

	updatedOrderMatrices := orderMatrices
	indexID := elevatorID - 1
	var fastestElevatorIndex int = indexID
	//fmt.Println("In placeOrder")

	//fastestElevatorIndex := findFastestElevator(elevatorStates, orderMatrices)

	//Cab calls are always directed to this elevator
	if newOrder.Button == dt.BtnCab {
		fastestElevatorIndex = indexID
	} else {
		fastestElevatorIndex = findFastestElevatorServeRquest(elevatorStates, orderMatrices, newOrder)
	}

	updatedOrderMatrices[fastestElevatorIndex][newOrder.Button][newOrder.Floor] = dt.New

	return updatedOrderMatrices
}

func findFastestElevator(elevatorStates [dt.ElevatorCount]dt.ElevatorState, orderMatrices [dt.ElevatorCount]dt.OrderMatrixType) int {
	var fastestElevatorIndex int = 0
	var fastestExecutionTime int = 1000
	fmt.Println("In findFastestElevator")
	for elevatorIndex, state := range elevatorStates {
		if state.IsFunctioning {
			fmt.Println("In findFastestElevator inside if isfunctioning statement")
			executionTime := TimeToIdle(state, orderMatrices[elevatorIndex])

			if executionTime < fastestExecutionTime {
				fastestExecutionTime = executionTime
				fastestElevatorIndex = elevatorIndex
			}
		}
	}
	return fastestElevatorIndex
}

func orderIsNew(order dt.OrderType, orderMatrices [dt.ElevatorCount]dt.OrderMatrixType) bool {
	for elev := 0; elev < dt.ElevatorCount; elev++ {
		switch orderMatrices[elev][order.Button][order.Floor] {
		case dt.Accepted:
			return false
		case dt.New:
			return false
		case dt.Acknowledged:
			return false
		default:
		}
	}
	return true
}

// Testing new cost fucntion
func findFastestElevatorServeRquest(elevatorStates [dt.ElevatorCount]dt.ElevatorState, orderMatrices [dt.ElevatorCount]dt.OrderMatrixType, newOrder dt.OrderType) int {
	var fastestElevatorIndex int = 0
	var fastestExecutionTime int = 1000
	fmt.Println("In findFastestElevator")
	for elevatorIndex, state := range elevatorStates {
		if state.IsFunctioning {
			fmt.Println("In findFastestElevator inside if isfunctioning statement")
			executionTime := timeToServeRequest(state, orderMatrices[elevatorIndex], newOrder)

			if executionTime < fastestExecutionTime {
				fastestExecutionTime = executionTime
				fastestElevatorIndex = elevatorIndex
			}
		}
	}
	return fastestElevatorIndex
}
