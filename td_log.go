
package main

import (
	"log"
	"os"
)

var tdLogF *os.File    = nil
var tdLog  *log.Logger = nil

func LogStart() {
	tdLogF = FileOpen(LogFileName, false)
	tdLog  = log.New(tdLogF, "", log.LstdFlags)
}

func LogPrint(v ...interface{}) {
	tdLog.Print(v...)
}

func LogPrintf(fmt string, v ...interface{}) {
	tdLog.Printf(fmt, v...)
}

