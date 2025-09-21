package config

import "github.com/spf13/viper"

type Config struct {
	MQTTBroker     string
	MQTTPort       int
	MQTTTopic      string
	MIDIName       string
	MIDIBufferSize int
	MaxSysexLength int
	LogFile        string
}

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic("Failed to read config: " + err.Error())
	}

	cfg := Config{
		MQTTBroker:     viper.GetString("mqtt.broker"),
		MQTTPort:       viper.GetInt("mqtt.port"),
		MQTTTopic:      viper.GetString("mqtt.topic"),
		MIDIName:       viper.GetString("midi.name"),
		MIDIBufferSize: viper.GetInt("midi.buffer_size"),
		MaxSysexLength: viper.GetInt("midi.max_sysex_length"),
		LogFile:        viper.GetString("logger.filename"),
	}

	if cfg.LogFile == "" {
		cfg.LogFile = "logs/midi_to_mqtt.log"
	}

	return cfg
}
