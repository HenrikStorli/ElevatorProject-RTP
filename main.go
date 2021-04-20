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
	"./scheduler"
	"./statehandler"
)

func main() {

	elevatorID, port := parseFlag()

	if !netmodule.IsValidID(elevatorID) {
		panic("Elevator ID is out of bounds")
	}

	ports := netmodule.NetworkPorts{
		PeerTxPort:  cf.PeerTxPort,
		PeerRxPort:  cf.PeerRxPort,
		BcastRxPort: cf.BcastRxPort,
		BcastTxPort: cf.BcastTxPort,
	}

	orderMatrixBufferSize := cf.ButtonCount * cf.FloorCount

	stateUpdateCh := make(chan [cf.ElevatorCount]dt.ElevatorState, 1)
	orderUpdateCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType, 1)

	driverStateUpdateCh := make(chan dt.ElevatorState, 1)
	acceptedOrderCh := make(chan dt.OrderType, orderMatrixBufferSize)
	completedOrderFloorCh := make(chan int)

	connectNetworkCh := make(chan bool)

	scheduledOrdersCh := make(chan dt.OrderType, 10)
	buttonCallCh := make(chan dt.OrderType, 10)

	outgoingStateCh := make(chan dt.ElevatorState, 1)
	incomingStateCh := make(chan dt.ElevatorState)

	outgoingOrderCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType, 1)
	incomingOrderCh := make(chan [cf.ElevatorCount]dt.OrderMatrixType)

	disconnectedIDCh := make(chan int)
	connectedIDCh := make(chan int)

	motorDirCh := make(chan dt.MoveDirectionType)
	floorIndicatorCh := make(chan int)
	doorOpenCh := make(chan bool)
	stopLampCh := make(chan bool)
	buttonLampCh := make(chan iomodule.ButtonLampType, orderMatrixBufferSize)

	floorSensorCh := make(chan int)
	stopBtnCh := make(chan bool)
	obstructionSwitchCh := make(chan bool)

	time.Sleep(time.Second)
	fmt.Println("Starting Modules...")

	go netmodule.RunNetworkModule(
		elevatorID,
		ports,
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

	go scheduler.RunOrdersSchedulerModule(
		elevatorID,
		buttonCallCh,
		stateUpdateCh, orderUpdateCh,
		scheduledOrdersCh,
		buttonLampCh,
	)

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
