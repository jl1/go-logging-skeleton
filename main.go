package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

var verboseFlag bool
var debugFlag bool

func init() {
	flag.BoolVar(&verboseFlag, "v", false, "Verbose. Also print messages to stdout")
	flag.BoolVar(&debugFlag, "debug", false, "Set log level to debug")
	flag.Parse()

	logDir := "./logs/"

	// setup lopgging
	if err := initLogging(verboseFlag, debugFlag, logDir); err != nil {
		fmt.Println("failed to setup logging", err)
	}
	log.Info("log file started")

}

var log *logrus.Logger

func initLogging(verboseFlag, debugFlag bool, logDir string) error {

	log = logrus.New()
	if debugFlag {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	log.SetOutput(io.Discard)
	log.SetFormatter(new(customFormatter))

	stdoutLevels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
	}

	if verboseFlag {
		stdoutLevels = append(stdoutLevels, logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel)

	}

	if debugFlag {
		stdoutLevels = append(stdoutLevels, logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel)
	}

	log.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer:    os.Stdout,
		LogLevels: stdoutLevels,
	})

	logFile := &lumberjack.Logger{
		Filename:   logDir + "/output.log",
		MaxSize:    1, // MB
		MaxBackups: 3,
		MaxAge:     28, // Days
	}

	log.AddHook(&writer.Hook{ // Send info and debug logs to stdout
		Writer:    logFile,
		LogLevels: logrus.AllLevels,
	})

	return nil
}

type customFormatter struct{}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	formattedTime := entry.Time.Format(time.RFC3339)
	level := fmt.Sprintf("%-7s", entry.Level)
	level = strings.ToUpper(string(level[0])) + level[1:]
	msg := fmt.Sprintf("%s | %s | %s\n", formattedTime, level, entry.Message)

	return []byte(msg), nil
}

func main() {
	log.Info("entered main")
	log.Debug("debug message")
	log.Trace("trace message")
	log.Warn("warn message")
	log.Error("error message")
	log.Fatal("fatal message")
	log.Info("leaving main")
}
