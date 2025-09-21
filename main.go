package main

import (
	"fmt"
	"log"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	dll                  = windows.NewLazySystemDLL("teVirtualMIDI64.dll")
	teVirtualMIDICreate  = dll.NewProc("virtualMIDICreatePortEx2")
	teVirtualMIDIClose   = dll.NewProc("virtualMIDIClosePort")
	teVirtualMIDIGetData = dll.NewProc("virtualMIDIGetData")
)

func main() {
	name, err := windows.UTF16PtrFromString("GoVirtualMIDI")
	if err != nil {
		log.Fatalf("Ошибка конвертации имени: %v", err)
	}

	handle, _, errCreate := teVirtualMIDICreate.Call(
		uintptr(unsafe.Pointer(name)), // LPCWSTR
		0,                             // callback
		0,                             // userData
		65535,                         // maxSysexLength
		0,                             // manufacturer
		0,                             // product
		1,                             // flags
	)
	if handle == 0 {
		log.Fatalf("Ошибка создания порта: %v", errCreate)
	}
	defer teVirtualMIDIClose.Call(handle)

	buf := make([]byte, 65535)
	for {
		length := uint32(len(buf))
		r, _, e := teVirtualMIDIGetData.Call(
			handle,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&length)),
		)
		if r == 0 {
			log.Fatalf("Ошибка чтения: %v", e)
		}
		fmt.Printf("MIDI: % X\n", buf[:length])
	}
}
