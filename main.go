package main

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/lacolle87/eqmlog"
	"github.com/spf13/viper"
	"golang.org/x/sys/windows"
	"log/slog"
)

var (
	dll                  = windows.NewLazySystemDLL("teVirtualMIDI64.dll")
	teVirtualMIDICreate  = dll.NewProc("virtualMIDICreatePortEx2")
	teVirtualMIDIClose   = dll.NewProc("virtualMIDIClosePort")
	teVirtualMIDIGetData = dll.NewProc("virtualMIDIGetData")
	cfg                  Config
	logFile              string
)

type Config struct {
	MQTTBroker     string
	MQTTPort       int
	MQTTTopic      string
	MIDIName       string
	MIDIBufferSize int
	MaxSysexLength int
}

func main() {
	loadConfig()

	if err := setupLogger(); err != nil {
		panic("Logger setup failed: " + err.Error())
	}

	client := setupMQTT()
	defer client.Disconnect(250)

	handle := createMIDIPort(cfg.MIDIName, cfg.MaxSysexLength)
	defer teVirtualMIDIClose.Call(handle)

	readAndPublishMIDI(handle, client, cfg.MIDIBufferSize, cfg.MQTTTopic)
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic("Failed to read config: " + err.Error())
	}

	cfg = Config{
		MQTTBroker:     viper.GetString("mqtt.broker"),
		MQTTPort:       viper.GetInt("mqtt.port"),
		MQTTTopic:      viper.GetString("mqtt.topic"),
		MIDIName:       viper.GetString("midi.name"),
		MIDIBufferSize: viper.GetInt("midi.buffer_size"),
		MaxSysexLength: viper.GetInt("midi.max_sysex_length"),
	}

	logFile = viper.GetString("logger.filename")
	if logFile == "" {
		logFile = "logs/midi_to_mqtt.log"
	}
}

func setupLogger() error {
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	multiWriter := eqmlog.LoadLogger()
	logger := slog.New(slog.NewTextHandler(multiWriter, nil))
	slog.SetDefault(logger)
	slog.Info("Logger initialized", "file", logFile)
	return nil
}

func setupMQTT() mqtt.Client {
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

func createMIDIPort(name string, maxSysex int) uintptr {
	ptr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		slog.Error("Port name conversion error", "err", err)
		panic(err)
	}
	handle, _, errCreate := teVirtualMIDICreate.Call(
		uintptr(unsafe.Pointer(ptr)),
		0,
		0,
		uintptr(maxSysex),
		0,
		0,
		1,
	)
	if handle == 0 {
		slog.Error("Failed to create MIDI port", "err", errCreate)
		panic(errCreate)
	}
	slog.Info("Virtual MIDI port created", "port", name)
	return handle
}

func readAndPublishMIDI(handle uintptr, client mqtt.Client, bufSize int, topic string) {
	buf := make([]byte, bufSize)
	for {
		length := uint32(len(buf))
		r, _, e := teVirtualMIDIGetData.Call(
			handle,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&length)),
		)
		if r == 0 {
			slog.Error("Error reading MIDI", "err", e)
			panic(e)
		}
		payload := string(buf[:length])
		token := client.Publish(topic, 0, false, payload)
		token.Wait()
		slog.Info("MIDI sent", "hex", payload)
	}
}
