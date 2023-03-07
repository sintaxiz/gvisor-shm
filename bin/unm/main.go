package main

import (
	"bytes"
	"fmt"
	"syscall"
	"unsafe"
)

func main() {

	//addr := uintptr(0x7f45221f6000)
	var uname syscall.Utsname
	// *(*uint64)(unsafe.Pointer(addr + unsafe.Sizeof(int(0)))) = uint64(uintptr(unsafe.Pointer(&uname)))
	// *(*int)(unsafe.Pointer(addr)) = 64
	// fmt.Printf("addr for uname: %x\n", unsafe.Pointer(&uname))

	// //r, _, errno := syscall.Syscall(syscall.SYS_UNAME, uintptr(unsafe.Pointer(&uname)), 0, 0)
	// for !(*(*int)(unsafe.Pointer(addr)) == 0) {

	// }
	// res := (unsafe.Pointer(addr + unsafe.Sizeof(int(0))))
	// fmt.Printf("res: %x\n", res)
	sysname := *(*[65]byte)(unsafe.Pointer(&uname.Sysname))
	nodename := *(*[65]byte)(unsafe.Pointer(&uname.Nodename))
	release := *(*[65]byte)(unsafe.Pointer(&uname.Release))
	version := *(*[65]byte)(unsafe.Pointer(&uname.Version))
	machine := *(*[65]byte)(unsafe.Pointer(&uname.Machine))

	fmt.Println("System Name: ", string(bytes.TrimRight(sysname[:], "\x00")))
	fmt.Println("Node Name: ", string(bytes.TrimRight(nodename[:], "\x00")))
	fmt.Println("Release: ", string(bytes.TrimRight(release[:], "\x00")))
	fmt.Println("Version: ", string(bytes.TrimRight(version[:], "\x00")))
	fmt.Println("Machine: ", string(bytes.TrimRight(machine[:], "\x00")))
}
