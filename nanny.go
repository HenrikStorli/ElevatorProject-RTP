package main

import (
	"bufio"
	"flag"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

const (
	cmdName string = "main"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	elevatorID, port := parseFlag()

	idFlag := "-id"
	portFlag := "-port"

	var runningProcess *exec.Cmd

	for {
		fmt.Println("###         Starting process       ###")

		if runtime.GOOS == "windows" {
			runningProcess = exec.Command(cmdName+".exe", idFlag, strconv.Itoa(elevatorID), portFlag, strconv.Itoa(port))
		} else {
			runningProcess = exec.Command(cmdName, idFlag, portFlag)
		}
		fmt.Println(runningProcess)
		fmt.Println("  ")

		readIOFromProcess(runningProcess)

		err := runningProcess.Run()
		checkError(err)

		runningProcess.Wait()

		fmt.Println("### Process terminated, restarting ###")

		time.Sleep(time.Second)
	}

}

func readIOFromProcess(runningProcess *exec.Cmd) {
	cmdReaderIO, err := runningProcess.StdoutPipe()
	checkError(err)
	cmdReaderErr, err := runningProcess.StderrPipe()
	checkError(err)

	scannerIO := bufio.NewScanner(cmdReaderIO)
	scannerErr := bufio.NewScanner(cmdReaderErr)
	go func() {
		for scannerIO.Scan() {
			fmt.Printf("  > %s\n", scannerIO.Text())
		}
		for scannerErr.Scan() {
			fmt.Printf("  > %s\n", scannerErr.Text())
		}
	}()
}

func parseFlag() (int, int) {
	var elevatorID int
	var port int
	flag.IntVar(&elevatorID, "id", 1, "Id of the elevator")
	flag.IntVar(&port, "port", 15657, "IP port to harware server")
	flag.Parse()
	return elevatorID, port
}
