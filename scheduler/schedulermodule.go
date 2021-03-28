package scheduler

import (
	"fmt"

	dt "../datatypes"
	"../iomodule"
)

func RunOrdersScheduler(
	elevatorID int,
	newOrderCh <-chan dt.OrderType,
	elevatorStatesCh <-chan [dt.ElevatorCount]dt.ElevatorState,
	orderMatricesCh <-chan [dt.ElevatorCount]dt.OrderMatrixType,
	updateOrderMatricesCh chan<- [dt.ElevatorCount]dt.OrderMatrixType,
	//Interface towards iomodule
	buttonLampCh chan<- iomodule.ButtonLampType,
) {
	var elevatorStates [dt.ElevatorCount]dt.ElevatorState
	var orderMatrices [dt.ElevatorCount]dt.OrderMatrixType

	//Reset all button lamps at init
	go setButtonLamps(orderMatrices, buttonLampCh)

	for {
		select {
		case newOrder := <-newOrderCh:
			if orderIsNew(newOrder, orderMatrices) {
				updatedOrderMatrices := placeOrder(elevatorID, newOrder, elevatorStates, orderMatrices)
				updateOrderMatricesCh <- updatedOrderMatrices
			}

		case elevatorStatesUpdate := <-elevatorStatesCh:
			elevatorStates = elevatorStatesUpdate
		case orderMatricesUpdate := <-orderMatricesCh:

			go setButtonLamps(orderMatricesUpdate, buttonLampCh)

			orderMatrices = orderMatricesUpdate
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
