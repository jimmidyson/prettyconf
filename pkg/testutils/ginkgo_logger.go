package testutils

import (
	"fmt"
	"io"

	"github.com/go-logr/logr"
)

type GinkgoLogger struct {
	io.Writer
	keysAndValues []interface{}
}

var _ logr.Logger = &GinkgoLogger{}

func (*GinkgoLogger) Enabled() bool {
	return true
}

func (l *GinkgoLogger) Info(msg string, keysAndValues ...interface{}) {
	fmt.Fprintln(l.Writer, msg, l.keysAndValues, keysAndValues)
}

func (l *GinkgoLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	fmt.Fprintln(l.Writer, msg, err, l.keysAndValues, keysAndValues)
}

func (l *GinkgoLogger) V(level int) logr.InfoLogger {
	return l
}

func (l *GinkgoLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return &GinkgoLogger{
		l.Writer,
		append(l.keysAndValues, keysAndValues...),
	}
}

func (l *GinkgoLogger) WithName(name string) logr.Logger {
	return l
}
