package logging

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func New(file, initialContext string) (*logrus.Entry, func() error, error) {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	l.SetLevel(logrus.DebugLevel)

	if strings.TrimSpace(file) == "" {
		l.SetOutput(os.Stdout)
		return l.WithField("context", initialContext), func() error { return nil }, nil
	}

	fh, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o744)
	if err != nil {
		return nil, nil, err
	}

	l.SetOutput(io.MultiWriter(os.Stdout, fh))

	return l.WithField("context", initialContext), fh.Close, nil
}
