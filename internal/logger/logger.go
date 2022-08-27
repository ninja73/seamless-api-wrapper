package logger

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
	"io"
)

func InitLogger(out io.Writer, level log.Level) {
	log.SetOutput(out)
	log.SetLevel(level)
	log.SetReportCaller(true)

	log.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		CallerFirst:     true,
	})
}
