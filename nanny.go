package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	cf "./config"
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
	restartCh := make(chan bool)

	for {
		fmt.Println("###         Starting process       ###")

		if runtime.GOOS == "windows" {
			runningProcess = exec.Command(cmdName+".exe", idFlag, strconv.Itoa(elevatorID), portFlag, strconv.Itoa(port))
			// Copies printouts from main module into shell
			go readIOFromProcess(runningProcess, restartCh)
		} else {
			runningProcess = exec.Command("./"+cmdName, idFlag, strconv.Itoa(elevatorID), portFlag, strconv.Itoa(port))
			runningProcess.Stdout = os.Stdout
			runningProcess.Stderr = os.Stderr
		}
		fmt.Println(runningProcess)
		fmt.Println("  ")

		err := runningProcess.Run()
		_, ok := err.(*exec.ExitError)
		// Ignore Exit error, we don't want to crash if the main module crash
		if !ok {
			checkError(err)
		} else {
			fmt.Println(err)
		}

		runningProcess.Wait()

		fmt.Println("### Process terminated, restarting ###")
		restartCh <- true

		time.Sleep(time.Second)
	}

}

func readIOFromProcess(runningProcess *exec.Cmd, restartCh chan bool) {
	cmdReaderIO, err := runningProcess.StdoutPipe()
	checkError(err)
	cmdReaderErr, err := runningProcess.StderrPipe()
	checkError(err)

	restartCh1 := make(chan bool)
	restartCh2 := make(chan bool)

	scannerIO := bufio.NewScanner(cmdReaderIO)
	scannerErr := bufio.NewScanner(cmdReaderErr)
	go printFromScanner(scannerIO, restartCh1)
	go printFromScanner(scannerErr, restartCh2)

	for {
		select {
		case <-restartCh:
			restartCh1 <- true
			restartCh2 <- true
			return
		}
	}
}

func printFromScanner(scanner *bufio.Scanner, restartCh chan bool) {
	for scanner.Scan() {
		select {
		case <-restartCh:
			return
		default:
			fmt.Printf("  > %s\n", scanner.Text())
		}
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
