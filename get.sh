#!/bin/sh
set -e

# SEU REPOSITÓRIO OFICIAL
REPO="luizhanauer/updater-manager"

# 1. Detecta Arquitetura
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ASSET_ARCH="amd64"
        ;;
    aarch64)
        ASSET_ARCH="arm64"
        ;;
    *)
        echo "❌ Arquitetura não suportada pelo instalador automático: $ARCH"
        exit 1
        ;;
esac

echo ">>> 📦 Instalador do Updater Manager"
echo ">>> Fonte: https://github.com/$REPO"

# 2. Verifica Dependências
if ! command -v curl >/dev/null; then
    echo "❌ Erro: 'curl' é necessário para baixar o instalador."
    exit 1
fi
if ! command -v tar >/dev/null; then
    echo "❌ Erro: 'tar' é necessário para descompactar."
    exit 1
fi

# 3. Prepara Download
TMP_DIR=$(mktemp -d)
FILENAME="updater-manager_linux_${ASSET_ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/latest/download/${FILENAME}"

echo ">>> ⬇️  Baixando a última versão ($ASSET_ARCH)..."
# -f: Falha se der 404
# -L: Segue redirecionamentos do GitHub
if ! curl -f -L "$URL" -o "$TMP_DIR/$FILENAME"; then
    echo "❌ Erro ao baixar: $URL"
    echo "Verifique se:"
    echo "  1. O repositório é público."
    echo "  2. Já existe uma Release publicada (não apenas Draft)."
    rm -rf "$TMP_DIR"
    exit 1
fi

echo ">>> 📂 Extraindo arquivos..."
# Extrai para a pasta temporária
tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"

echo ">>> 🚀 Executando instalação..."
INSTALLER="$TMP_DIR/install.sh"
chmod +x "$INSTALLER"

# Executa o install.sh (que já tem toda a lógica de criar pastas e services)
if [ "$(id -u)" -eq 0 ]; then
    sh "$INSTALLER"
else
    echo "🔒 Permissão de administrador necessária."
    sudo sh "$INSTALLER"
fi

# 4. Limpeza
rm -rf "$TMP_DIR"
echo ">>> ✅ Setup finalizado com sucesso."