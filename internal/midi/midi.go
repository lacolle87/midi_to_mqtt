package midi

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"golang.org/x/sys/windows"
	"log/slog"
	"unsafe"
)

var (
	dll                  = windows.NewLazySystemDLL("teVirtualMIDI64.dll")
	teVirtualMIDICreate  = dll.NewProc("virtualMIDICreatePortEx2")
	teVirtualMIDIClose   = dll.NewProc("virtualMIDIClosePort")
	teVirtualMIDIGetData = dll.NewProc("virtualMIDIGetData")
)

func CreatePort(name string, maxSysex int) uintptr {
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

func ClosePort(handle uintptr) {
	teVirtualMIDIClose.Call(handle)
}

func ReadAndPublish(handle uintptr, client mqtt.Client, bufSize int, topic string) {
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

		payload := make([]byte, length)
		copy(payload, buf[:length])

		token := client.Publish(topic, 0, false, payload)
		token.Wait()

		if token.Error() != nil {
			slog.Error("MQTT publish failed", "err", token.Error())
		}
	}
}
