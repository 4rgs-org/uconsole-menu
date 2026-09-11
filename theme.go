package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Theme struct {
	Name string

	Mode             string // dark | light
	Accent           string
	Selection        string
	Muted            string

	Background       string
	DarkBackground   string
	DarkerBackground string
	LighterBackground string

	Foreground       string
	DarkForeground   string
	LightForeground  string
	BrightForeground string

	Red, Yellow, Orange, Green, Cyan, Blue, Magenta, Brown string
	BrightRed, BrightYellow, BrightGreen, BrightCyan, BrightBlue string
}

func defaultTheme() Theme {
	t := Theme{
		Name: "default",
		Mode: "dark",

		Accent:    "#7aa2f7",
		Selection: "#292e42",
		Muted:     "#414868",

		Background:        "#1a1b26",
		DarkBackground:    "#13141c",
		DarkerBackground:  "#0e0e14",
		LighterBackground: "#24283b",

		Foreground:       "#c0caf5",
		DarkForeground:   "#565f89",
		LightForeground:  "#b4bee6",
		BrightForeground: "#c0caf5",

		Red:    "#f7768e",
		Yellow: "#e0af68",
		Orange: "#eb927b",
		Green:  "#9ece6a",
		Cyan:   "#449dab",
		Blue:   "#7aa2f7",
		Magenta: "#ad8ee6",
		Brown:  "#75493d",
	}
	return t
}

func LoadTheme() Theme {
	t := defaultTheme()

	if name := currentThemeName(); name != "" {
		t.Name = name
	}

	themeRoot := detectCurrentTheme()
	if themeRoot == "" {
		logf("LoadTheme: no theme root, using default")
		if t.Name == "" {
			t.Name = "default"
		}
		return t
	}
	if t.Name == "" {
		t.Name = filepath.Base(themeRoot)
	}

	colorsFile := filepath.Join(themeRoot, "colors.toml")
	f, err := os.Open(colorsFile)
	if err != nil {
		logf("LoadTheme: cannot open %s: %v", colorsFile, err)
		return t
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.Index(line, "=")
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		val = strings.Trim(val, "\"")
		if val == "" {
			continue
		}
		switch key {
		case "mode":
			t.Mode = val
		case "accent":
			t.Accent = val
		case "selection":
			t.Selection = val
		case "muted":
			t.Muted = val
		case "background":
			t.Background = val
		case "dark_background":
			t.DarkBackground = val
		case "darker_background":
			t.DarkerBackground = val
		case "lighter_background":
			t.LighterBackground = val
		case "foreground":
			t.Foreground = val
		case "dark_foreground":
			t.DarkForeground = val
		case "light_foreground":
			t.LightForeground = val
		case "bright_foreground":
			t.BrightForeground = val
		case "red":
			t.Red = val
		case "yellow":
			t.Yellow = val
		case "orange":
			t.Orange = val
		case "green":
			t.Green = val
		case "cyan":
			t.Cyan = val
		case "blue":
			t.Blue = val
		case "magenta":
			t.Magenta = val
		case "brown":
			t.Brown = val
		case "bright_red":
			t.BrightRed = val
		case "bright_yellow":
			t.BrightYellow = val
		case "bright_green":
			t.BrightGreen = val
		case "bright_cyan":
			t.BrightCyan = val
		case "bright_blue":
			t.BrightBlue = val
		}
	}
	logf("LoadTheme: %s (mode=%s, accent=%s, bg=%s)", t.Name, t.Mode, t.Accent, t.Background)
	return t
}

func detectCurrentTheme() string {
	candidates := []string{
		"/home/alarm/.local/state/omarchy/current/theme",
		"/home/alarm/.config/omarchy/current/theme",
		"/usr/share/omarchy/themes/current",
	}
	for _, p := range candidates {
		st, err := os.Stat(p)
		if err != nil || !st.IsDir() {
			continue
		}
		return p
	}
	name := os.Getenv("OMARCHY_THEME")
	if name != "" {
		p := filepath.Join("/usr/share/omarchy/themes", name)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	if name := currentThemeName(); name != "" {
		p := filepath.Join("/usr/share/omarchy/themes", name)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	return ""
}

func currentThemeName() string {
	candidates := []string{
		"/home/alarm/.local/state/omarchy/current/theme.name",
		"/home/alarm/.config/omarchy/current/theme.name",
	}
	for _, p := range candidates {
		if data, err := os.ReadFile(p); err == nil {
			n := strings.TrimSpace(string(data))
			if n != "" {
				return n
			}
		}
	}
	return ""
}

// FallbackColor devuelve val si no está vacío, si no fallback.
func FallbackColor(val, fallback string) string {
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	return val
}

// dim oscurece el color mezclándolo con el background.
// Simplificación: devuelve el input. El caller usa muted para oscurecer.
func (t Theme) Dim(c string) string { return c }

// hexToRGBA convierte "#rrggbb" en una cadena "rgba(r,g,b,a)" (CSS).
// Si no se puede, devuelve el input sin cambios.
func hexToRGBA(hex string, alpha float64) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return fmt.Sprintf("rgba(0,0,0,%.2f)", alpha)
	}
	r := hex[0:2]
	g := hex[2:4]
	b := hex[4:6]
	var rv, gv, bv int
	fmt.Sscanf(r, "%x", &rv)
	fmt.Sscanf(g, "%x", &gv)
	fmt.Sscanf(b, "%x", &bv)
	return fmt.Sprintf("rgba(%d,%d,%d,%.2f)", rv, gv, bv, alpha)
}
