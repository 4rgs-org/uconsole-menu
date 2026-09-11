package main

import (
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// navbar mantiene la posición en el árbol del menú.
type navbar struct {
	root  *MenuNode
	stack []*MenuNode // path desde la raíz
}

func (n *navbar) current() *MenuNode {
	if len(n.stack) == 0 {
		return n.root
	}
	return n.stack[len(n.stack)-1]
}

func (n *navbar) canGoBack() bool {
	return len(n.stack) > 0
}

func (n *navbar) push(node *MenuNode) {
	n.stack = append(n.stack, node)
}

func (n *navbar) pop() {
	if len(n.stack) == 0 {
		return
	}
	n.stack = n.stack[:len(n.stack)-1]
}

// uiState mantiene los widgets vivos durante navegaciones.
type visibleItem struct {
	node *MenuNode
	path []*MenuNode // submenús padre desde la raíz hasta el nodo
}

type uiState struct {
	theme Theme
	nav   *navbar

	win        *gtk.Window
	list       *gtk.ListBox
	listScroll *gtk.ScrolledWindow
	headerTitle *gtk.Label
	searchBox  *gtk.Box
	entry      *gtk.SearchEntry

	// visible[i] corresponde a la fila i del ListBox tras el último refreshList.
	// path permite activar correctamente resultados encontrados en submenús.
	visible []visibleItem
}

func createMainWindow(theme Theme, root *MenuNode) *gtk.Window {
	ui := &uiState{
		theme: theme,
		nav:   &navbar{root: root},
	}
	ui.build()
	return ui.win
}

func (ui *uiState) build() {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		logf("WindowNew err: %v", err)
		return
	}
	win.SetTitle("uConsole")
	win.SetDefaultSize(320, 200)
	win.SetResizable(false)
	win.SetDecorated(false)
	win.SetKeepAbove(true)
	win.SetSkipTaskbarHint(true)
	win.SetSkipPagerHint(true)
	win.SetTypeHint(gdk.WINDOW_TYPE_HINT_DIALOG)
	win.SetPosition(gtk.WIN_POS_CENTER)
	win.SetName("menu-window")
	ui.win = win

	card, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	card.SetName("card")
	win.Add(card)

	headerBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	headerBox.SetName("header")
	pixbuf, err := gdk.PixbufNewFromFileAtScale("/usr/local/share/uconsole/clockworkpi.png", 10, 10, true)
	if err == nil {
		logo, imageErr := gtk.ImageNewFromPixbuf(pixbuf)
		if imageErr == nil {
			headerBox.Add(logo)
		}
	}
	ui.headerTitle, _ = gtk.LabelNew("uConsole")
	ui.headerTitle.SetName("header-title")
	ui.headerTitle.SetHAlign(gtk.ALIGN_START)
	headerBox.Add(ui.headerTitle)
	card.Add(headerBox)

	ui.entry = buildSearch()
	ui.searchBox = buildSearchBox(ui.entry)
	card.Add(ui.searchBox)

	ui.list, ui.listScroll = buildList()
	card.Add(ui.listScroll)

	win.Connect("destroy", func() {
		logf("window destroy")
		gtk.MainQuit()
	})

	ui.entry.Connect("search-changed", func() {
		ui.refreshList()
	})
	ui.entry.Connect("activate", func(e *gtk.SearchEntry) {
		ui.activateSelected()
	})

	win.Connect("key-press-event", func(w *gtk.Window, ev *gdk.Event) bool {
		key := gdk.EventKeyNewFromEvent(ev)
		if key == nil {
			return false
		}
		switch key.KeyVal() {
		case gdk.KEY_Escape:
			logf("Esc -> close")
			w.Close()
			return true
		}
		return false
	})

	ui.entry.Connect("key-press-event", func(e *gtk.SearchEntry, ev *gdk.Event) bool {
		key := gdk.EventKeyNewFromEvent(ev)
		if key == nil {
			return false
		}
		switch key.KeyVal() {
		case gdk.KEY_Down:
			moveSelection(ui.list, 1)
			return true
		case gdk.KEY_Up:
			moveSelection(ui.list, -1)
			return true
		case gdk.KEY_Page_Down:
			moveSelection(ui.list, 6)
			return true
		case gdk.KEY_Page_Up:
			moveSelection(ui.list, -6)
			return true
		case gdk.KEY_Home:
			selectFirstVisible(ui.list)
			return true
		case gdk.KEY_End:
			selectLastVisible(ui.list)
			return true
		case gdk.KEY_Left:
			if ui.searchNonEmpty() {
				return false
			}
			if ui.nav.canGoBack() {
				ui.goBack()
				return true
			}
		case gdk.KEY_BackSpace:
			if ui.searchNonEmpty() {
				return false
			}
			if ui.nav.canGoBack() {
				ui.goBack()
				return true
			}
		}
		return false
	})

	ui.list.Connect("row-activated", func(l *gtk.ListBox, row *gtk.ListBoxRow) {
		ui.activateRow(row)
	})
	ui.list.Connect("key-press-event", func(l *gtk.ListBox, ev *gdk.Event) bool {
		key := gdk.EventKeyNewFromEvent(ev)
		if key == nil {
			return false
		}
		switch key.KeyVal() {
		case gdk.KEY_Return, gdk.KEY_KP_Enter:
			ui.activateSelected()
			return true
		case gdk.KEY_Right:
			ui.activateSelected()
			return true
		case gdk.KEY_Left:
			if ui.nav.canGoBack() {
				ui.goBack()
				return true
			}
		case gdk.KEY_Escape:
			ui.win.Close()
			return true
		}
		return false
	})

	ui.refreshList()
	win.ShowAll()
	logf("window shown")
}

