package scheduler_test

import (
	"fmt"
	"testing"
	"time"

	dt "../datatypes"
	"../scheduler"
)

// func TestSchedulerModule(*testing.T) {
// 	newOrderIOCh := make(chan dt.OrderType)
// 	newOrderSHCh := make(chan dt.OrderType)
// 	elevatorStatesCh := make(chan [dt.ElevatorCount]dt.ElevatorState)
// 	orderMatricesCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)
// 	updateOrderMatricesCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)

// 	go scheduler.RunOrdersScheduler(newOrderIOCh, newOrderSHCh,
// 		elevatorStatesCh, orderMatricesCh, updateOrderMatricesCh)

// 	var mockStates [dt.ElevatorCount]dt.ElevatorState
// 	var mockOrders [dt.ElevatorCount]dt.OrderMatrixType
// 	mockStates[0].IsFunctioning = true
// 	go func() {
// 		elevatorStatesCh <- mockStates
// 		orderMatricesCh <- mockOrders
// 		time.Sleep(10 * time.Millisecond)
// 		newOrderIOCh <- dt.OrderType{Button: dt.BtnHallUp, Floor: 1}
// 		time.Sleep(10 * time.Millisecond)
// 		newOrderSHCh <- dt.OrderType{Button: dt.BtnHallDown, Floor: 3}
// 	}()

// 	updateOrderMatrices := <-updateOrderMatricesCh
// 	fmt.Println(updateOrderMatrices)

// 	updateOrderMatrices = <-updateOrderMatricesCh
// 	fmt.Println(updateOrderMatrices)

// }

func TestSchedulerModule(*testing.T) {
	newOrderIOCh := make(chan dt.OrderType)
	newOrderSHCh := make(chan dt.OrderType)
	elevatorStatesCh := make(chan [dt.ElevatorCount]dt.ElevatorState)
	orderMatricesCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)
	updateOrderMatricesCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)

	go scheduler.RunOrdersScheduler(newOrderIOCh, newOrderSHCh,
		elevatorStatesCh, orderMatricesCh, updateOrderMatricesCh)
	//Define input
	var mockStates [dt.ElevatorCount]dt.ElevatorState
	var mockOrders [dt.ElevatorCount]dt.OrderMatrixType
	mockOrders[0][0][0] = dt.New
	mockOrders[1][0][0] = dt.New
	mockOrders[1][1][0] = dt.New
	mockOrders[1][2][0] = dt.Accepted
	fmt.Println(mockOrders)

	mockStates[0].IsFunctioning = true
	mockStates[0].MovingDirection = dt.MovingUp
	mockStates[0].Floor = 1



	elevatorStatesCh <- mockStates
	orderMatricesCh <- mockOrders
	time.Sleep(10 * time.Millisecond)
	newOrderIOCh <- dt.OrderType{Button: dt.BtnHallUp, Floor: 1}
	time.Sleep(10 * time.Millisecond)
	newOrderSHCh <- dt.OrderType{Button: dt.BtnHallDown, Floor: 3}


	updateOrderMatrices := <-updateOrderMatricesCh
	fmt.Println(updateOrderMatrices)

	updateOrderMatrices = <-updateOrderMatricesCh
	fmt.Println(updateOrderMatrices)

}
