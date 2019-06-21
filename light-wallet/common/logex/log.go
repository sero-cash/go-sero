package logex

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"time"
	"os"
)

type Log struct {
	Name string
	Path string
}

var this *logrus.Logger

func (log Log) Setup() (*os.File, error) {
	log2 := logrus.New()
	log2.SetLevel(logrus.DebugLevel)
	log2.SetOutput(os.Stdout)
	log2.Formatter = &logrus.TextFormatter{
		DisableColors:  true,
		FullTimestamp:  true,
		DisableSorting: true,
	}

	curTime := time.Now()
	logFileName := fmt.Sprintf("%s/%s_%04d-%02d-%02d.log", log.Path, log.Name, curTime.Year(), curTime.Month(), curTime.Day())
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)

	if err != nil {
		fmt.Printf("try create logfile[%s] error[%s]\n", logFileName, err.Error())
		return nil, err
	}
	//log2.SetOutput(logFile)

	this = log2
	return logFile, err
}

func Info(args ...interface{}) {
	this.Infoln(args)
}

func Infof(format string, args ...interface{}) {
	this.Infof(format, args)
}

func Debug(args ...interface{}) {
	this.Debugln(args)
}

func Debugf(format string, args ...interface{}) {
	this.Debugf(format, args)
}

func Error(args ...interface{}) {
	this.Errorln(args)
}

func Errorf(format string, args ...interface{}) {
	this.Errorf(format, args)
}

func Fatal(args ...interface{}) {
	this.Fatal(args)
}

func Fatalf(format string, args ...interface{}) {
	this.Fatalf(format, args)
}
