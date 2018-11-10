package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		panic("You must have commnad line two arguments")
	}
	switch os.Args[1] {
	case "run":
		run()
	default:
		panic("what???")
	}
}

func run() {
	fmt.Printf("Running %v\n", os.Args[2:])

	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	// Setup Stdin, Stdout, Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run())
}


func must(err error) {
	if err != nil {
		panic(err)
	}
}
