package iomodule_test

import (
	"flag"
	"fmt"
	"testing"

	dt "../datatypes"
	"../iomodule"
)

func TestRunIOModule(*testing.T) {

	_, port := parseFlag()

	motorDirCh := make(chan dt.MoveDirectionType)
	floorIndicatorCh := make(chan int)
	doorOpenCh := make(chan bool)
	stopLampCh := make(chan bool)

	buttonEventCh := make(chan dt.OrderType)
	floorSensorCh := make(chan int)
	stopBtnCh := make(chan bool)
	obstructionSwitchCh := make(chan bool)
	buttonLampCh := make(chan iomodule.ButtonLampType)

	go iomodule.RunIOModule(port,
		motorDirCh,
		floorIndicatorCh,
		doorOpenCh,
		stopLampCh,
		buttonLampCh,
		buttonEventCh,
		floorSensorCh,
		stopBtnCh,
		obstructionSwitchCh,
	)

	for {
		select {
		case buttonEvent := <-buttonEventCh:
			fmt.Printf("Received button event %v \n", buttonEvent)
		case floorSensor := <-floorSensorCh:
			fmt.Printf("Received floor sensor %v \n", floorSensor)
		case stopBtn := <-stopBtnCh:
			fmt.Printf("Received stop button %v \n", stopBtn)
		case obstructionSwitch := <-obstructionSwitchCh:
			fmt.Printf("Received obstruction switch %v \n", obstructionSwitch)
		}
	}
}

func parseFlag() (int, int) {
	var elevatorID int
	var port int
	flag.IntVar(&elevatorID, "id", 1, "Id of the elevator")
	flag.IntVar(&port, "port", 15657, "IP port to harware server")
	flag.Parse()
	return elevatorID, port
}
