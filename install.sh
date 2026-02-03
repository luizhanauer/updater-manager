#!/bin/sh
set -e

# CORREÇÃO 1: Garante que o script rode dentro da pasta onde ele está
cd "$(dirname "$0")"

GROUP_NAME="linux-updater"
DATA_DIR="/var/lib/updater"
BIN_DIR="/usr/local/bin"

echo ">>> 🔒 Instalando Linux Updater (Versão Completa)..."

# CORREÇÃO 2: Compatibilidade com 'sh' (substitui $EUID)
if [ "$(id -u)" -ne 0 ]; then
  echo "Erro: Execute como root (sudo ./install.sh)"
  exit 1
fi

REAL_USER=$SUDO_USER
if [ -z "$REAL_USER" ]; then
  echo "Erro: Não foi possível detectar o usuário real."
  exit 1
fi

# 1. Criação de Grupo de Segurança
echo ">>> 1. Configurando grupos..."
if ! getent group $GROUP_NAME > /dev/null; then
  groupadd $GROUP_NAME
  echo "Grupo '$GROUP_NAME' criado."
fi

usermod -a -G $GROUP_NAME $REAL_USER
echo "Usuário '$REAL_USER' adicionado ao grupo '$GROUP_NAME'."

# 2. Parar Serviço Antigo
echo ">>> 2. Parando serviço antigo..."
systemctl stop updater-daemon.service 2>/dev/null || true

# 3. Instalar Binários
echo ">>> 3. Copiando binários..."
mkdir -p $BIN_DIR

# Agora isso vai funcionar porque demos 'cd' no início
if [ -d "bin" ]; then
    cp bin/updater-daemon $BIN_DIR/
    cp bin/updater-client $BIN_DIR/
else
    echo "❌ Erro Crítico: Pasta 'bin/' não encontrada em $(pwd)"
    exit 1
fi

chmod 755 $BIN_DIR/updater-daemon
chmod 755 $BIN_DIR/updater-client

# 4. Configurar Persistência de Dados
echo ">>> 4. Configurando persistência ($DATA_DIR)..."
mkdir -p $DATA_DIR
chown root:$GROUP_NAME $DATA_DIR
chmod 2775 $DATA_DIR

if [ ! -f "$DATA_DIR/goals.json" ]; then
    echo '{"apps":{}}' > "$DATA_DIR/goals.json"
    chown root:$GROUP_NAME "$DATA_DIR/goals.json"
    chmod 0664 "$DATA_DIR/goals.json"
fi

# 5. Instalar Serviço Systemd
echo ">>> 5. Configurando Systemd..."
cp updater-daemon.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable updater-daemon
systemctl start updater-daemon

# 6. Instalar Atalho Desktop
echo ">>> 6. Criando atalho..."
if [ -f "updater-client.desktop" ]; then
    cp updater-client.desktop /usr/share/applications/
fi

echo "--------------------------------------------------------"
echo "✅ Instalação Concluída com Sucesso!"
echo "--------------------------------------------------------"