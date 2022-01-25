package logging

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

func New(file, initialContext string) (*logrus.Entry, func() error, error) {
	fh, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		return nil, nil, err
	}

	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	l.SetOutput(io.MultiWriter(os.Stdout, fh))
	l.SetLevel(logrus.DebugLevel)

	return l.WithField("context", initialContext), func() error {
		return fh.Close()
	}, nil
}
