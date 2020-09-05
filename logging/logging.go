// Package logging provides web access logging.
package logging

import (
	"fmt"
	"github.com/keep94/weblogs"
	"github.com/keep94/weblogs/loggers"
	"io"
	"net/http"
	"time"
)

type key int

const (
	kUserName key = iota
)

// SetUserName sets the current user name for logging.
func SetUserName(r *http.Request, name string) {
	values := weblogs.Values(r)
	if values != nil {
		values[kUserName] = name
	}
}

// ApacheCommonLoggerWithLatency provides apache common logs with latency
// in milliseconds following content size.
func ApacheCommonLoggerWithLatency() weblogs.Logger {
	return commonLogger{}
}

type loggerBase struct {
}

func (l loggerBase) NewSnapshot(r *http.Request) weblogs.Snapshot {
	return loggers.NewSnapshot(r)
}

func (l loggerBase) NewCapture(w http.ResponseWriter) weblogs.Capture {
	return &loggers.Capture{ResponseWriter: w}
}

type commonLogger struct {
	loggerBase
}

func (l commonLogger) Log(w io.Writer, log *weblogs.LogRecord) {
	s := log.R.(*loggers.Snapshot)
	c := log.W.(*loggers.Capture)
	fmt.Fprintf(w, "%s - %s [%s] \"%s %s %s\" %d %d %d\n",
		loggers.StripPort(s.RemoteAddr),
		userName(log),
		log.T.Format("02/Jan/2006:15:04:05 -0700"),
		s.Method,
		s.URL.RequestURI(),
		s.Proto,
		c.Status(),
		c.Size(),
		log.Duration/time.Millisecond)
}

func userName(log *weblogs.LogRecord) string {
	value, ok := log.Values[kUserName]
	if ok {
		return value.(string)
	}
	return "-"
}
