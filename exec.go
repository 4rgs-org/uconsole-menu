package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// termPrefix es el prefijo que usan los items que requieren terminal.
const termPrefix = "foot -e "

// popupAppID lo usa la regla for_window de Sway para hacer la terminal
// flotante y centrada (ver ~/.config/sway/config).
const popupAppID = "uconsole-popup"

// popupTerm es el prefijo de terminal con el app-id del popup.
const popupTerm = "foot --app-id=" + popupAppID + " -e "

// splitCommand clasifica un comando del menú.
// Devuelve el binario real, el comando sin el prefijo de terminal,
// y si necesita una terminal para correr.
//
//	"foot -e btop"        -> ("btop",      "btop",            true)
//	"foot -e sudo ncdu /" -> ("ncdu",      "sudo ncdu /",     true)
//	"firefox"             -> ("firefox",   "firefox",         false)
func splitCommand(cmd string) (bin string, inner string, needsTerm bool) {
	inner = cmd
	for _, p := range []string{popupTerm, termPrefix} {
		if strings.HasPrefix(cmd, p) {
			inner = strings.TrimSpace(strings.TrimPrefix(cmd, p))
			needsTerm = true
			break
		}
	}

	fields := strings.Fields(inner)
	if len(fields) == 0 {
		return "", inner, needsTerm
	}
	bin = fields[0]

	// "sudo ncdu /" -> el binario que importa es ncdu, no sudo.
	if bin == "sudo" {
		for _, f := range fields[1:] {
			if strings.HasPrefix(f, "-") || strings.Contains(f, "=") {
				continue
			}
			bin = f
			break
		}
	}
	return bin, inner, needsTerm
}

// binaryExists resuelve el binario en el PATH y en los directorios estándar.
func binaryExists(bin string) bool {
	if bin == "" {
		return false
	}
	if strings.ContainsRune(bin, '/') {
		st, err := os.Stat(bin)
		return err == nil && !st.IsDir()
	}
	if _, err := exec.LookPath(bin); err == nil {
		return true
	}
	for _, dir := range []string{"/usr/local/bin", "/usr/bin", "/bin", "/usr/sbin", "/sbin"} {
		if st, err := os.Stat(filepath.Join(dir, bin)); err == nil && !st.IsDir() {
			return true
		}
	}
	return false
}

// shellQuote envuelve s en comillas simples seguras para bash.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// wrapTerminal abre el comando en una terminal flotante y centrada, y la
// deja abierta si falla para que el error quede a la vista.
func wrapTerminal(inner string) string {
	script := inner + `; rc=$?; if [ $rc -ne 0 ]; then echo; echo "[exit $rc] Enter para cerrar"; read -r; fi`
	return popupTerm + "/bin/bash -lc " + shellQuote(script)
}

// showError abre el popup con el mensaje de error y espera Enter.
func showError(msg string) {
	script := "printf '%s\\n' " + shellQuote(msg) + `; echo; echo "Enter para cerrar"; read -r`
	spawn(popupTerm + "/bin/bash -lc " + shellQuote(script))
}

// executeCommand valida y lanza un item del menú.
func executeCommand(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		logf("skip empty command")
		return
	}

	bin, inner, needsTerm := splitCommand(cmd)
	logf("execute: cmd=%q bin=%q term=%v", cmd, bin, needsTerm)

	if !binaryExists(bin) {
		logf("missing binary %q for %q", bin, cmd)
		showError(fmt.Sprintf("%s no esta instalado.\n\nComando: %s", bin, cmd))
		return
	}

	if needsTerm {
		spawn(wrapTerminal(inner))
		return
	}
	spawn(inner)
}

// spawn lanza el comando en la sesión Wayland del usuario alarm.
func spawn(cmd string) {
	shell := "/bin/bash"
	if _, err := exec.LookPath("bash"); err != nil {
		shell = "/bin/sh"
	}

	c := exec.Command("sudo", "-u", "alarm", "-H",
		"WAYLAND_DISPLAY=wayland-1",
		"XDG_RUNTIME_DIR=/run/user/1000",
		"DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus",
		"HOME=/home/alarm",
		"PATH=/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin",
		"LANG=C.UTF-8",
		"LC_ALL=C.UTF-8",
		"XDG_CURRENT_DESKTOP=sway",
		"XDG_SESSION_TYPE=wayland",
		"XDG_SESSION_DESKTOP=sway",
		"USER=alarm",
		"LOGNAME=alarm",
		"SHELL=/bin/bash",
		"TERM=foot-extra",
		shell, "-c", "cd /home/alarm && exec "+cmd,
	)
	// El menu corre como root con CWD /root. foot hereda ese CWD y falla con
	// "failed to change working directory to /root: Permission denied" al
	// bajar privilegios a alarm, matando la terminal al instante.
	c.Dir = "/home/alarm"
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := c.Start(); err != nil {
		logf("start err: %v", err)
		return
	}
	go func() {
		if err := c.Wait(); err != nil {
			logf("child exit err: %v", err)
		} else {
			logf("child exit ok")
		}
	}()
}
