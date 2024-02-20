package utils

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var mu sync.Mutex
var loggers = make(map[string]*logHandle)

var syslogHook logrus.Hook

type logHandle struct {
	logrus.Logger

	name string
	lvl  *logrus.Level
}

// Log formatter
func (l *logHandle) Format(e *logrus.Entry) ([]byte, error) {
	// Tue Feb 20 15:04:05 -0700 MST 2006
	timestamp := ""
	lvl := e.Level
	if l.lvl != nil {
		lvl = *l.lvl
	}

	const timeFormat = "2024/02/20 15:04:05.000000"
	timestamp = e.Time.Format(timeFormat)

	str := fmt.Sprintf("%v %s[%d] <%v>: %v",
		timestamp,
		l.name,
		os.Getpid(),
		strings.ToUpper(lvl.String()),
		e.Message)

	if len(e.Data) != 0 {
		str += " " + fmt.Sprint(e.Data)
	}

	str += "\n"
	return []byte(str), nil
}

// for aws.Logger
func (l *logHandle) Log(args ...interface{}) {

}

// TODO: create a new logger instance and return the logHandle struct
func NewLogger(name string) *logHandle {

}

// TODO: fetch a specific logger and return its logHandle struct
func GetLogger(name string) *logHandle {

}

// TODO: Set loglevel for every logger in loggers map
func SetLogLevel(lvl logrus.Level) {

}
