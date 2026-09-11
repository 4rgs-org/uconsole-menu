# uconsole-menu

>Native GTK3 launcher for the uConsole, Omarchy-style.

Replacement for wofi/Rofi as the default application launcher in Sway. Built with Go + gotk3, reads the live Omarchy theme palette at startup, renders a card with the active colors, and supports nested submenus, global search across all installed binaries, and `app_id` rules for floating terminals.

## Demo

```
┌── uConsole ──
│  Search packages
│  System    ›
│  Apps      ›
│  Omarchy   ›
│  Install   ›
│  Binaries  ›
└──────────
```

## Dependencies

- GTK 3.24, Sway/Wayland
- `bash` o `sh` para el `install.sh`
- glibc estándar

## Install

Una línea, vía curl al instalador del repo:

```sh
curl -sSL https://raw.githubusercontent.com/4rgs-org/uconsole-menu/main/install.sh | sh
```


Para una versión específica:

```sh
curl -sSL https://raw.githubusercontent.com/4rgs-org/uconsole-menu/main/install.sh | sh -s -- --tag v0.1.0
```

Para instalar en otro directorio (útil para `~/.local/bin` sin root):

```sh
INSTALL_DIR=$HOME/.local/bin REPO=uconsole-menu bash install.sh
```

El instalador:
- Detecta la arquitectura con `uname -m` (aarch64, x86_64, armv7).
- Baja el tarball de la release seleccionada.
- Verifica `SHA256SUMS` antes de extraer.
- Copia el binario a `/usr/local/bin/uconsole-menu` (o el destino elegido).

## Post-install

Tras instalar el binario:

1. Crear el wrapper de sesión que exporta Wayland:

   ```sh
   sudo tee /home/alarm/.local/bin/uconsole-menu >/dev/null <<'EOF'
   #!/bin/bash
   exec /usr/local/bin/uconsole-menu
   EOF
   sudo chmod 755 /home/alarm/.local/bin/uconsole-menu
   ```

2. Vincularlo a Sway:

   ```ini
   # ~/.config/sway/config
   set $mod Mod4
   bindsym $mod+space exec /home/alarm/.local/bin/uconsole-menu
   for_window [app_id="uconsole-menu"] border none
   for_window [app_id="uconsole-popup"] floating enable, resize set 900 500, move position center
   ```

3. Recargar Sway:

   ```sh
   swaymsg reload
   ```

## Build from source

```sh
git clone https://github.com/4rgs-org/uconsole-menu
cd uconsole-menu
go build -o uconsole-menu .
sudo install -m 0755 uconsole-menu /usr/local/bin/uconsole-menu
```

## Usage

| Acción | Resultado |
| --- | --- |
| ``/home/alarm/.local/bin/uconsole-menu`` | Wrapper bash que exporta Wayland e invoca el binario. |
| ``Mod4+Space` (Sway)` | Abre el menú. Cambiable desde *System → Modifier key*. |
| `Arrow Up / Down` | Mueve la selección en la lista. |
| ``Enter`` | Abre un submenú o ejecuta el item resaltado. |
| `Backspace / `←`` | Vuelve al submenú padre. |
| ``Esc`` | Cierra el menú sin ejecutar nada. |

## License

MIT — ver [LICENSE](LICENSE).

## Liability

Software is provided as-is. The maintainers are not responsible for loss of work or data caused by accidental command execution. Always confirm the menu entry before pressing Enter.
