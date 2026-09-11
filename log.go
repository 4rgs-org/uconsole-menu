package main

import (
	"fmt"
	"log"
	"os"
)

var appLog *log.Logger

func setupLog() {
	logPath := "/tmp/uconsole-menu.log"
	if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		appLog = log.New(f, "", log.LstdFlags|log.Lmicroseconds)
	} else {
		appLog = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)
	}
	appLog.Printf("=== start pid=%d argv=%v ===", os.Getpid(), os.Args)
	if dir, err := os.UserHomeDir(); err == nil {
		appLog.Printf("homedir=%s", dir)
	}
	if wd, err := os.Getwd(); err == nil {
		appLog.Printf("cwd=%s", wd)
	}
	xdg := os.Getenv("XDG_RUNTIME_DIR")
	wayland := os.Getenv("WAYLAND_DISPLAY")
	appLog.Printf("env WAYLAND_DISPLAY=%q XDG_RUNTIME_DIR=%q", wayland, xdg)
}

func logf(format string, args ...interface{}) {
	if appLog == nil {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
		return
	}
	appLog.Printf(format, args...)
}

func recoverPanic() {
	if r := recover(); r != nil {
		logf("PANIC recovered: %v", r)
	}
}
