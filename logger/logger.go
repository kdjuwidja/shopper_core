package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var l *logrus.Logger

func Init() error {
	l = logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})

	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	l.SetLevel(lvl)

	return nil
}

func Close() {
	l = nil
}

func Info(args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": "core",
	}).Info(args...)
}

func Infof(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": "core",
	}).Infof(format, args...)
}

func Error(args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": "core",
	}).Error(args...)
}

func Errorf(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": "core",
	}).Errorf(format, args...)
}

func Debug(args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": "core",
	}).Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": "core",
	}).Debugf(format, args...)
}

func Warn(args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": "core",
	}).Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": "core",
	}).Warnf(format, args...)
}

func Fatal(args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": "core",
	}).Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": "core",
	}).Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": "core",
	}).Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": "core",
	}).Panicf(format, args...)
}
