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

	flagString := "--ID " + strconv.Itoa(elevatorID) + " --port " + strconv.Itoa(port)

	var runningProcess *exec.Cmd

	for {
		fmt.Println("###         Starting process       ###")

		if runtime.GOOS == "windows" {
			runningProcess = exec.Command(cmdName+".exe", flagString)
		} else {
			runningProcess = exec.Command(cmdName, flagString)
		}
		fmt.Println(runningProcess)
		fmt.Println("  ")

		cmdReader, err := runningProcess.StdoutPipe()
		checkError(err)

		scanner := bufio.NewScanner(cmdReader)
		go func() {
			for scanner.Scan() {
				fmt.Printf("  > %s\n", scanner.Text())
			}
		}()

		err = runningProcess.Run()
		checkError(err)

		runningProcess.Wait()

		fmt.Println("### Process terminated, restarting ###")

		time.Sleep(time.Second)
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
