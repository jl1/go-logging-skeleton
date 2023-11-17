package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

var log *logrus.Logger

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
		fileLevels = append(fileLevels, logrus.DebugLevel, logrus.TraceLevel)
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

	// get the current file's full path.
	_, currentFilePath, _, _ := runtime.Caller(0)

	// derive the directory from the full path.
	currentDir := filepath.Dir(currentFilePath)

	// the program's intended name would typically be the name of the directory.
	programName := filepath.Base(currentDir)

	logFile := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, programName+".log"),
		MaxSize:    10, // MB
		MaxBackups: 10,
		MaxAge:     7, // Days
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

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	formattedTime := entry.Time.Format(time.RFC3339)
	level := fmt.Sprintf("%-7s", entry.Level)
	level = strings.ToUpper(string(level[0])) + level[1:]

	var msg string
	if debugFlag {
		caller := ""
		if entry.Caller != nil {
			filename := filepath.Base(entry.Caller.File) // extract filename
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
	logrus.TraceLevel: "\x1b[90m", // grey
	logrus.DebugLevel: "\x1b[34m", // blue
	logrus.InfoLevel:  "\x1b[97m", // white
	logrus.WarnLevel:  "\x1b[33m", // yellow
	logrus.ErrorLevel: "\x1b[31m", // red
	logrus.FatalLevel: "\x1b[35m", // magenta
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
			filename := filepath.Base(entry.Caller.File) // extract filename
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
