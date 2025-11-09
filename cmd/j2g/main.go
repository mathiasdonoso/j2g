package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/mathiasdonoso/j2g/internal/cli"
)

//go:embed usage.txt
var UsageText string

//go:embed error.txt
var ErrorText string

func initLogger() {
	d, _ := strconv.Atoi(os.Getenv("DEBUG"))
	if d != 1 {
		return
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))

	slog.Debug("debug mode enabled")
}

func checkFlags() {
	var showHelp = flag.Bool("h", false, "show help")
	var showHelpLong = flag.Bool("help", false, "show help")
	flag.Parse()

	if *showHelp || *showHelpLong {
		fmt.Printf("\n%s\n", UsageText)
		os.Exit(0)
	}
}

func showErrorMessage() {
	d, _ := strconv.Atoi(os.Getenv("DEBUG"))
	if d != 1 {
		fmt.Printf("\n%s\n", ErrorText)
	}
}

func main() {
	checkFlags()
	initLogger()

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
		slog.Debug(err.Error())
		showErrorMessage()
		os.Exit(1)
	}
	defer fmt.Println()
	defer output.Flush()
}
