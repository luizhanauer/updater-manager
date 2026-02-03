#!/bin/sh
set -e

REPO="luizhanauer/updater-manager"

# 1. Detecta Arquitetura
ARCH=$(uname -m)
case $ARCH in
    x86_64) ASSET_ARCH="amd64" ;;
    aarch64) ASSET_ARCH="arm64" ;;
    *) echo "❌ Arquitetura não suportada: $ARCH"; exit 1 ;;
esac

echo ">>> 📦 Instalador do Updater Manager"
echo ">>> Fonte: https://github.com/$REPO"

# 2. Verifica Dependências
if ! command -v curl >/dev/null; then echo "❌ Erro: 'curl' necessário."; exit 1; fi
if ! command -v tar >/dev/null; then echo "❌ Erro: 'tar' necessário."; exit 1; fi

# 3. Download
TMP_DIR=$(mktemp -d)
FILENAME="updater-manager_linux_${ASSET_ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/latest/download/${FILENAME}"

echo ">>> ⬇️  Baixando..."
if ! curl -f -L "$URL" -o "$TMP_DIR/$FILENAME"; then
    echo "❌ Erro ao baixar release."
    rm -rf "$TMP_DIR"
    exit 1
fi

echo ">>> 📂 Extraindo..."
tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"

echo ">>> 🚀 Instalando..."
# CORREÇÃO CRÍTICA: Entrar no diretório antes de rodar
cd "$TMP_DIR"
chmod +x install.sh

if [ "$(id -u)" -eq 0 ]; then
    sh ./install.sh
else
    # O sudo vai rodar o script no diretório atual ($TMP_DIR)
    sudo sh ./install.sh
fi

# Volta para onde estava e limpa
cd - > /dev/null
rm -rf "$TMP_DIR"
echo ">>> ✅ Setup finalizado."