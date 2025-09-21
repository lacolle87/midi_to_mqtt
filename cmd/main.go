package main

import (
	"midi_to_mqtt/internal/config"
	"midi_to_mqtt/internal/logger"
	"midi_to_mqtt/internal/midi"
	"midi_to_mqtt/internal/mqtt_pkg"
)

func main() {
	cfg := config.LoadConfig()
	logger.SetupLogger(cfg.LogFile)

	mqttClient := mqtt_pkg.SetupClient(cfg)
	defer mqttClient.Disconnect(250)

	handle := midi.CreatePort(cfg.MIDIName, cfg.MaxSysexLength)
	defer midi.ClosePort(handle)

	midi.ReadAndPublish(handle, mqttClient, cfg.MIDIBufferSize, cfg.MQTTTopic)
}
