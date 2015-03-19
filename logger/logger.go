package logger

import (
	"flag"
	"log"
	"os"
)

var level = flag.Int("v", 2, "Verbosity level. One of 0,1,2,3 for ERR, WARN, INF, DBG respectively")

var debugLog = log.New(os.Stderr, "DBG ", log.Ldate|log.Ltime)
var infoLog = log.New(os.Stderr, "INF ", log.Ldate|log.Ltime)
var warnLog = log.New(os.Stderr, "WRN ", log.Ldate|log.Ltime)
var errorLog = log.New(os.Stderr, "ERR ", log.Ldate|log.Ltime)

const (
	ERROR = iota
	WARN
	INFO
	DEBUG
)

func Debugf(format string, v ...interface{}) {
	if *level >= DEBUG {
		debugLog.Printf(format, v...)
	}
}
func Debugln(v ...interface{}) {
	if *level >= DEBUG {
		debugLog.Println(v...)
	}
}

func Infof(format string, v ...interface{}) {
	if *level >= INFO {
		infoLog.Printf(format, v...)
	}
}
func Infoln(v ...interface{}) {
	if *level >= INFO {
		infoLog.Println(v...)
	}
}

func Warnf(format string, v ...interface{}) {
	if *level >= WARN {
		warnLog.Printf(format, v...)
	}
}
func Warnln(v ...interface{}) {
	if *level >= WARN {
		warnLog.Println(v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if *level >= ERROR {
		errorLog.Printf(format, v...)
	}
}
func Errorln(v ...interface{}) {
	if *level >= ERROR {
		errorLog.Println(v...)
	}
}

func Fatal(v ...interface{}) {
	errorLog.Fatal(v...)
}
