package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mathiasdonoso/j2g/internal/cli"
)

func main() {
	outputPtr := flag.String("o", "stdout", "output")
	flag.Parse()

	var reader io.Reader
	var writer io.Writer

	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if stat.Mode()&os.ModeCharDevice == 0 {
		fmt.Printf("Input is stdin\n")
		reader = os.Stdin
	} else {
		filename := os.Args[1]
		fmt.Printf("Input is the file %s\n", filename)
		reader, err = os.Open(filename)
		if err != nil {
			panic(err)
		}
		if c, ok := reader.(io.Closer); ok {
			defer c.Close()
		}
	}

	if *outputPtr != "stdout" {
		fmt.Printf("Output is the file %s\n", *outputPtr)
		writer, err = os.Create(*outputPtr)
		if err != nil {
			panic(err)
		}
		if c, ok := writer.(io.Closer); ok {
			defer c.Close()
		}
	} else {
		fmt.Printf("Output is stdout\n")
		writer = os.Stdout
	}

	cli := cli.J2G{
		Input:  bufio.NewReader(reader),
		Output: bufio.NewWriter(writer),
	}

	err = cli.Start()
	if err != nil {
		panic(err)
	}
}
