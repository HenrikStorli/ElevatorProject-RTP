package scheduler_test

import (
	"fmt"
	"testing"
	"time"

	dt "../datatypes"
	"../scheduler"
)

func TestSchedulerModule(*testing.T) {
	newOrderIOCh := make(chan dt.OrderType)
	newOrderSHCh := make(chan dt.OrderType)
	elevatorStatesCh := make(chan [dt.ElevatorCount]dt.ElevatorState)
	orderMatricesCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)
	updateOrderMatricesCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)

	go scheduler.RunOrdersScheduler(newOrderIOCh, newOrderSHCh,
		elevatorStatesCh, orderMatricesCh, updateOrderMatricesCh)

	var mockStates [dt.ElevatorCount]dt.ElevatorState
	var mockOrders [dt.ElevatorCount]dt.OrderMatrixType

	elevatorStatesCh <- mockStates
	orderMatricesCh <- mockOrders
	time.Sleep(10 * time.Millisecond)
	newOrderIOCh <- dt.OrderType{Button: dt.BtnHallUp, Floor: 1}
	time.Sleep(10 * time.Millisecond)
	newOrderSHCh <- dt.OrderType{Button: dt.BtnHallDown, Floor: 3}

	for {
		select {
		case updateOrderMatrices := <-updateOrderMatricesCh:
			fmt.Println(updateOrderMatrices)
		}
	}

}
