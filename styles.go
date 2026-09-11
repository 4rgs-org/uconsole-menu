package main

import (
	"fmt"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

// applyStyles registra el CSS de Omarchy para screen.
// Estilo: card angosto, fondo sólido, borde cuadrado, texto gris claro,
// iconos a la izquierda, sin descripciones, submenús jerárquicos.
func applyStyles(screen *gdk.Screen, theme Theme) {
	if screen == nil {
		return
	}
	provider, err := gtk.CssProviderNew()
	if err != nil {
		logf("CssProviderNew err: %v", err)
		return
	}
	css := buildCSS(theme)
	if err := provider.LoadFromData(css); err != nil {
		lines := strings.Split(css, "\n")
		for i := 96; i < len(lines) && i < 108; i++ { logf("CSS %d: %s", i+1, lines[i]) }
		logf("LoadFromData err: %v", err)
		return
	}
	gtk.AddProviderForScreen(screen, provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func buildCSS(t Theme) string {
	bg := FallbackColor(t.Background, "#1a1b26")
	fg := FallbackColor(t.Foreground, "#a9b1d6")
	accent := FallbackColor(t.Accent, "#7aa2f7")
	muted := FallbackColor(t.Muted, "#414868")

	// Las filas comparten el fondo sólido. El foco se indica sólo con una
	// sombra/borde oscuro, sin alterar el color del item seleccionado.
	focusShadow := hexToRGBA(t.DarkerBackground, 0.95)

	var b strings.Builder
	fmt.Fprintf(&b, `
window#menu-window {
    background-color: %s;
}
#card {
    background-color: %s;
    color: %s;
    border: none;
    border-radius: 0;
    padding: 0;
}

#header {
    background-color: %s;
    color: %s;
    font-family: "JetBrainsMono Nerd Font", monospace;
    font-weight: bold;
    padding: 12px 14px 10px 14px;
    border-bottom: 1px solid %s;
}
#header-title {
    color: %s;
    font-size: 15px;
    font-family: "JetBrainsMono Nerd Font", monospace;
}

#search {
    background-color: %s;
    padding: 8px 12px 6px 12px;
    border-bottom: 1px solid %s;
}
#search entry {
    background-color: %s;
    color: %s;
    border: 1px solid %s;
    border-radius: 0;
    padding: 6px 10px;
    font-family: "JetBrainsMono Nerd Font", monospace;
    font-size: 13px;
}
#search entry:focus {
    border-color: %s;
}
#search prompt {
    color: %s;
    font-weight: bold;
    padding-right: 6px;
}

#list-scroll {
    background-color: %s;
    padding: 0;
}
#list {
    background-color: %s;
}
#list list { background-color: %s; }
#list list row { padding: 0; border-radius: 0; background-color: %s; }

#row {
    background-color: transparent;
    padding: 0;
}
#row-box {
    padding: 8px 12px 8px 6px;
}
#row-icon {
    color: %s;
    font-family: "JetBrainsMono Nerd Font", monospace;
    font-size: 15px;
    min-width: 32px;
}
#row-label {
    color: %s;
    font-family: "JetBrainsMono Nerd Font", monospace;
    font-size: 13px;
    font-weight: normal;
}
#row-chevron {
    color: %s;
    font-family: "JetBrainsMono Nerd Font", monospace;
    font-size: 13px;
    min-width: 14px;
}
#row.is-submenu {
    background-color: transparent;
}
#row:selected {
    background-color: %s;
    border: 1px solid %s;
    box-shadow: inset 0 0 0 1px %s;
}
#row:selected #row-label,
#row:selected #row-icon,
#row:selected #row-chevron {
    color: %s;
}
#row:hover:not(:selected) {
    background-color: transparent;
}

`,
		bg,
		bg, fg,
		bg, fg, muted,
		fg,
		bg, muted,
		bg, fg, muted,
		accent,
		fg,
		bg,
		bg,
		bg,
		bg,
		fg,
		fg,
		muted,
		focusShadow, focusShadow, focusShadow,
		fg,
	)
	return b.String()
}
