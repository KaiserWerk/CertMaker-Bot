package logging

import log "github.com/sirupsen/logrus"

var logger *log.Entry

func SetLogger(l *log.Entry) {
	logger = l
}

func GetLogger() *log.Entry {
	return logger
}