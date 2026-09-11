package main

import (
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func main() {
	defer recoverPanic()
	setupLog()
	logf("main: starting uconsole-menu")

	if os.Getenv("WAYLAND_DISPLAY") == "" {
		os.Setenv("WAYLAND_DISPLAY", "wayland-1")
	}
	if os.Getenv("XDG_RUNTIME_DIR") == "" {
		os.Setenv("XDG_RUNTIME_DIR", "/run/user/1000")
	}

	gtk.Init(&os.Args)
	logf("gtk.Init OK")

	screen, err := gdk.ScreenGetDefault()
	if err != nil || screen == nil {
		logf("screen err: %v / nil", err)
		os.Exit(1)
	}

	theme := LoadTheme()
	logf("loaded theme: %s accent=%s bg=%s", theme.Name, theme.Accent, theme.Background)
	applyStyles(screen, theme)

	root := BuildMenu()
	count := countItems(root)
	logf("built menu tree: %d items total", count)

	win := createMainWindow(theme, root)
	if win == nil {
		logf("createMainWindow failed")
		os.Exit(1)
	}
	logf("entering gtk.Main()")
	gtk.Main()
	logf("gtk.Main returned")
}

func countItems(n *MenuNode) int {
	c := 0
	for _, child := range n.Children {
		c++
		if len(child.Children) > 0 {
			c += countItems(child)
		}
	}
	return c
}
