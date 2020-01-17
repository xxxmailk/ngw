package log

import "github.com/sirupsen/logrus"

var log = logger()

func logger() *logrus.Entry {
	l := logrus.New()
	l.ReportCaller = true
	l.Level = logrus.TraceLevel
	return logrus.NewEntry(l)
}

func GetLogger() *logrus.Entry {
	return log
}
