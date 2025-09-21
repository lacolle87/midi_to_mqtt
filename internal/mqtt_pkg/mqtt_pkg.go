package mqtt_pkg

import (
	"fmt"
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"midi_to_mqtt/internal/config"
)

func SetupClient(cfg config.Config) mqtt.Client {
	brokerURL := fmt.Sprintf("tcp://%s:%d", cfg.MQTTBroker, cfg.MQTTPort)
	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID("go-midi")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		slog.Error("MQTT connect error", "err", token.Error())
		panic(token.Error())
	}

	slog.Info("Connected to MQTT broker", "broker", brokerURL)
	return client
}
