package knet

import "log"

type Logger interface {
	Printf(fmt string, args ...interface{})
}

type stdLogger struct{}

func (l *stdLogger) Printf(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}

var defaultLogger Logger = &stdLogger{}

func SetLogger(logger Logger) {
	defaultLogger = logger
}
