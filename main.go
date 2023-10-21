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

func initLogging(verboseFlag, debugFlag bool, logDir string) error {

	log = logrus.New()
	log.SetReportCaller(true)
	log.SetFormatter(&logrus.TextFormatter{ForceColors: true})

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
	log.SetFormatter(new(customFormatterWithColour))

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

	log.AddHook(&fileHook{
		writer:    logFile,
		logLevels: fileLevels,
	})

	return nil
}

type fileHook struct {
	writer    io.Writer
	logLevels []logrus.Level
}

func (hook *fileHook) Levels() []logrus.Level {
	return hook.logLevels
}

func (hook *fileHook) Fire(entry *logrus.Entry) error {
	line, err := new(customFormatter).Format(entry)
	if err != nil {
		return err
	}
	_, err = hook.writer.Write(line)
	return err
}

type customFormatter struct{}

// var logrusColors = logrus.TextFormatter{}.ForceColors().GetColorMap()

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

type customFormatterWithColour struct{}

var logrusColors = map[logrus.Level]string{
	// set TraceLevel to dark grey
	logrus.TraceLevel: "\x1b[90m", // grey
	// set DebugLevel to blue
	logrus.DebugLevel: "\x1b[34m", // blue
	logrus.InfoLevel:  "\x1b[97m", // white
	logrus.WarnLevel:  "\x1b[33m", // yellow
	logrus.ErrorLevel: "\x1b[31m", // red
	logrus.FatalLevel: "\x1b[35m", // magenta
	// orange (approximated with bright yellow)
	logrus.PanicLevel: "\x1b[35m", // magenta
}

func (f *customFormatterWithColour) Format(entry *logrus.Entry) ([]byte, error) {
	formattedTime := entry.Time.Format(time.RFC3339)
	color, ok := logrusColors[entry.Level]
	if !ok {
		color = "\x1b[37m" // default to white if color not found
	}

	var msg string

	if debugFlag {
		level := fmt.Sprintf("%-7s", entry.Level)
		level = strings.ToUpper(string(level[0])) + level[1:]

		caller := ""
		if entry.Caller != nil {
			filename := filepath.Base(entry.Caller.File) // Extract only the filename
			caller = fmt.Sprintf("%s:%d", filename, entry.Caller.Line)
		}

		msg = fmt.Sprintf("%s%s | %s | %16s | %s\x1b[0m\n", color, formattedTime, level, caller, entry.Message)
	} else {
		level := fmt.Sprintf("%s%-7s\x1b[0m", color, entry.Level)
		level = strings.ToUpper(string(level[0])) + level[1:]
		msg = fmt.Sprintf("%s | %s | %s\n", formattedTime, level, entry.Message)
	}

	return []byte(msg), nil
}
