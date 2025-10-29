package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/mathiasdonoso/j2g/internal/cli"
)

func main() {
	reader := os.Stdin
	writer := os.Stdout

	cli := cli.J2G{
		Input:  bufio.NewReader(reader),
		Output: bufio.NewWriter(writer),
	}

	err := cli.Start()
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n")
}
