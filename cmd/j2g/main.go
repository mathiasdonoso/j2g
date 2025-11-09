package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/mathiasdonoso/j2g/internal/cli"
)

//go:embed usage.txt
var UsageText string

//go:embed error.txt
var ErrorText string

func main() {
	var showHelp = flag.Bool("h", false, "show help")
	var showHelpLong = flag.Bool("help", false, "show help")
	flag.Parse()

	if *showHelp || *showHelpLong {
		fmt.Printf("\n%s\n", UsageText)
		os.Exit(0)
	}

	input := bufio.NewReader(os.Stdin)
	output := bufio.NewWriter(os.Stdout)

	if input == nil || output == nil {
		os.Exit(0)
	}

	cli := cli.J2G{
		Input:  input,
		Output: output,
	}

	err := cli.Start()
	if err != nil {
		fmt.Printf("\n%s\n", ErrorText)
		os.Exit(1)
	}
	defer fmt.Println()
	defer output.Flush()
}
