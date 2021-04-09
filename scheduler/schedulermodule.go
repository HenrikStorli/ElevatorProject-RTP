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
	go setButtonLamps(elevatorID, orderMatrices, buttonLampCh)

	for {
		select {
		case newOrder := <-newOrderCh:
			if orderIsNew(elevatorID, newOrder, orderMatrices) {
				updatedOrderMatrices := placeOrder(elevatorID, newOrder, elevatorStates, orderMatrices)
				updateOrderMatricesCh <- updatedOrderMatrices
			}

		case elevatorStatesUpdate := <-elevatorStatesCh:
			elevatorStates = elevatorStatesUpdate
		case orderMatricesUpdate := <-orderMatricesCh:

			go setButtonLamps(elevatorID, orderMatricesUpdate, buttonLampCh)

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

	fmt.Printf("Directing Order %v to elevator %d \n", newOrder, fastestElevatorIndex+1)

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

func orderIsNew(elevatorID int, order dt.OrderType, orderMatrices [dt.ElevatorCount]dt.OrderMatrixType) bool {
	ownIndexID := elevatorID - 1

	for indexID := range orderMatrices {
		//Ignore cab calls from different elevators
		if order.Button == dt.BtnCab && ownIndexID != indexID {
			continue
		}
		switch orderMatrices[indexID][order.Button][order.Floor] {
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

	for elevatorIndex, state := range elevatorStates {
		if state.IsFunctioning {

			executionTime := timeToServeRequest(state, orderMatrices[elevatorIndex], newOrder)

			if executionTime < fastestExecutionTime {
				fastestExecutionTime = executionTime
				fastestElevatorIndex = elevatorIndex
			}
		}
	}
	return fastestElevatorIndex
}

func setButtonLamps(elevatorID int, newOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType, buttonLampCh chan<- iomodule.ButtonLampType) {

	ownIndexID := elevatorID - 1
	for rowIndex, row := range newOrderMatrices[0] {
		btn := dt.ButtonType(rowIndex)
		for floor, _ := range row {
			lampStatus := false
			order := dt.OrderType{Button: btn, Floor: floor}
			for indexID, orderMatrix := range newOrderMatrices {

				//cab calls lights up only on own elevator
				if btn != dt.BtnCab || ownIndexID == indexID {
					if orderMatrix[rowIndex][floor] == dt.Accepted {
						lampStatus = true
					}
				}
			}
			buttonLampCh <- iomodule.ButtonLampType{Order: order, TurnOn: lampStatus}
		}
	}

}
