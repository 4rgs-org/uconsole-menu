#!/bin/sh
# uconsole-apps installer
# Uso:
#   curl -sSL https://raw.githubusercontent.com/4rgs-org/<repo>/main/install.sh | sh
#   curl -sSL ... | sh -s -- --tag v0.1.0   (por defecto: latest)
#
# Variables reconocidas:
#   REPO            nombre del repo (obligatorio)
#   BINARY          nombre del binario (por defecto: nombre del repo)
#   INSTALL_DIR     destino (por defecto: /usr/local/bin)
#   TAG             tag a bajar (por defecto: latest release)

set -e

REPO="${REPO:?REPO no seteada, ej: REPO=uconsole-menu}"
BINARY="${BINARY:-$REPO}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
TAG="${TAG:-}"

usage() {
  cat <<EOF
Uso: REPO=$REPO [BINARY=$BINARY] [INSTALL_DIR=$INSTALL_DIR] [TAG=<tag>]
EOF
}

# Parseo de flags opcionales tipo --tag vX.Y.Z
while [ $# -gt 0 ]; do
  case "$1" in
    --tag) TAG="$2"; shift 2 ;;
    --install-dir) INSTALL_DIR="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Argumento desconocido: $1" >&2; usage; exit 2 ;;
  esac
done

ARCH="$(uname -m)"
case "$ARCH" in
  aarch64) ASSET_ARCH="arm64" ;;
  x86_64)  ASSET_ARCH="x86_64"  ;;
  armv7l)  ASSET_ARCH="armv7"   ;;
  *)
    echo "Arquitectura no soportada: $ARCH" >&2
    exit 1
    ;;
esac

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux) ;;
  *) echo "Sistema operativo no soportado: $OS" >&2; exit 1 ;;
esac

# Resolver el tag. Si no se pasó, tomar el último release.
if [ -z "$TAG" ]; then
  if command -v curl >/dev/null 2>&1; then
    TAG=$(curl -fsSL "https://api.github.com/repos/4rgs-org/$REPO/releases/latest" \
      | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)
  fi
  if [ -z "$TAG" ]; then
    echo "No se pudo determinar el último release. Pasá TAG=vX.Y.Z" >&2
    exit 1
  fi
fi

ASSET="${BINARY}_${OS}_${ASSET_ARCH}.tar.gz"
URL="https://github.com/4rgs-org/${REPO}/releases/download/${TAG}/${ASSET}"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT INT TERM

echo "Descargando $ASSET desde $URL"
curl -fL --retry 3 -o "$tmp/$ASSET" "$URL"

echo "Verificando SHA256"
curl -fL --retry 3 -o "$tmp/${ASSET}.sha256" \
  "https://github.com/4rgs-org/${REPO}/releases/download/${TAG}/SHA256SUMS"

(cd "$tmp" && sha256sum -c --ignore-missing "${ASSET}.sha256") || {
  echo "Falla la verificación de suma de verificación" >&2
  exit 1
}

tar -xzf "$tmp/$ASSET" -C "$tmp"

if [ -w "$INSTALL_DIR" ]; then
  install -m 0755 "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
else
  echo "Se necesitan permisos para escribir en $INSTALL_DIR" >&2
  SUDO=""
  [ "$(id -u)" -ne 0 ] && SUDO="sudo "
  $SUDO install -m 0755 "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
fi

echo "$BINARY instalado en $INSTALL_DIR/$BINARY (versión $TAG)"
