package iomodule_test

import (
	"fmt"
	"testing"

	dt "../datatypes"
	"../iomodule"
)

func TestRunIOModule(*testing.T) {

	motorDirCh := make(chan dt.MoveDirectionType)
	floorIndicatorCh := make(chan int)
	doorOpenCh := make(chan bool)
	stopLampCh := make(chan bool)

	buttonEventCh := make(chan dt.OrderType)
	floorSensorCh := make(chan int)
	stopBtnCh := make(chan bool)
	obstructionSwitchCh := make(chan bool)
	buttonLampCh := make(chan iomodule.ButtonLampType)

	go iomodule.RunIOModule(
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
