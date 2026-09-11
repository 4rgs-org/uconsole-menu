package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MenuNode es un nodo del árbol del menú.
// Si Action == "" y Children != nil, es un submenú.
// Si Action != "" es un item final.
type MenuNode struct {
	Icon     string
	Label    string
	Action   string
	Children []*MenuNode
}

func (n *MenuNode) IsSubmenu() bool {
	return n.Action == "" && len(n.Children) > 0
}

// SearchText devuelve el texto searchable para la búsqueda fuzzy local.
func (n *MenuNode) SearchText() string {
	return strings.ToLower(n.Icon + " " + n.Label)
}

// BuildMenu arma el árbol raíz del menú. Categorías principales como submenús.
func BuildMenu() *MenuNode {
	root := &MenuNode{Icon: " ", Label: "uconsole-menu"}

	root.Children = append(root.Children, &MenuNode{
		Icon:   " ",
		Label:  "Search packages",
		Action: "uconsole-installer",
	})
	root.Children = append(root.Children, buildSystemMenu())
	root.Children = append(root.Children, buildAppsMenu())
	root.Children = append(root.Children, buildOmarchyMenu())
	root.Children = append(root.Children, buildInstallersMenu())
	root.Children = append(root.Children, buildBinariesMenu())
	return root
}

func buildSystemMenu() *MenuNode {
	n := &MenuNode{Icon: " ", Label: "System"}
	n.Children = []*MenuNode{
		{Icon: " ", Label: "btop", Action: "foot -e btop"},
		{Icon: " ", Label: "Logs", Action: "foot -e sudo journalctl -f"},
		{Icon: " ", Label: "Network", Action: "foot -e sudo nmtui"},
		{Icon: " ", Label: "Wallpaper", Action: "uconsole-wallpaper"},
		{Icon: " ", Label: "Keybindings", Action: "omarchy-menu-keybindings"},
		{Icon: " ", Label: "Modifier key", Action: "foot --app-id=uconsole-popup -e uconsole-modifier"},
		{Icon: " ", Label: "Lock", Action: "omarchy-apply-lock"},
		{Icon: " ", Label: "Reboot", Action: "systemctl reboot"},
		{Icon: " ", Label: "Shutdown", Action: "systemctl poweroff"},
		{Icon: " ", Label: "Hardware", Action: "omarchy-apply-hardware"},
	}
	return n
}

func buildAppsMenu() *MenuNode {
	n := &MenuNode{Icon: " ", Label: "Apps"}
	n.Children = []*MenuNode{
		{Icon: " ", Label: "Terminal", Action: "foot"},
		{Icon: " ", Label: "Editor (nano)", Action: "foot -e nano"},
		{Icon: " ", Label: "Clipboard", Action: "omarchy-menu-clipboard"},
		{Icon: " ", Label: "Emoji", Action: "omarchy-menu-emoji"},
		{Icon: " ", Label: "Files", Action: "omarchy-menu-file"},
		{Icon: " ", Label: "Screenshot", Action: "omarchy-capture-screenshot"},
	}
	return n
}

func buildOmarchyMenu() *MenuNode {
	n := &MenuNode{Icon: " ", Label: "Omarchy"}
	n.Children = []*MenuNode{
		{Icon: " ", Label: "Update", Action: "foot -e omarchy-update"},
		{Icon: " ", Label: "Update status", Action: "foot -e omarchy-update-status"},
		{Icon: " ", Label: "Install package", Action: "foot -e omarchy-pkg-install"},
		{Icon: " ", Label: "Remove package", Action: "foot -e omarchy-pkg-remove"},
		{Icon: " ", Label: "Theme select", Action: "omarchy-theme-set"},
		{Icon: " ", Label: "Theme list", Action: "foot -e omarchy-theme-list"},
		{Icon: " ", Label: "Font select", Action: "omarchy-font-set"},
	}
	return n
}

func buildInstallersMenu() *MenuNode {
	n := &MenuNode{Icon: " ", Label: "Install"}
	n.Children = parseInstallers()
	return n
}

func parseInstallers() []*MenuNode {
	var out []*MenuNode
	matches, err := filepath.Glob("/usr/local/bin/omarchy-install-*")
	if err != nil {
		logf("glob err: %v", err)
		return out
	}
	type pair struct{ name, summary string }
	seen := map[string]bool{}
	var pairs []pair
	for _, m := range matches {
		base := filepath.Base(m)
		name := strings.TrimPrefix(base, "omarchy-install-")
		if seen[name] {
			continue
		}
		seen[name] = true
		summary := readOmarchyHeader(m, "summary")
		if summary == "" {
			summary = name
		}
		pairs = append(pairs, pair{name, summary})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].summary < pairs[j].summary
	})
	for _, p := range pairs {
		// Los installers son scripts interactivos (pacman/AUR): van en terminal.
		out = append(out, &MenuNode{
			Icon:   " ",
			Label:  p.summary,
			Action: "foot -e omarchy-install-" + p.name,
		})
	}
	return out
}

// buildBinariesMenu lista todos los ejecutables instalados en el PATH usado
// por uconsole-menu. Se deduplican por nombre y se ordenan alfabéticamente.
//
// Cada binario se abre en el popup foot: es la opción segura para CLIs y los
// comandos gráficos también pueden conectar a Wayland desde esa misma sesión.
func buildBinariesMenu() *MenuNode {
	n := &MenuNode{Icon: " ", Label: "Binaries"}
	seen := map[string]bool{}
	dirs := []string{
		"/home/alarm/.local/bin",
		"/usr/local/bin",
		"/usr/bin",
		"/usr/local/sbin",
		"/usr/sbin",
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			if seen[name] || strings.HasPrefix(name, ".") {
				continue
			}
			path := filepath.Join(dir, name)
			info, err := os.Stat(path)
			if err != nil || info.IsDir() || info.Mode()&0111 == 0 {
				continue
			}
			seen[name] = true
			n.Children = append(n.Children, &MenuNode{
				Icon:   " ",
				Label:  name,
				Action: "foot -e " + path,
			})
		}
	}

	sort.Slice(n.Children, func(i, j int) bool {
		return n.Children[i].Label < n.Children[j].Label
	})
	logf("loaded %d installed binaries", len(n.Children))
	return n
}

func readOmarchyHeader(path, key string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	prefix := "# omarchy:" + key + "="
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			val := strings.TrimPrefix(trimmed, prefix)
			return strings.TrimSpace(val)
		}
	}
	return ""
}
