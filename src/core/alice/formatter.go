package alice

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	colorCodePanic = "\x1b[1;31m"
	colorCodeFatal = "\x1b[1;31m"
	colorCodeError = "\x1b[31m"
	colorCodeWarn  = "\x1b[33m"
	colorCodeInfo  = "\x1b[37m"
	colorCodeDebug = "\x1b[32m"
	colorCodeTrace = "\x1b[36m"
	colorReset     = "\x1b[0m"
)

type Formatter struct{}

func (formatter *Formatter) colorCode(level logrus.Level) string {
	switch level {
	case logrus.PanicLevel:
		return colorCodePanic
	case logrus.FatalLevel:
		return colorCodeFatal
	case logrus.ErrorLevel:
		return colorCodeError
	case logrus.WarnLevel:
		return colorCodeWarn
	case logrus.InfoLevel:
		return colorCodeInfo
	case logrus.DebugLevel:
		return colorCodeDebug
	case logrus.TraceLevel:
		return colorCodeTrace
	default:
		return colorCodeInfo
	}
}

func (formatter *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	if runtime.GOOS == "windows" {
		return []byte(fmt.Sprintf(
			"%s[alice-bot-go][%s][%s]%s%s\n",
			formatter.colorCode(entry.Level),
			entry.Time.Format("2006-01-02 15:04:05"),
			strings.ToUpper(entry.Level.String()),
			entry.Message,
			colorReset,
		)), nil
	} else {
		return []byte(fmt.Sprintf(
			"%s[alice-bot-go][%s][%s]%s%s\n",
			entry.Time.Format("2006-01-02 15:04:05"),
			strings.ToUpper(entry.Level.String()),
			entry.Message,
		)), nil
	}
}
