package errors

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Logger struct {
	debugMode   bool
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
}

func NewLogger(setDebug bool) *Logger {
	return &Logger{
		debugMode:   setDebug,
		debugLogger: log.New(os.Stderr, "DEBUG: ", log.Ldate|log.Ltime),
		infoLogger:  log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime),
		warnLogger:  log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime),
		fatalLogger: log.New(os.Stderr, "FATAL: ", log.Ldate|log.Ltime),
	}

}

func (l *Logger) Debug(msg string) {
	if l.debugMode {
		l.debugLogger.Println(msg)
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *Logger) Info(msg string) {
	l.infoLogger.Println(msg)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(msg string) {
	l.warnLogger.Println(msg)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *Logger) Error(msg string) {
	l.errorLogger.Println(msg)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *Logger) Fatal(msg string) {
	l.fatalLogger.Println(msg)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format, args...))
}

func (l *Logger) CheckPanic(err error) {
	if err != nil {
		l.Fatal(err.Error())
		panic(err.Error())
	}
}
func (l *Logger) CheckError(err error) {
	if err != nil {
		l.Errorf("ERROR: %v", err.Error())
	}
}

func SetupCloseHandler() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	return c
}
