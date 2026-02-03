#!/bin/bash
set -e

# Configurações
GROUP_NAME="linux-updater"
DATA_DIR="/var/lib/updater"
BIN_DIR="/usr/local/bin"

echo ">>> 🔒 Instalando Linux Updater (Versão Completa com Persistência)..."

# 0. Verificações Prévias
if [ "$EUID" -ne 0 ]; then
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

# Adiciona o usuário atual ao grupo
usermod -a -G $GROUP_NAME $REAL_USER
echo "Usuário '$REAL_USER' adicionado ao grupo '$GROUP_NAME'."

# 2. Parar Serviço Antigo
echo ">>> 2. Parando serviço antigo..."
systemctl stop updater-daemon.service 2>/dev/null || true

# 3. Instalar Binários
echo ">>> 3. Copiando binários..."
mkdir -p $BIN_DIR

# Verifica se a pasta bin existe (compatibilidade com o build.sh)
if [ -d "bin" ]; then
    cp bin/updater-daemon $BIN_DIR/
    cp bin/updater-client $BIN_DIR/
else
    echo "⚠️  Pasta 'bin/' não encontrada no diretório atual."
    echo "Tentando copiar binários da raiz (caso esteja rodando manualmente)..."
    cp updater-daemon $BIN_DIR/ 2>/dev/null || true
    cp updater-client $BIN_DIR/ 2>/dev/null || true
fi

# Garante permissão de execução
chmod 755 $BIN_DIR/updater-daemon
chmod 755 $BIN_DIR/updater-client

# 4. Configurar Persistência de Dados (A PARTE NOVA)
echo ">>> 4. Configurando diretório de persistência ($DATA_DIR)..."
mkdir -p $DATA_DIR

# Define Root como dono, mas o Grupo como dono secundário
chown root:$GROUP_NAME $DATA_DIR

# Permissão 2775 (drwxrwsr-x)
# 2 (SetGID): Arquivos criados aqui herdam o grupo da pasta automaticamente
# 775: Root e Grupo podem escrever/ler/executar
chmod 2775 $DATA_DIR

# Cria o arquivo de metas vazio se não existir, para evitar erro na primeira leitura
if [ ! -f "$DATA_DIR/goals.json" ]; then
    echo '{"apps":{}}' > "$DATA_DIR/goals.json"
    # Permissões do arquivo: Root e Grupo podem ler e escrever
    chown root:$GROUP_NAME "$DATA_DIR/goals.json"
    chmod 0664 "$DATA_DIR/goals.json"
    echo "Arquivo goals.json criado."
fi

# 5. Instalar Serviço Systemd
echo ">>> 5. Configurando Systemd..."
if [ -f "updater-daemon.service" ]; then
    cp updater-daemon.service /etc/systemd/system/
    systemctl daemon-reload
    systemctl enable updater-daemon
    systemctl start updater-daemon
else
    echo "❌ Erro: Arquivo updater-daemon.service não encontrado!"
    exit 1
fi

# 6. Instalar Atalho Desktop
echo ">>> 6. Criando atalho no menu..."
if [ -f "updater-client.desktop" ]; then
    cp updater-client.desktop /usr/share/applications/
fi

echo "--------------------------------------------------------"
echo "✅ Instalação Concluída!"
echo "📂 Persistência: $DATA_DIR"
echo "🔌 Socket: /run/linux-updater/service.sock"
echo ""
echo "⚠️  MUITO IMPORTANTE:"
echo "Você precisa fazer LOGOUT e LOGIN novamente para que seu usuário"
echo "reconheça o novo grupo '$GROUP_NAME' e consiga acessar o Socket."
echo "--------------------------------------------------------"