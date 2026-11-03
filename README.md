# Weibo Control

Central de configuração e controle por software para mouses Weibo no Linux (Zorin OS / Ubuntu).

---

## 🎯 Por Que Controle por Software?

Durante a fase de testes e engenharia reversa do hardware (`YICHIP 3151:3020`), comprovamos que o receptor de rádio do mouse físico **não aceita comandos remotos de firmware** (o chip do mouse físico opera de forma unidirecional e analógica).

Para não criar uma interface de "placebo" (botões falsos que não fazem nada), o **Weibo Control** opera na camada do subsistema de entrada do Linux (`XInput` + `/dev/uinput`):
1. **DPI e Sensibilidade Real**: Ajusta a matriz linear de coordenadas com máxima fidelidade e zero perda de rastreamento.
2. **Remapeamento de Botões**: Detecta e reconfigura cliques dos botões físicos e laterais.
3. **Editor de Macros e Atalhos**: Dispara qualquer combinação de teclas via teclado virtual `/dev/uinput`.
4. **Perfis Persistentes**: Salva perfis personalizados em `~/.config/weibo-control/config.json`.

---

## 🚀 Como Testar o Driver no Terminal

Navegue até a pasta do driver:
```bash
cd /home/carlos/Drivers/weibo/weibo-driver
```

### 1. Testar Mudança Real de DPI (Você sente na ponta dos dedos)
Altere para 2400 DPI (o cursor vai acelerar imediatamente):
```bash
./bin/weibo-cli dpi 2400
```
Verifique o status da sensibilidade ativa:
```bash
./bin/weibo-cli dpi-status
```
Restaure para o DPI padrão (1200 DPI):
```bash
./bin/weibo-cli dpi 1200
```

### 2. Testar Monitoramento de Botões Físicos (Ao Vivo)
Execute e clique com o botão esquerdo, direito, meio e os **botões laterais**:
```bash
./bin/weibo-cli listen-buttons
```
*(Ele mostrará no terminal cada botão pressionado em tempo real! Pressione Ctrl+C para sair).*

### 3. Testar Injeção de Teclas e Macros
Emita um atalho de teclado através do `/dev/uinput`:
```bash
./bin/weibo-cli macro ctrl+c
./bin/weibo-cli macro volume_up
```

### 4. Executar o Daemon de Botões Laterais
Inicia o serviço que intercepta os botões laterais do mouse (botões 8 e 9) e dispara as ações configuradas no perfil:
```bash
./bin/weibo-cli daemon
```

---

## 🖥️ Interface Gráfica Fyne (`weibo-control`)

A estrutura da interface gráfica está configurada na pasta `weibo-control`.
Para compilar apps Fyne no Zorin OS / Ubuntu (que exigem headers de X11 e OpenGL via Cgo), execute uma única vez no terminal:

```bash
sudo apt-get install -y pkg-config libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libgl1-mesa-dev
```