func (ui *uiState) searchNonEmpty() bool {
	t, _ := ui.entry.GetText()
	return len(t) > 0
}

func (ui *uiState) refreshList() {
	cur := ui.nav.current()
	query, _ := ui.entry.GetText()
	query = strings.ToLower(strings.TrimSpace(query))

	ui.visible = ui.visible[:0]
	for {
		c := ui.list.GetRowAtIndex(0)
		if c == nil {
			break
		}
		ui.list.Remove(c)
	}

	if query == "" {
		// Sin búsqueda conserva el comportamiento normal del submenú actual.
		for _, child := range cur.Children {
			addRow(ui.list, child, "")
			ui.visible = append(ui.visible, visibleItem{node: child})
		}
	} else {
		// Con búsqueda recorre todo el árbol, incluyendo todos los binarios.
		var walk func(*MenuNode, []*MenuNode)
		walk = func(parent *MenuNode, path []*MenuNode) {
			for _, child := range parent.Children {
				childPath := append(append([]*MenuNode{}, path...), child)
				if child.IsSubmenu() {
					walk(child, childPath)
					continue
				}
				if !strings.Contains(strings.ToLower(child.SearchText()), query) {
					continue
				}
				breadcrumb := ""
				if len(path) > 0 {
					labels := make([]string, len(path))
					for i, p := range path {
						labels[i] = p.Label
					}
					breadcrumb = strings.Join(labels, " › ")
				}
				addRow(ui.list, child, breadcrumb)
				ui.visible = append(ui.visible, visibleItem{node: child, path: path})
			}
		}
		walk(ui.nav.root, nil)
	}
	added := len(ui.visible)

	// El encabezado queda fijo: las rutas y resultados ya se reflejan en la lista.
	ui.headerTitle.SetText("uConsole")

	if added > 0 {
		row := ui.list.GetRowAtIndex(0)
		if row != nil {
			ui.list.SelectRow(row)
		}
	} else {
		logf("refreshList: 0 items visible (cur=%q, query=%q)", cur.Label, query)
	}

	// El menú se adapta al contenido: compacto con pocos resultados y scroll
	// sólo cuando el submenú o búsqueda supera el alto máximo.
	const headerAndSearch = 92
	const rowHeight = 41
	const maxRows = 10
	rows := added
	if rows < 1 {
		rows = 1
	}
	if rows > maxRows {
		rows = maxRows
	}
	ui.listScroll.SetSizeRequest(320, rows*rowHeight)
	ui.win.Resize(320, headerAndSearch+rows*rowHeight)
	// Resize debe ejecutarse después de que GTK recalcule el layout; de lo
	// contrario el tamaño natural del ScrolledWindow vuelve a imponerse.
	glib.IdleAdd(func() bool {
		ui.win.Resize(320, headerAndSearch+rows*rowHeight)
		return false
	})
}

