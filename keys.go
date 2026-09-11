package main

import (
	"github.com/gotk3/gotk3/gtk"
)

// sameRow compara dos *gtk.ListBoxRow por su puntero nativo GTK (uintptr).
// gotk3 crea wrappers Go nuevos en cada llamada a GetRowAtIndex, por lo que
// comparar punteros Go falla aunque ambos representen el mismo objeto C.
func sameRow(a, b *gtk.ListBoxRow) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Native() == b.Native()
}

// isSelectableRow devuelve true si la fila es visible y seleccionable.
func isSelectableRow(row *gtk.ListBoxRow) bool {
	if row == nil || !row.GetVisible() {
		return false
	}
	return row.GetSelectable()
}

func moveSelection(list *gtk.ListBox, delta int) {
	cur := list.GetSelectedRow()
	curIdx := -1
	firstVisibleIdx := -1
	lastVisibleIdx := -1
	visible := 0
	for i := 0; i < 100000; i++ {
		row := list.GetRowAtIndex(i)
		if row == nil {
			break
		}
		if sameRow(row, cur) {
			curIdx = i
		}
		if isSelectableRow(row) {
			visible++
			lastVisibleIdx = i
			if firstVisibleIdx < 0 {
				firstVisibleIdx = i
			}
		}
	}
	if visible == 0 {
		return
	}

	target := curIdx
	if delta > 0 {
		if cur == nil {
			target = firstVisibleIdx - 1
		}
		for step := 0; step < delta; step++ {
			found := false
			for i := target + 1; i < 100000; i++ {
				row := list.GetRowAtIndex(i)
				if row == nil {
					break
				}
				if isSelectableRow(row) {
					target = i
					found = true
					break
				}
			}
			if !found {
				break
			}
		}
	} else if delta < 0 {
		if cur == nil {
			target = lastVisibleIdx + 1
		}
		for step := 0; step < -delta; step++ {
			found := false
			for i := target - 1; i >= 0; i-- {
				row := list.GetRowAtIndex(i)
				if row == nil {
					break
				}
				if isSelectableRow(row) {
					target = i
					found = true
					break
				}
			}
			if !found {
				break
			}
		}
	}

	if target >= 0 {
		row := list.GetRowAtIndex(target)
		if row != nil {
			list.SelectRow(row)
			scrollToRow(list, row)
		}
	}
}

// scrollToRow asegura que la fila seleccionada quede visible.
//
// El vadjustment pertenece al ScrolledWindow, no al ListBox: pedírselo al
// ListBox devuelve nil y el scroll nunca se movía.
//
// La posición sale de la geometría real de la fila (GetAllocation) en vez de
// asumir alto uniforme, porque los labels con ellipsize no miden todos igual.
func scrollToRow(list *gtk.ListBox, row *gtk.ListBoxRow) {
	if list == nil || row == nil {
		return
	}
	parent, err := list.GetParent()
	if err != nil || parent == nil {
		return
	}
	sw, ok := parent.(*gtk.ScrolledWindow)
	if !ok {
		return
	}
	adj := sw.GetVAdjustment()
	if adj == nil {
		return
	}

	alloc := row.GetAllocation()
	rowTop := float64(alloc.GetY())
	rowBottom := rowTop + float64(alloc.GetHeight())

	val := adj.GetValue()
	page := adj.GetPageSize()

	switch {
	case rowTop < val:
		// La fila quedó por encima del viewport.
		adj.SetValue(rowTop)
	case rowBottom > val+page:
		// La fila quedó por debajo: alinearla contra el borde inferior.
		target := rowBottom - page
		if max := adj.GetUpper() - page; target > max {
			target = max
		}
		adj.SetValue(target)
	}
}

func selectFirstVisible(list *gtk.ListBox) {
	for i := 0; i < 100000; i++ {
		row := list.GetRowAtIndex(i)
		if row == nil {
			return
		}
		if isSelectableRow(row) {
			list.SelectRow(row)
			return
		}
	}
}

func selectLastVisible(list *gtk.ListBox) {
	var last *gtk.ListBoxRow
	for i := 0; i < 100000; i++ {
		row := list.GetRowAtIndex(i)
		if row == nil {
			break
		}
		if isSelectableRow(row) {
			last = row
		}
	}
	if last != nil {
		list.SelectRow(last)
	}
}
