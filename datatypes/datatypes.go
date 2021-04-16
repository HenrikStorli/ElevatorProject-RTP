package datatypes

import (
	cf "../config"
)

type MoveDirectionType int

const (
	MovingDown    MoveDirectionType = -1
	MovingStopped                   = 0
	MovingUp                        = 1
)

type ButtonType int

const (
	BtnHallUp   ButtonType = 0
	BtnHallDown            = 1
	BtnCab                 = 2
)

type MachineStateType string

type OrderStateType int

const (
	Unknown OrderStateType = iota
	None
	New
	Acknowledged
	Accepted
	Completed
)

type OrderType struct {
	ElevatorID int
	Button     ButtonType
	Floor      int
}

//OrderMatrixType is the type for the order matrix
type OrderMatrixType [cf.ButtonCount][cf.FloorCount]OrderStateType

//ElevatorState ...
type ElevatorState struct {
	ElevatorID      int
	MovingDirection MoveDirectionType
	Floor           int
	State           MachineStateType
	IsFunctioning   bool
}

const (
	Init     		MachineStateType = "init"
	Idle 			MachineStateType = "idle"
	Moving   		MachineStateType = "moving"
	DoorOpen		MachineStateType = "door open"
)
