package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func remapMemory() {
	fmt.Fprintf(os.Stdout, "Remaping memory...\n")

	old_addr := 0x7f45221f7000
	old_size := 1
	new_size := 1
	MREMAP_FIXED := 0x2
	MREMAP_MAYMOVE := 0x1
	flags := MREMAP_MAYMOVE | MREMAP_FIXED
	new_addr := 0x7f45221f8000

	new_addr_ptr, _, err := syscall.Syscall6(syscall.SYS_MREMAP, uintptr(old_addr), uintptr(old_size), uintptr(new_size), uintptr(flags), uintptr(new_addr), uintptr(0x0))
	if err != 0 {
		fmt.Println(err)
		return
	}
	fmt.Fprintf(os.Stdout, "new_addr=%x\n", new_addr_ptr)

	fmt.Println("value: ", *(*int)(unsafe.Pointer(uintptr(new_addr))))
}

func createMemory() {
	fmt.Fprintf(os.Stdout, "Creating memory...\n")

	size := 4
	addr := 0x7f45221f7000
	prots := syscall.PROT_READ | syscall.PROT_WRITE
	flags := syscall.MAP_SHARED | syscall.MAP_ANONYMOUS | syscall.MAP_FIXED
	fd := -1
	a := 0
	shm_addr, _, err := syscall.Syscall6(syscall.SYS_MMAP, uintptr(addr), uintptr(size), uintptr(prots), uintptr(flags), uintptr(fd), uintptr(a))
	if err != 0 {
		fmt.Println(err)
	}
	fmt.Fprintf(os.Stdout, "shm_addr=%x\n", shm_addr)

	*(*int)(unsafe.Pointer(shm_addr)) = 666

}

func main() {
	createMemory()
	remapMemory()
}
