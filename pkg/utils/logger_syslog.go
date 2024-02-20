package utils

import (
	"log/syslog"

	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
)

type SyslogHook struct {
	*logrus_syslog.SyslogHook
}

func InitLoggers(logToSyslog bool) {
	if logToSyslog {
		hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_DEBUG|syslog.LOG_USER, "")
		if err != nil {
			// println("Unable to connect to local syslog daemon")
			return
		}
		syslogHook = &SyslogHook{hook}

		for _, l := range loggers {
			l.Hooks.Add(syslogHook)
		}
	}
}
