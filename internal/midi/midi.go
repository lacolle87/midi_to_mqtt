package midi

import (
	"runtime"
	"sync"
	"sync/atomic"

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

type ringBuffer struct {
	data [][]byte
	head uint32
	tail uint32
	mask uint32
}

func newRingBuffer(size int) *ringBuffer {
	if size&(size-1) != 0 {
		panic("ring buffer size must be power of two")
	}
	return &ringBuffer{
		data: make([][]byte, size),
		mask: uint32(size - 1),
	}
}

func (r *ringBuffer) push(b []byte) bool {
	h := atomic.LoadUint32(&r.head)
	t := atomic.LoadUint32(&r.tail)
	if h-t > r.mask {
		return false
	}
	r.data[h&r.mask] = b
	atomic.StoreUint32(&r.head, h+1)
	return true
}

func (r *ringBuffer) pop() ([]byte, bool) {
	h := atomic.LoadUint32(&r.head)
	t := atomic.LoadUint32(&r.tail)
	if h == t {
		return nil, false
	}
	b := r.data[t&r.mask]
	r.data[t&r.mask] = nil
	atomic.StoreUint32(&r.tail, t+1)
	return b, true
}

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
	pool := sync.Pool{New: func() any { return make([]byte, bufSize) }}
	ring := newRingBuffer(256)

	go func() {
		for {
			msg, ok := ring.pop()
			if !ok {
				runtime.Gosched()
				continue
			}

			token := client.Publish(topic, 0, false, msg)
			token.Wait()
			if token.Error() != nil {
				slog.Error("MQTT publish failed", "err", token.Error())
			}

			pool.Put(msg[:cap(msg)])
		}
	}()

	for {
		buf := pool.Get().([]byte)
		if cap(buf) < bufSize {
			buf = make([]byte, bufSize)
		} else {
			buf = buf[:cap(buf)]
		}
		length := uint32(cap(buf))
		r, _, e := teVirtualMIDIGetData.Call(
			handle,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&length)),
		)
		if r == 0 {
			slog.Error("Error reading MIDI", "err", e)
			panic(e)
		}
		msg := buf[:length]
		if !ring.push(msg) {
			slog.Warn("Ring buffer full, dropping MIDI data")
			pool.Put(buf[:cap(buf)])
		}
	}
}
