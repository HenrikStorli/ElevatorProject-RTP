# Elevator Project
================

## Summary
-------
Controlling `n` elevators working in parallel across `m` floors.


## Building and running the project

To run:

`go run ./main.go -id <elevator ID, 0, 1, 2...> -port <port, 54321>`
- The elevator ID is a unique number starting from 0, and must not exceed the number of allowed elevators -1.
- The port number must match the port used by the elevator harware driver server, or the elevator simulator.

To build:

`go build main.go`

This will create an executable called "main".

To fix import issues:
run `go env -w GO111MODULE=auto`

Windows:
- remember to add to environment GOPATH (Miljøvariabler på norsk)
- Open Control Panel » System » Advanced » Environment Variables
- Click on GOPATH and select edit
- add path to project, like "C:/go/project-gruppe-63"
- apparently only one path should be set in GOPATH, so delete the existing one


## Credits
The golang packages "elevio" and "network" have been borrowed from https://github.com/TTK4145/driver-go, and https://github.com/TTK4145/Network-go. The cost function is inspired by the cost function in https://github.com/TTK4145/Project-resources 

