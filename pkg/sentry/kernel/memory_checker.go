package kernel

import (
	"fmt"
	"syscall"
	"unsafe"

	"gvisor.dev/gvisor/pkg/log"
)

type SharedMemoryManager struct {
	memoryChunks       map[int]uintptr
	shmMem             uintptr
	isDestroy          bool
	currentFreeAddress uint64
}

// continuously check memory
func (smm *SharedMemoryManager) Start() {
	log.Debugf("Start smm thread")
	log.Debugf("smm.shmMem = %x", smm.shmMem)
	for !smm.isDestroy {
		if *(*int)(unsafe.Pointer(smm.shmMem)) != 0 {
			fmt.Println("pid: %d", *(*int)(unsafe.Pointer(smm.shmMem)))
			*(*int)(unsafe.Pointer(smm.shmMem + unsafe.Sizeof(int(0)))) = 1337
			*(*int)(unsafe.Pointer(smm.shmMem)) = 0
		}
	}
}

func (smm *SharedMemoryManager) AddProcess(pid int) {
	log.Debugf("Adding process to smm...")

	size := 1
	addr := 0x7f45221f6000

	prots := syscall.PROT_READ | syscall.PROT_WRITE
	flags := syscall.MAP_SHARED | syscall.MAP_ANONYMOUS | syscall.MAP_FIXED
	fd := -1
	a := 0

	proc_mem, _, err := syscall.Syscall6(syscall.SYS_MMAP, uintptr(addr), uintptr(size), uintptr(prots), uintptr(flags), uintptr(fd), uintptr(a))
	if err != 0 {
		log.Warningf("Can not map memory, reason: %s", err.Error())
		return
	}
	log.Infof("proc_mem=%x for pid=%d", proc_mem, pid)
}

func (smm *SharedMemoryManager) CreateAddr(pid int) uintptr {
	log.Debugf("Creating address in smm for pid=%d...", pid)

	new_addr_ptr := uintptr(0x7f45221f7300 + 0x100*pid)
	smm.memoryChunks[pid] = new_addr_ptr
	return new_addr_ptr
}

func (smm *SharedMemoryManager) AfterAddingProcess(pid int) {
	log.Debugf("Doing things after adding process to smm...")
	// old_addr := DEFAULT_ADDR_FOR_NEW_PROC
	// new_addr := smm.createAddr(pid)
	addr := smm.currentFreeAddress
	log.Debugf("Current free address: %x", addr)
	smm.currentFreeAddress = smm.currentFreeAddress + 0x100

	old_addr := 0x7f45221f6000
	old_size := 1
	new_size := 1
	MREMAP_FIXED := 0x2
	MREMAP_MAYMOVE := 0x1
	flags := MREMAP_MAYMOVE | MREMAP_FIXED
	new_addr := addr

	new_addr_ptr, _, err := syscall.Syscall6(syscall.SYS_MREMAP, uintptr(old_addr), uintptr(old_size), uintptr(new_size), uintptr(flags), uintptr(new_addr), uintptr(0x0))
	if err != 0 {
		log.Debugf("Can not mremap memory for smm, reason: %s", err.Error())
		fmt.Println(err)
		return
	}
	smm.memoryChunks[pid] = new_addr_ptr
	log.Infof("new_addr=%x for pid=%d", new_addr, pid)

}

func (smm *SharedMemoryManager) DeleteProcess(pid int) {
	log.Debugf("Deleting process from smm...")

	delete(smm.memoryChunks, pid)
}

// create memory chunk
func (smm *SharedMemoryManager) CreateMemory(size int) error {
	log.Debugf("Creating memory for smm...")

	smm.memoryChunks = make(map[int]uintptr)
	smm.isDestroy = false
	smm.currentFreeAddress = 0x7f45221f7100

	//fmt.Fprintf(os.Stdout, "Creating memory...")
	addr := 0x7f45221f7000
	prots := syscall.PROT_READ | syscall.PROT_WRITE
	flags := syscall.MAP_SHARED | syscall.MAP_ANONYMOUS | syscall.MAP_FIXED
	fd := -1
	a := 0
	shm_addr, _, err := syscall.Syscall6(syscall.SYS_MMAP, uintptr(addr), uintptr(size), uintptr(prots), uintptr(flags), uintptr(fd), uintptr(a))
	if err != 0 {
		log.Debugf("Can not map memory for smm, reason: %s", err.Error())
		return err
	}
	smm.shmMem = shm_addr
	return nil
}

// set flag to end cycle
func (smm *SharedMemoryManager) Destroy() {
	log.Debugf("Destroying smm...")

	smm.isDestroy = true
}
