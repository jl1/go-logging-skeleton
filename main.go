package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

var verboseFlag bool
var debugFlag bool
var log *logrus.Logger

func init() {
	flag.BoolVar(&verboseFlag, "v", false, "verbose output. print info messages to stdout")
	flag.BoolVar(&debugFlag, "debug", false, "Set log level to trace for file and stdout. Overrides -v")
	flag.Parse()

	logDir := "./logs/"

	if err := initLogging(verboseFlag, debugFlag, logDir); err != nil {
		fmt.Println("failed to setup logging", err)
	}
	log.Info("log file started")

}

func main() {
	log.Info("entered main")
	log.Debug("debug message")
	log.Trace("trace message")
	log.Warn("warn message")
	log.Error("error message")
	log.Info("leaving main")
	log.Fatal("fatal message")
}

func initLogging(verboseFlag, debugFlag bool, logDir string) error {

	log = logrus.New()
	log.SetReportCaller(true)

	stdoutLevels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
	}

	fileLevels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}

	if debugFlag {
		fileLevels = append(fileLevels, logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel)
		stdoutLevels = append(stdoutLevels, logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel)
		log.SetLevel(logrus.TraceLevel)
	} else if verboseFlag {
		stdoutLevels = append(stdoutLevels, logrus.InfoLevel)
		log.SetLevel(logrus.InfoLevel)
	}

	log.SetOutput(io.Discard)
	log.SetFormatter(new(customFormatter))

	log.AddHook(&writer.Hook{
		Writer:    os.Stdout,
		LogLevels: stdoutLevels,
	})

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
			return err
		}
	}

	logFile := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "output.log"),
		MaxSize:    1, // MB
		MaxBackups: 3,
		MaxAge:     28, // Days
	}

	log.AddHook(&writer.Hook{
		Writer:    logFile,
		LogLevels: fileLevels,
	})

	return nil
}

type customFormatter struct{}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	formattedTime := entry.Time.Format(time.RFC3339)
	level := fmt.Sprintf("%-7s", entry.Level)
	level = strings.ToUpper(string(level[0])) + level[1:]

	var msg string
	if debugFlag {
		caller := ""
		if entry.Caller != nil {
			filename := filepath.Base(entry.Caller.File) // Extract only the filename
			caller = fmt.Sprintf("%s:%d", filename, entry.Caller.Line)
		}
		msg = fmt.Sprintf("%s | %s | %16s | %s\n", formattedTime, level, caller, entry.Message)
	} else {
		msg = fmt.Sprintf("%s | %s | %s\n", formattedTime, level, entry.Message)
	}

	return []byte(msg), nil
}
