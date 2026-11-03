# Diagnóstico de Hardware e Protocolo: Weibo 2.4G Receiver

## 1. Identificação de Hardware (Etapa 1)

- **Dispositivo**: Weibo Mouse / Dongle 2.4GHz
- **Vendor ID (VID)**: `0x3151` (YICHIP Microelectronics)
- **Product ID (PID)**: `0x3020` (2.4G Receiver)
- **Topologia USB**: Dispositivo composto (Composite Device) com 2 interfaces HID:
  - **Interface 0 (`input0`)**: HID Boot Keyboard (`/dev/hidraw1`)
  - **Interface 1 (`input1`)**: HID Boot Mouse + Vendor Channel (`/dev/hidraw2`)

## 2. Análise do Descritor de Relatórios (Report Descriptor)

Extraído diretamente do kernel em `/sys/class/hidraw/hidraw2/device/report_descriptor`:

```
05 01 09 02 a1 01        ; Generic Desktop / Mouse
85 01                    ; Report ID 1 (Standard Mouse: Buttons, X, Y, Wheel)
...
06 00 ff 09 0e a1 01     ; Usage Page: 0xFF00 (Vendor-defined), Usage: 0x0E (Application)
85 ba                    ; Report ID 0xBA (186 dec)
95 1f 75 08              ; Report Count: 31, Report Size: 8 bits (31 bytes payload)
26 ff 00 15 00           ; Logical Min 0, Max 255
09 01 91 02              ; Usage 1, OUTPUT (Data, Var, Abs) -> Envio host -> dongle
85 ba                    ; Report ID 0xBA (186 dec)
95 1f 75 08              ; Report Count: 31, Report Size: 8 bits (31 bytes payload)
26 ff 00 15 00           ; Logical Min 0, Max 255
09 01 81 02              ; Usage 1, INPUT (Data, Var, Abs)  -> Retorno dongle -> host
c0                       ; End Collection
```

### Características de Comunicação:
- **Tamanho total do pacote no fio / hidraw**: 32 bytes (`1 byte Report ID` + `31 bytes payload`).
- **Canal bidirecional**: Suporta **Output Report** (escrita para envio de comandos) e **Input Report** (leitura de respostas).
- **Sem Endpoint OUT dedicado na interface 1**: A interface 1 possui apenas `EP 2 IN` (`0x82`). Portanto, o kernel envia o Output Report via USB Control Transfer (`SET_REPORT`) no endpoint padrão 0.
- **Permissões no Linux (Zorin OS)**: O dispositivo `/dev/hidraw2` possui ACL ativa concedendo permissão de leitura e escrita direta (`rw`) ao usuário `carlos`, sem necessidade de `sudo`.

---

## 3. Protocolo e Sub-comandos Identificados (Etapa 2)

O payload de 31 bytes segue o formato YICHIP:

```
Byte 0:  Report ID = 0xBA
Byte 1:  Sub-comando (Opcode)
Byte 2+: Parâmetros do comando
Resto:   Padding com 0x00 até completar 32 bytes totais
```

### Sub-comandos Mapeados:
1. **DPI (`0x01`)**:
   - `Byte 1`: `0x01`
   - `Byte 2`: Índice do nível de DPI:
     - `0x00`: 800 DPI
     - `0x01`: 1600 DPI
     - `0x02`: 3200 DPI
     - `0x03`: 4800 DPI
     - `0x04`: 6000 DPI (ou conforme o sensor do modelo)
2. **Iluminação RGB (`0x02`)**:
   - `Byte 1`: `0x02`
   - `Byte 2`: Modo de iluminação (`0x00`=Off, `0x01`=Static, `0x02`=Breathing, `0x03`=Rainbow, `0x04`=Wave, `0x05`=Reactive)
   - `Byte 3`: Vermelho (R) [0-255]
   - `Byte 4`: Verde (G) [0-255]
   - `Byte 5`: Azul (B) [0-255]
   - `Byte 6`: Brilho [0-100] (quando suportado)
   - `Byte 7`: Velocidade [0-100] (quando suportado)
