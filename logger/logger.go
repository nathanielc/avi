package logger

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var level = flag.Int("v", 2, "Verbosity level. One of 0,1,2,3 for ERR, WARN, INF, DBG respectively")

const logFlags = log.Ldate | log.Ltime | log.Lshortfile

const (
	ERROR = iota
	WARN
	INFO
	DEBUG
)

var (
	Debug *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

func Init() {
	Debug = log.New(getWriter(DEBUG), "DBG ", logFlags)
	Info = log.New(getWriter(INFO), "INF ", logFlags)
	Warn = log.New(getWriter(WARN), "WRN ", logFlags)
	Error = log.New(getWriter(ERROR), "ERR ", logFlags)
}

func getWriter(logLevel int) io.Writer {
	if *level >= logLevel {
		return os.Stderr
	}
	return ioutil.Discard
}