func (ui *uiState) goBack() {
	ui.nav.pop()
	ui.refreshList()
}

func (ui *uiState) activateSelected() {
	row := ui.list.GetSelectedRow()
	if row == nil {
		return
	}
	ui.activateRow(row)
}

func (ui *uiState) activateRow(row *gtk.ListBoxRow) {
	if row == nil {
		return
	}
	idx := row.GetIndex()
	if idx < 0 || idx >= len(ui.visible) {
		logf("activateRow: idx %d fuera de rango (%d visibles)", idx, len(ui.visible))
		return
	}
	item := ui.visible[idx]
	node := item.node
	if node.IsSubmenu() {
		logf("drill into %q (%d children)", node.Label, len(node.Children))
		ui.nav.push(node)
		ui.entry.SetText("")
		ui.refreshList()
		return
	}
	logf("activate: %q -> %s", node.Label, node.Action)
	executeCommand(node.Action)
	ui.win.Close()
}

// helpers -------------------------------------------------------------

func buildSearchBox(entry *gtk.SearchEntry) *gtk.Box {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	box.SetName("search")
	prompt, _ := gtk.LabelNew("❯")
	prompt.SetName("prompt")
	box.Add(prompt)
	box.Add(entry)
	return box
}

func buildSearch() *gtk.SearchEntry {
	e, _ := gtk.SearchEntryNew()
	e.SetPlaceholderText(" type to filter")
	return e
}

func buildList() (*gtk.ListBox, *gtk.ScrolledWindow) {
	scroll, _ := gtk.ScrolledWindowNew(nil, nil)
	scroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	scroll.SetName("list-scroll")
	scroll.SetSizeRequest(320, -1)
	list, _ := gtk.ListBoxNew()
	list.SetName("list")
	list.SetSelectionMode(gtk.SELECTION_SINGLE)
	scroll.Add(list)
	return list, scroll
}

func buildFooter() *gtk.Box {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	box.SetName("footer")
	type pair struct{ text, key string }
	parts := []pair{
		{"↑↓", "nav"},
		{"↵", "open"},
		{"⌫", "back"},
		{"Esc", "close"},
	}
	for i, p := range parts {
		if i > 0 {
			sep, _ := gtk.LabelNew("│")
			box.Add(sep)
		}
		k, _ := gtk.LabelNew("")
		k.SetMarkup("<tt>" + keycap(p.text) + "</tt>")
		box.Add(k)
		t, _ := gtk.LabelNew(p.key)
		box.Add(t)
	}
	return box
}

// keycap renderiza un keystroke como badge monochrome.
func keycap(s string) string {
	esc := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "&", "&amp;"), "<", "&lt;"), ">", "&gt;")
	return `<span foreground="#1a1b26" background="#7aa2f7"> ` + esc + ` </span>`
}

// addRow agrega una fila con icono, label y contexto opcional de búsqueda.
func addRow(list *gtk.ListBox, n *MenuNode, breadcrumb string) *gtk.ListBoxRow {
	row, _ := gtk.ListBoxRowNew()
	row.SetName("row")
	if n.IsSubmenu() {
		styleCtx, _ := row.GetStyleContext()
		styleCtx.AddClass("is-submenu")
	}

	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	box.SetName("row-box")

	icon, _ := gtk.LabelNew(n.Icon)
	icon.SetName("row-icon")
	icon.SetHAlign(gtk.ALIGN_CENTER)
	icon.SetSizeRequest(36, -1)
	box.Add(icon)

	label, _ := gtk.LabelNew(n.Label)
	label.SetName("row-label")
	label.SetHAlign(gtk.ALIGN_START)
	label.SetEllipsize(3)
	box.Add(label)

	right := breadcrumb
	if n.IsSubmenu() {
		right = "›"
	}
	context, _ := gtk.LabelNew(right)
	context.SetName("row-chevron")
	context.SetEllipsize(3)
	context.SetMaxWidthChars(18)
	box.PackEnd(context, false, false, 0)

	row.Add(box)
	list.Add(row)
	row.ShowAll()
	return row
}
