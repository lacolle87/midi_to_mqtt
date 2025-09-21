# MIDI to MQTT Bridge

## About

This Go application creates a virtual MIDI port using the [Tobias Erichsen VirtualMIDI driver](https://www.tobias-erichsen.de/). It receives MIDI events from a DAW in real time and forwards them over MQTT, enabling MIDI notes or control changes to drive lighting, effects, or other devices. This bridges music software with IoT setups for interactive shows and real-time hardware control.

---

## Features

- Create a virtual MIDI port on Windows.
- Read MIDI messages, including SysEx data.
- Publish MIDI data to an MQTT broker.
- Control lights, LEDs, or effects via microcontrollers.
- Minimal setup, configurable via `config.yaml`.

---

## Requirements

- Go 1.25+
- Windows OS
- MQTT broker (e.g., Mosquitto)
- [`teVirtualMIDI64.dll`](https://www.tobias-erichsen.de/) **must be downloaded manually** and placed in the project root directory.

> ⚠️ **Important:** This repository does **not** include the `teVirtualMIDI64.dll` file. You need to download it yourself from [Tobias Erichsen’s website](https://www.tobias-erichsen.de/).

---

## Installation

1. Clone the repository:

```bash
git clone https://github.com/lacolle87/midi_to_mqtt.git
cd midi_to_mqtt
````

2. Download [`teVirtualMIDI64.dll`](https://www.tobias-erichsen.de/) and place it in the project root.

3. Install Go dependencies:

```bash
go mod tidy
```

---

## Usage

Run the application:

```bash
go run cmd/app/main.go
```

Or build an executable:

```bash
go build -o midi_to_mqtt cmd/app/main.go
./midi_to_mqtt
```

---

## Logging

Logs are written to `logs/app.log` by default. You can change this path in `config.yaml`.

---

## Docker (Optional)

A `compose.yaml` is included for running a local MQTT broker with Mosquitto:

```bash
docker compose up -d
```

---

## Use Cases

* Trigger lights, LEDs, or effects from MIDI controllers or DAWs.
* Connect microcontrollers via MQTT for real-time device control.
* Enable interactive music shows or installations.

---

## License

This project is licensed under the MIT License.
