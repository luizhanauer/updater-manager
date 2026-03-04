#  Updater Manager

![Updater Manager Screenshot](https://raw.githubusercontent.com/luizhanauer/updater-manager/assets/screenshot.png)

Um gerenciador de aplicativos centralizado para Linux. Fornece uma interface gráfica simples para instalar, atualizar e remover aplicativos de um catálogo remoto customizável, com um daemon robusto para gerenciar os pacotes em segundo plano.

O objetivo é simplificar a distribuição e manutenção de um conjunto de aplicativos em múltiplos sistemas, seja para uma equipe, uma empresa ou para uso pessoal.

---

## ✨ Funcionalidades

- **Interface Gráfica Simples:** Gerencie seus aplicativos com uma UI limpa e intuitiva construída com Wails e Vue.js.
- **Atualizações Automáticas:** Configure aplicativos para serem mantidos atualizados automaticamente em segundo plano.
- **Daemon Robusto:** Um serviço de sistema (`systemd`) cuida de todas as operações pesadas, como downloads e instalações.
- **Pré-verificação de Instalação:** Simula instalações para detectar problemas (como dependências quebradas) *antes* de modificar o sistema, evitando estados inconsistentes.
- **Catálogo Centralizado:** Gerencie a lista de aplicativos disponíveis a partir de um único arquivo `catalog.json` hospedado remotamente.
- **Segurança:** Usa permissões de grupo do Linux para garantir a comunicação segura entre a interface do usuário e o daemon de sistema.

---

## 🚀 Instalação

### Método Rápido (Recomendado)

Este comando irá baixar e executar o script `get.sh`, que detecta a arquitetura do seu sistema, baixa a última versão e executa o instalador `install.sh` com as permissões necessárias.

```bash
curl -sSL https://raw.githubusercontent.com/luizhanauer/updater-manager/main/get.sh | bash
```

> **⚠️ Importante!**
> Após a instalação, você **precisa fazer logout e login novamente** na sua sessão de usuário. Isso é necessário para que o sistema reconheça sua nova associação ao grupo `linux-updater` e permita que a interface gráfica se comunique com o daemon.

### Método Manual

1.  Vá para a página de Releases.
2.  Baixe o arquivo `.tar.gz` correspondente à sua arquitetura (ex: `updater-manager_linux_amd64.tar.gz`).
3.  Extraia o conteúdo e execute o script de instalação:

    ```bash
    # Nome do arquivo pode variar
    tar -xzf updater-manager_linux_amd64.tar.gz
    
    # Entre na pasta criada
    cd updater-manager_linux_amd64
    
    # Execute o instalador como root
    sudo ./install.sh
    ```
4.  Não se esqueça de fazer **logout e login** após a conclusão.

---

## 🛠️ Como Funciona

O projeto é dividido em três componentes principais:

1.  **Daemon (`updater-daemon`)**
    - Um serviço Go que roda em segundo plano, gerenciado pelo `systemd`.
    - É responsável por todas as operações de pacotes: verificar versões, baixar, simular e executar instalações/remoções (`.deb`).
    - Expõe um Unix Socket em `/run/linux-updater/service.sock` para comunicação.

2.  **Client (`updater-client`)**
    - Uma aplicação de desktop construída com Wails (Go + Vue.js).
    - Fornece a interface para o usuário visualizar o catálogo, instalar/remover aplicativos e configurar atualizações automáticas.
    - Comunica-se com o Daemon através do Unix Socket.

3.  **Registry (`catalog.json`)**
    - Um arquivo JSON simples hospedado em um repositório Git.
    - Funciona como a "fonte da verdade", listando todos os aplicativos disponíveis, suas versões, URLs de download e checksums.

---

## 👨‍💻 Desenvolvimento

Para compilar o projeto do zero, você precisará de:
- Go (1.21+)
- Node.js (20+)
- Wails CLI
- Dependências do GTK3 e WebKit2GTK (consulte a documentação do Wails).

**Comandos de Build:**

```bash
# Instalar Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Compilar o Daemon
cd daemon
go build -o ../dist/bin/updater-daemon

# Compilar o Client
cd client
wails build -o updater-client
```

---

## 📝 Licença

Este projeto é distribuído sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.
