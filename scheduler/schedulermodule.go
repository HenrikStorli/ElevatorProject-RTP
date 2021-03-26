package scheduler

import (
	dt "../datatypes"
)

func RunOrdersScheduler(
	newOrderIOCh <-chan dt.OrderType,
	newOrderSHCh <-chan dt.OrderType,
	elevatorStatesCh <-chan [dt.ElevatorCount]dt.ElevatorState,
	orderMatricesCh <-chan [dt.ElevatorCount]dt.OrderMatrixType,
	updateOrderMatricesCh chan<- [dt.ElevatorCount]dt.OrderMatrixType,

) {
	var elevatorStatesCopy [dt.ElevatorCount]dt.ElevatorState
	var orderMatricesCopy [dt.ElevatorCount]dt.OrderMatrixType

	for {
		select {
		case newOrder := <-newOrderIOCh:
			updatedOrderMatrices := placeOrder(newOrder, elevatorStatesCopy, orderMatricesCopy)
			updateOrderMatricesCh <- updatedOrderMatrices

		case newOrder := <-newOrderSHCh:
			updatedOrderMatrices := placeOrder(newOrder, elevatorStatesCopy, orderMatricesCopy)
			updateOrderMatricesCh <- updatedOrderMatrices

		case elevatorStatesUpdate := <-elevatorStatesCh:
			elevatorStatesCopy = elevatorStatesUpdate
		case orderMatricesUpdate := <-orderMatricesCh:
			orderMatricesCopy = orderMatricesUpdate
		}
	}

}

func placeOrder(
	newOrder dt.OrderType,
	elevatorStates [dt.ElevatorCount]dt.ElevatorState,
	orderMatrices [dt.ElevatorCount]dt.OrderMatrixType,
) [dt.ElevatorCount]dt.OrderMatrixType {
	var updatedOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType

	fastestElevatorIndex := findFastestElevator(elevatorStates, orderMatrices)

	updatedOrderMatrices[fastestElevatorIndex][newOrder.Button][newOrder.Floor] = dt.New

	return updatedOrderMatrices
}

func findFastestElevator(elevatorStates [dt.ElevatorCount]dt.ElevatorState, orderMatrices [dt.ElevatorCount]dt.OrderMatrixType) int {
	var fastestElevatorIndex int = 0
	var fastestExecutionTime int = 1000

	for elevatorIndex, state := range elevatorStates {
		if state.IsFunctioning {
			executionTime := TimeToIdle(state, orderMatrices[elevatorIndex])

			if executionTime < fastestExecutionTime {
				fastestExecutionTime = executionTime
				fastestElevatorIndex = elevatorIndex
			}
		}
	}
	return fastestElevatorIndex
}
