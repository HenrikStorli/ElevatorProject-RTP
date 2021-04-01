package statehandler

import (
	"time"

	dt "../datastructures"
)

type OrderWithTime struct {
	Order            dt.OrderType
	ExpectedDuration time.Duration
}

type orderWDType struct {
	expectedDuration dt.duration
	startTime        time.Time
	active           bool
}

type orderWDMatrixType [dt.ButtonCount][dt.FloorCount]orderWDType

//Sends a timeout signal on a timeout channel when the duration has timeout
func RunOrderWatchdog(orderInCh <-chan OrderWithTime, orderOutCh chan<- dt.OrderType, completedFloorInCh <-chan int, completedFloorOutCh chan<- int, resetCh chan<- int) {

	var orderWatchdogMatrix orderWDMatrixType

	for {

		select {
		case orderIn := <-orderInCh:
			updatedMatrix := addWDOrders(orderIn, orderWatchdogMatrix)

			orderWatchdogMatrix = updatedMatrix

			go func() { orderOutCh <- orderIn }()

		case completedFloor := <-completedFloorInCh:
			updatedMatrix := completeWDOrders(completedFloor, orderWatchdogMatrix)

			orderWatchdogMatrix = updatedMatrix

			go func() { completedFloorOutCh <- completedFloor }()
		}

		if checkTimeout(orderWatchdogMatrix) {
			resetCh <- true
		}
		time.Sleep(time.Millisecond * 10)
	}
}

func completeWDOrders(completedFloor int, WDMatrix orderWDMatrixType) orderWDMatrixType {
	updatedMatrix := WDMatrix
	for rowIndex, _ := range WDMatrix {
		updatedMatrix[rowIndex][completedFloor].active = false
	}
	return updatedMatrix
}

func addWDOrders(orderIn OrderWithTime, oldMatrix orderWDMatrixType) orderWDMatrixType {
	updatedMatrix := oldMatrix

	btn := orderIn.Order.Button
	rowIndex := int(btn)
	floor := orderIn.Order.Floor

	updatedMatrix[rowIndex][floor] = orderWDType{
		expectedDuration: orderIn.ExpectedDuration,
		startTime:        time.Now(),
		active:           true,
	}

	return updatedMatrix
}

func checkTimeout(WDmatrix orderWDMatrixType) bool {
	for _, row := range WDmatrix {
		for _, cell := range row {
			if cell.active {
				initialTime := cell.startTime
				expectedDuration := cell.expectedDuration
				elapsedTime := time.Now().Sub(initialTime)

				if elapsedTime > expectedDuration {
					return true
				}
			}
		}
	}

	return false
}
