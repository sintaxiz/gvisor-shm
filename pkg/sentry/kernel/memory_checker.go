package kernel

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"

)



type SharedMemoryManager struct {
	memory_chunks map[int]uintptr
	shm_mem       uintptr
}

// continuously check memory
func (smm *SharedMemoryManager) Start() {
	for {
		time.Sleep(1 * time.Second)
		if *(*int)(unsafe.Pointer(smm.shm_mem)) != 0 {
			fmt.Println("pid: %d", *(*int)(unsafe.Pointer(smm.shm_mem)))
			fmt.Println("syscallno: %d", *(*int)(unsafe.Pointer(smm.shm_mem + unsafe.Sizeof(int(0)))))
		}
	}
}

func (smm *SharedMemoryManager) AddProcess(pid int, memaddr uintptr) {
	smm.memory_chunks[pid] = memaddr
}

// create memory chunk
func (smm *SharedMemoryManager) CreateMemory(size int) error {
	fmt.Fprintf(os.Stdout, "Creating memory...")
	addr := 0x7f45221f7000
	prots := syscall.PROT_READ | syscall.PROT_WRITE
	flags := syscall.MAP_SHARED | syscall.MAP_ANONYMOUS | syscall.MAP_FIXED
	fd := -1
	a := 0
	shm_addr, _, err := syscall.Syscall6(syscall.SYS_MMAP, uintptr(addr), uintptr(size), uintptr(prots), uintptr(flags), uintptr(fd), uintptr(a))
	if err != 0 {
		fmt.Println(err)
		return err
	}
	smm.shm_mem = shm_addr
	return nil
}
