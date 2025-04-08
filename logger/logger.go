package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var l *logrus.Logger

var serviceName string

func Init(s string) error {
	l = logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})

	serviceName = s

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
		"service": serviceName,
	}).Info(args...)
}

func Infof(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Infof(format, args...)
}

func Error(args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Error(args...)
}

func Errorf(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Errorf(format, args...)
}

func Debug(args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Debugf(format, args...)
}

func Trace(args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Trace(args...)
}

func Warn(args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Warnf(format, args...)
}

func Fatal(args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Panicf(format, args...)
}

func Tracef(format string, args ...interface{}) {
	l.WithFields(logrus.Fields{
		"service": serviceName,
	}).Tracef(format, args...)
}
