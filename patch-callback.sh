#!/bin/bash
set -e
# Parchea gotk3 v0.6.4: gdk_since_3_22.go usa callback.Assign() sin importar
# el paquete, lo que rompe la compilación en cualquier host.
# El bug está documentado upstream; el fix es trivial.
DST="$(find / -path '*/gotk3@v0.6.4/gdk/gdk_since_3_22.go' 2>/dev/null | head -1)"
if [ -z "$DST" ] || [ ! -f "$DST" ]; then
  echo "patch-callback: no encontré gdk_since_3_22.go en el module cache"
  exit 0
fi
if grep -q '"github.com/gotk3/gotk3/internal/callback"' "$DST"; then
  echo "patch-callback: ya está aplicado en $DST"
  exit 0
fi
# Inserta el import después de "github.com/gotk3/gotk3/glib"
python3 - "$DST" <<'PY'
import sys
p = sys.argv[1]
with open(p) as f: s = f.read()
old = 'import (\n\t"unsafe"\n\n\t"github.com/gotk3/gotk3/glib"\n)\n'
new = 'import (\n\t"unsafe"\n\n\t"github.com/gotk3/gotk3/glib"\n\t"github.com/gotk3/gotk3/internal/callback"\n)\n'
if old in s:
    open(p, "w").write(s.replace(old, new, 1))
    print(f"patch-callback: aplicado en {p}")
else:
    print(f"patch-callback: bloque no encontrado en {p}, omitido")
PY
