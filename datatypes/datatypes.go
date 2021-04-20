package datatypes

import (
	cf "../config"
)

type MoveDirectionType int

const (
	MovingInvalid MoveDirectionType = -99
	MovingDown                      = -1
	MovingNeutral                   = 0
	MovingUp                        = 1
)

type ButtonType int

const (
	ButtonHallUp   ButtonType = 0
	ButtonHallDown            = 1
	ButtonCab                 = 2
)

type DriverStateType string

type OrderStateType int

const (
	UnknownOrder OrderStateType = iota
	NoOrder
	NewOrder
	AckedOrder
	AcceptedOrder
	CompletedOrder
)

type OrderType struct {
	ElevatorID int
	Button     ButtonType
	Floor      int
}

//OrderMatrixType is the type for the order matrix
type OrderMatrixType [cf.ButtonCount][cf.FloorCount]OrderStateType

type ElevatorState struct {
	ElevatorID      int
	MovingDirection MoveDirectionType
	Floor           int
	State           DriverStateType
	IsFunctioning   bool
}

const (
	InitState     DriverStateType = "init"
	IdleState     DriverStateType = "idle"
	MovingState   DriverStateType = "moving"
	DoorOpenState DriverStateType = "door open"
	ErrorState    DriverStateType = "error"
	InvalidState  DriverStateType = "invalid"
)
