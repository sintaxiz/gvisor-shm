package boot

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// create memory chunk
func CreateMemory(size int) (uintptr, error) {
	fmt.Fprintf(os.Stdout, "Creating memory...")
	addr := 0x7f45221f7000
	prots := syscall.PROT_READ | syscall.PROT_WRITE
	flags := syscall.MAP_SHARED | syscall.MAP_ANONYMOUS | syscall.MAP_FIXED
	fd := -1
	a := 0
	shm_addr, _, err := syscall.Syscall6(syscall.SYS_MMAP, uintptr(addr), uintptr(size), uintptr(prots), uintptr(flags), uintptr(fd), uintptr(a))
	if err != 0 {
		fmt.Println(err)
		return 0, err
	}
	return shm_addr, nil
}

// If memory has changed ends job
func CheckMemoryContAndQuit(shm_mem uintptr) {
	changed := false
	for !changed {
		fmt.Println("Check, go to sleep")
		time.Sleep(1 * time.Second)
		if *(*int)(unsafe.Pointer(shm_mem)) != 0 {
			fmt.Println("shm_mem value: %d", *(*int)(unsafe.Pointer(shm_mem)))
		}
	}
	fmt.Println("Memory is changed, exit...")
}
