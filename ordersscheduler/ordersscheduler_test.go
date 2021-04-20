package ordersscheduler_test

import (
	"fmt"
	"testing"
	"time"

	cf "../config"
	dt "../datatypes"
	"../iomodule"
	"../ordersscheduler"
)

func TestSchedulerModule(*testing.T) {
	newOrderCh := make(chan dt.OrderType)
	elevatorStatesCh := make(chan [cf.ElevatorCount]dt.ElevatorState)
	orderMatricesCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType)
	updateOrderMatricesCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType)
	buttonLampCh := make(chan iomodule.ButtonLampType)

	elevatorID := 0

	go ordersscheduler.RunOrdersSchedulerModule(elevatorID, newOrderCh,
		elevatorStatesCh, orderMatricesCh, updateOrderMatricesCh, buttonLampCh)
	//Define input
	var mockStates [cf.ElevatorCount]dt.ElevatorState
	var mockOrders [cf.ElevatorCount]dt.OrderMatrixType
	mockOrders[0][0][0] = dt.NewOrder
	mockOrders[1][0][0] = dt.NewOrder
	mockOrders[1][1][0] = dt.NewOrder
	mockOrders[1][2][0] = dt.AcceptedOrder
	fmt.Println(mockOrders)

	mockStates[0].IsFunctioning = true
	mockStates[0].MovingDirection = dt.MovingUp
	mockStates[0].Floor = 1

	elevatorStatesCh <- mockStates
	orderMatricesCh <- mockOrders
	time.Sleep(10 * time.Millisecond)
	newOrderCh <- dt.OrderType{Button: dt.ButtonHallUp, Floor: 1}
	time.Sleep(10 * time.Millisecond)
	newOrderCh <- dt.OrderType{Button: dt.ButtonHallDown, Floor: 3}

	updateOrderMatrices := <-updateOrderMatricesCh
	fmt.Println(updateOrderMatrices)

	updateOrderMatrices = <-updateOrderMatricesCh
	fmt.Println(updateOrderMatrices)

}
