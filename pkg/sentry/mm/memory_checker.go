package mm

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"

	"gvisor.dev/gvisor/pkg/log"
	"gvisor.dev/gvisor/pkg/memutil"
)

type SharedMemoryManager struct {
	memory_chunks map[int]uintptr
	shm_mem       uintptr
	memfd         int
	isOn          bool
}

func (smm *SharedMemoryManager) New() {
	smm.isOn = false
}

// continuously check memory
func (smm *SharedMemoryManager) Start() {
	smm.isOn = true
	for {
		time.Sleep(1 * time.Second)
		fmt.Println("shm_mem: %d", *(*int)(unsafe.Pointer(smm.shm_mem)))
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

	memfd, err := memutil.CreateMemFD("shm-mem", 0)
	if err != nil {
		return fmt.Errorf("error creating memfd: %v", err)
	}
	log.Debugf("Create memfd")

	fmt.Fprintf(os.Stdout, "Creating memory...")
	addr := 0x7f45221f7000
	prots := syscall.PROT_READ | syscall.PROT_WRITE
	flags := syscall.MAP_SHARED | syscall.MAP_FIXED

	fd := memfd
	a := 0
	shm_addr, _, errno := syscall.Syscall6(syscall.SYS_MMAP, uintptr(addr), uintptr(size), uintptr(prots), uintptr(flags), uintptr(fd), uintptr(a))
	if errno != 0 {
		fmt.Println(errno)
		return errno
	}
	smm.memfd = memfd
	smm.shm_mem = shm_addr
	return nil
}
