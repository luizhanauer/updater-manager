# 📦 Updater Manager

![Updater Manager Screenshot](https://raw.githubusercontent.com/luizhanauer/updater-manager/main/portfolio-assets/screenshot.png)

Um gerenciador de aplicativos centralizado para Linux. Fornece uma interface gráfica simples para instalar, atualizar e remover aplicativos de um catálogo remoto customizável, com um daemon robusto para gerenciar os pacotes em segundo plano.

O objetivo é simplificar a distribuição e manutenção de um conjunto de aplicativos em múltiplos sistemas, seja para uma equipe, uma empresa ou para uso pessoal em ambientes Ubuntu 24.04.

---

## ✨ Funcionalidades

- **Interface Gráfica Simples:** Gerencie seus aplicativos com uma UI limpa e intuitiva construída com Wails (Go + Vue.js com TypeScript).
- **Atualizações Automáticas:** Configure aplicativos para serem mantidos atualizados automaticamente em segundo plano.
- **Daemon Robusto:** Um serviço de sistema (`systemd`) cuida de todas as operações pesadas, como downloads e instalações.
- **Pré-verificação de Instalação:** Simula instalações para detectar problemas (como dependências quebradas) *antes* de modificar o sistema, evitando estados inconsistentes.
- **Catálogo Centralizado:** Gerencie a lista de aplicativos disponíveis a partir de um único arquivo `catalog.json` hospedado remotamente.
- **Segurança:** Usa permissões de grupo do Linux para garantir a comunicação segura entre a interface do usuário e o daemon de sistema.

---

## 🚀 Instalação

### Método Rápido (Recomendado)

Este comando irá baixar e executar o script de instalação oficial através do GitHub Pages, que detecta a arquitetura do seu sistema, baixa a última versão compilada e configura os serviços com as permissões necessárias.

```bash
curl -fsSL https://luizhanauer.github.io/updater-manager/get.sh | sh
```

> **⚠️ Importante!**
> Após a instalação, você **precisa fazer logout e login novamente** na sua sessão de usuário. Isso é necessário para que o sistema reconheça sua nova associação ao grupo `linux-updater` e permita que a interface gráfica se comunique com o daemon via Socket.

### Método Manual

1. Vá para a aba de **Releases** no repositório.
2. Baixe o arquivo `.tar.gz` correspondente à sua arquitetura (ex: `updater-manager_linux_amd64.tar.gz`).
3. Extraia o conteúdo e execute o script de instalação:
```bash
# Extrai o arquivo baixado
tar -xzf updater-manager_linux_amd64.tar.gz -C updater-manager

# Entra no diretório extraído
cd updater-manager

# Execute o instalador como root
sudo ./install.sh

```


4. Faça **logout e login** após a conclusão da instalação.

---

## 🛠️ Arquitetura

O projeto é dividido em três componentes principais, respeitando a separação de responsabilidades:

1. **Daemon (`updater-daemon`)**
* Serviço Go que roda em segundo plano, gerenciado pelo `systemd`.
* Executa operações de pacotes: verificação de versões, downloads, simulações e instalações de arquivos `.deb`.
* Expõe um Unix Socket restrito em `/run/linux-updater/service.sock` para comunicação.


2. **Client (`updater-client`)**
* Aplicação desktop construída com Wails.
* Consome os dados e fornece a interface para o usuário visualizar o catálogo e configurar as rotinas.
* Comunica-se com o Daemon exclusivamente através do Unix Socket.


3. **Registry (`catalog.json`)**
* Arquivo JSON hospedado remotamente (Git ou HTTP).
* Funciona como a "fonte da verdade", listando aplicativos, versões, URLs e checksums.



---

## 👨‍💻 Desenvolvimento

Para compilar o projeto localmente, você precisará das seguintes ferramentas instaladas:

* Go (1.25.6+)
* Node.js (22+)
* Wails CLI
* Dependências do sistema: `libgtk-3-dev`, `libwebkit2gtk-4.1-dev`, `build-essential`

### Comandos de Build

```bash
# 1. Instalar Wails CLI
go install [github.com/wailsapp/wails/v2/cmd/wails@latest](https://github.com/wailsapp/wails/v2/cmd/wails@latest)

# 2. Compilar o Daemon
mkdir -p dist/bin
cd daemon
go mod tidy
go build -ldflags="-s -w" -o ../dist/bin/updater-daemon
cd ..

# 3. Compilar o Client
cd client
go mod tidy
wails build -platform linux/amd64 -tags webkit2_41 -clean -o updater-client
mv build/bin/updater-client ../dist/bin/
cd ..
```

---

## Contribuição

Contribuições são bem-vindas! Se você encontrar algum problema ou tiver sugestões para melhorar a aplicação, sinta-se à vontade para abrir uma issue ou enviar um pull request.

Se você gostou do meu trabalho e quer me agradecer, você pode me pagar um café :)

<a href="https://www.paypal.com/donate/?hosted_button_id=SFR785YEYHC4E" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" style="height: 40px !important;width: 150px !important;" ></a>

## Licença

Este projeto está licenciado sob a Licença MIT. Consulte o arquivo LICENSE para obter mais informações.

<p align="center">Desenvolvido por <strong>Luiz Hanauer</strong></p>