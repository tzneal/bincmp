package objdump

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

// run executes a process and returns a scanner that allows parsing stdout line
// by line.
func run(args ...string) *bufio.Scanner {
	cmd := exec.Command(args[0], args[1:]...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error executing command: %v\n", args)
		os.Exit(1)
	}
	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error starting command: %v\n", args)
		os.Exit(1)
	}

	return bufio.NewScanner(cmdReader)
}
