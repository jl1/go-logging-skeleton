package main

import (
	"flag"
	"fmt"
	"os"
)

var verboseFlag bool
var debugFlag bool

func init() {
	flag.BoolVar(&verboseFlag, "v", false, "verbose output. print info messages to stdout")
	flag.BoolVar(&debugFlag, "debug", false, "Set log level to trace for file and stdout. Overrides -v")
	flag.Parse()

	logDir := "./logs/"

	if err := initLogging(verboseFlag, debugFlag, logDir); err != nil {
		fmt.Println("failed to setup logging", err)
		os.Exit(1)
	}
	log.Info("logging initialized")

}

func main() {
	log.Trace("trace message")
	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")
	log.Fatal("fatal message")
}
