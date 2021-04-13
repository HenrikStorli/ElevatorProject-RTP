package scheduler

import (
	"fmt"

	cf "../config"
	dt "../datatypes"
	"../iomodule"
)

func RunOrdersScheduler(
	elevatorID int,
	newOrderCh <-chan dt.OrderType,
	elevatorStatesCh <-chan [cf.ElevatorCount]dt.ElevatorState,
	orderMatricesCh <-chan [cf.ElevatorCount]dt.OrderMatrixType,
	newScheduledOrderCh chan<- dt.OrderType,
	//Interface towards iomodule
	buttonLampCh chan<- iomodule.ButtonLampType,
) {
	var elevatorStates [cf.ElevatorCount]dt.ElevatorState
	var orderMatrices [cf.ElevatorCount]dt.OrderMatrixType

	//Reset all button lamps at init
	go setButtonLamps(elevatorID, orderMatrices, buttonLampCh)

	for {
		select {
		case newOrder := <-newOrderCh:
			if orderIsNew(elevatorID, newOrder, orderMatrices) {

				scheduledOrder := placeOrder(elevatorID, newOrder, elevatorStates, orderMatrices)
				newScheduledOrderCh <- scheduledOrder
			}

		case elevatorStatesUpdate := <-elevatorStatesCh:
			elevatorStates = elevatorStatesUpdate
		case orderMatricesUpdate := <-orderMatricesCh:

			setButtonLamps(elevatorID, orderMatricesUpdate, buttonLampCh)

			orderMatrices = orderMatricesUpdate
		}
	}
}

func placeOrder(
	elevatorID int,
	newOrder dt.OrderType,
	elevatorStates [cf.ElevatorCount]dt.ElevatorState,
	orderMatrices [cf.ElevatorCount]dt.OrderMatrixType,
) dt.OrderType {

	var scheduledOrder dt.OrderType = newOrder

	var fastestElevatorIndex int = elevatorID
	//fmt.Println("In placeOrder")

	//fastestElevatorIndex := findFastestElevator(elevatorStates, orderMatrices)

	//Cab calls are always directed to this elevator
	if newOrder.Button == dt.BtnCab {
		fastestElevatorIndex = elevatorID
	} else {
		fastestElevatorIndex = findFastestElevatorServeRquest(elevatorStates, orderMatrices, newOrder)
	}

	scheduledOrder.ElevatorID = fastestElevatorIndex

	fmt.Printf("Directing Order %v to elevator %d \n", newOrder, fastestElevatorIndex+1)

	return scheduledOrder
}

func findFastestElevator(elevatorStates [cf.ElevatorCount]dt.ElevatorState, orderMatrices [cf.ElevatorCount]dt.OrderMatrixType) int {
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

func orderIsNew(elevatorID int, order dt.OrderType, orderMatrices [cf.ElevatorCount]dt.OrderMatrixType) bool {

	for indexID := range orderMatrices {
		//Ignore cab calls from different elevators
		if order.Button == dt.BtnCab && elevatorID != indexID {
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
func findFastestElevatorServeRquest(elevatorStates [cf.ElevatorCount]dt.ElevatorState, orderMatrices [cf.ElevatorCount]dt.OrderMatrixType, newOrder dt.OrderType) int {
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

func setButtonLamps(elevatorID int, newOrderMatrices [cf.ElevatorCount]dt.OrderMatrixType, buttonLampCh chan<- iomodule.ButtonLampType) {

	for rowIndex, row := range newOrderMatrices[0] {
		btn := dt.ButtonType(rowIndex)
		for floor, _ := range row {
			lampStatus := false
			order := dt.OrderType{Button: btn, Floor: floor}
			for indexID, orderMatrix := range newOrderMatrices {

				//cab calls lights up only on own elevator
				if btn != dt.BtnCab || elevatorID == indexID {
					if orderMatrix[rowIndex][floor] == dt.Accepted {
						lampStatus = true
					}
				}
			}
			buttonLampCh <- iomodule.ButtonLampType{Order: order, TurnOn: lampStatus}
		}
	}

}
