package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/mathiasdonoso/j2g/internal/cli"
)

func main() {
	input := bufio.NewReader(os.Stdin)
	output := bufio.NewWriter(os.Stdout)
	cli := cli.J2G{
		Input:  input,
		Output: output,
	}

	err := cli.Start()
	if err != nil {
		panic(err)
	}
	defer fmt.Println()
	defer output.Flush()
}
