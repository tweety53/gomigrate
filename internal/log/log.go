package log

import (
	"log"
)

const (
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorGray   = "\033[37m"
	resetColor  = "\033[00m"
)

func Info(s string) {
	log.Print(colorGreen + s + resetColor)
}

func Warn(s string) {
	log.Print(colorYellow + s + resetColor)
}

func Err(s string) {
	log.Print(colorRed + s + resetColor)
}

func Infof(s string, args ...interface{}) {
	log.Printf(colorGreen+s+resetColor, args...)
}

func Warnf(s string, args ...interface{}) {
	log.Printf(colorYellow+s+resetColor, args...)
}

func Errf(s string, args ...interface{}) {
	log.Printf(colorRed+s+resetColor, args...)
}

func Infoln(args ...interface{}) {
	log.Println(colorGreen, args, resetColor)
}

func Warnln(args ...interface{}) {
	log.Println(colorYellow, args, resetColor)
}

func Errln(args ...interface{}) {
	log.Println(colorRed, args, resetColor)
}
