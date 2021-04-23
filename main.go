package main

import (
	"flag"
	"fmt"
	"time"

	cf "./config"
	dt "./datatypes"
	"./elevatordriver"
	"./iomodule"
	"./netmodule"
	"./ordersscheduler"
	"./statehandler"
)

func main() {

	elevatorID, port := parseFlag()

	if !netmodule.IsValidID(elevatorID) {
		panic("Elevator ID is out of bounds")
	}

	chBufferSize := cf.ButtonCount * cf.FloorCount

	// Channels for elevator driver module
	driverStateUpdateCh := make(chan dt.ElevatorState, 1)
	acceptedOrderCh := make(chan dt.OrderType, chBufferSize)
	completedOrderFloorCh := make(chan int)
	connectNetworkCh := make(chan bool)

	// Channels for order scheduler module
	stateUpdateCh := make(chan [cf.ElevatorCount]dt.ElevatorState, 1)
	orderUpdateCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType, 1)
	scheduledOrdersCh := make(chan dt.OrderType, 10)

	// Channels for network module
	outgoingStateCh := make(chan dt.ElevatorState, 1)
	incomingStateCh := make(chan dt.ElevatorState)

	outgoingOrderCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType, 1)
	incomingOrderCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType)

	disconnectedIDCh := make(chan int)
	connectedIDCh := make(chan int)

	// Channels for IOModule
	motorDirCh := make(chan dt.MoveDirectionType)
	floorIndicatorCh := make(chan int)
	doorOpenCh := make(chan bool)
	stopLampCh := make(chan bool)
	buttonLampCh := make(chan iomodule.ButtonLampType, chBufferSize)

	buttonCallCh := make(chan dt.OrderType, 10)
	floorSensorCh := make(chan int)
	stopBtnCh := make(chan bool)
	obstructionSwitchCh := make(chan bool)

	fmt.Println("Starting Modules...")

	go netmodule.RunNetworkModule(
		elevatorID,
		outgoingStateCh, incomingStateCh,
		outgoingOrderCh, incomingOrderCh,
		disconnectedIDCh, connectedIDCh,
		connectNetworkCh,
	)

	go iomodule.RunIOModule(
		port,
		motorDirCh,
		floorIndicatorCh,
		doorOpenCh,
		stopLampCh,
		buttonLampCh,
		buttonCallCh,
		floorSensorCh,
		stopBtnCh,
		obstructionSwitchCh,
	)

	go statehandler.RunStateHandlerModule(
		elevatorID,
		incomingOrderCh, outgoingOrderCh,
		incomingStateCh, outgoingStateCh,
		disconnectedIDCh, connectedIDCh,
		stateUpdateCh, orderUpdateCh,
		scheduledOrdersCh,
		buttonCallCh,
		driverStateUpdateCh,
		acceptedOrderCh, completedOrderFloorCh,
	)

	go elevatordriver.RunElevatorDriverModule(
		elevatorID,
		driverStateUpdateCh,
		completedOrderFloorCh, acceptedOrderCh,
		connectNetworkCh,
		floorSensorCh, stopBtnCh, obstructionSwitchCh,
		floorIndicatorCh, motorDirCh, doorOpenCh, stopLampCh,
	)

	go ordersscheduler.RunOrdersSchedulerModule(
		elevatorID,
		buttonCallCh,
		stateUpdateCh, orderUpdateCh,
		scheduledOrdersCh,
		buttonLampCh,
	)

	fmt.Println("Successfully initialised modules!")

	for {
		time.Sleep(10 * time.Millisecond)
	}

}

func parseFlag() (int, int) {
	var elevatorID int
	var port int
	flag.IntVar(&elevatorID, "id", 0, "Id of the elevator")
	flag.IntVar(&port, "port", cf.DefaultIOPort, "IP port to harware server")
	flag.Parse()
	return elevatorID, port
}
