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

func isDebugMode() bool {
	d, _ := strconv.Atoi(os.Getenv("DEBUG"))
	return d == 1
}

func initLogger() {
	if !isDebugMode() {
		return
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))

	slog.Debug("debug mode enabled")
}

func checkFlags() {
	var showHelp bool
	flag.BoolVar(&showHelp, "h", false, "show help")
	flag.BoolVar(&showHelp, "help", false, "show help")
	flag.Parse()

	if showHelp {
		fmt.Printf("\n%s\n", UsageText)
		os.Exit(0)
	}
}

func showErrorMessage() {
	if !isDebugMode() {
		fmt.Printf("\n%s\n", ErrorText)
	}
}

func main() {
	checkFlags()
	initLogger()

	input := bufio.NewReader(os.Stdin)
	output := bufio.NewWriter(os.Stdout)

	defer func() {
		output.Flush()
		fmt.Println()
	}()

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
}
