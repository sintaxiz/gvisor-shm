package kernel

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"gvisor.dev/gvisor/pkg/log"
)

type SharedMemoryManager struct {
	memoryChunks       map[int]uintptr
	tasks              map[int]*Task
	pidPointer         unsafe.Pointer
	resultPointer      unsafe.Pointer
	isDestroy          bool
	currentFreeAddress uint64
}

const (
	// sizes in bytes
	SizePerProcess = 1

	// memory pointers
	InitAddress      = 0x7f45221f7000
	InitFreeAddress  = 0x7f45221f8000
	PidMemoryAddress = 0x7f45221f6000

	// mremap syscall constants
	MREMAP_FIXED   = 0x2
	MREMAP_MAYMOVE = 0x1

	// For flag "syscall.MAP_ANONYMOUS"
	DefaultMmapFd = -1

	// for debugging
	TestResult = 1337
)

// continuously check memory
func (smm *SharedMemoryManager) Start() {
	attachToCPU()
	log.Debugf("Start smm thread.")
	log.Debugf("Address for reading pid = %x", smm.pidPointer)

	for !smm.isDestroy {
		p := smm.getPid()
		if p != 0 {
			// fmt.Println("pid: %d", *(*int)(unsafe.Pointer(smm.shmMem)))
			// *(*int)(unsafe.Pointer(smm.shmMem + unsafe.Sizeof(int(0)))) = 1337
			// *(*int)(unsafe.Pointer(smm.shmMem)) = 0
			smm.tasks[p].Arch().SetSyscallNo(39)
			smm.tasks[p].doSyscall()
			smm.setResult(TestResult)
			smm.endCommunication()
		}
	}
}

func (smm *SharedMemoryManager) getPid() int {
	return *(*int)(smm.pidPointer)
}

func (smm *SharedMemoryManager) setResult(res int) {
	*(*int)(smm.resultPointer) = TestResult
}

func (smm *SharedMemoryManager) endCommunication() {
	*(*int)(smm.pidPointer) = 0
}

func (smm *SharedMemoryManager) AddProcess(pid int, t *Task) {
	log.Debugf("Adding process to smm...")

	size := SizePerProcess
	addr := InitAddress

	smm.tasks = make(map[int]*Task)

	proc_mem, _, err := anonMmap(addr, size)
	if err != 0 {
		log.Warningf("Can not map memory, reason: %s", err.Error())
		return
	}
	log.Infof("proc_mem=%x for pid=%d", proc_mem, pid)
	smm.tasks[pid] = t
}

func (smm *SharedMemoryManager) CreateAddr(pid int) uintptr {
	log.Debugf("Creating address in smm for pid=%d...", pid)

	new_addr_ptr := uintptr(InitFreeAddress + SizePerProcess*pid)
	smm.memoryChunks[pid] = new_addr_ptr
	return new_addr_ptr
}

func (smm *SharedMemoryManager) AfterAddingProcess(pid int) {
	log.Debugf("Doing things after adding process to smm...")
	// old_addr := DEFAULT_ADDR_FOR_NEW_PROC
	// new_addr := smm.createAddr(pid)
	addr := smm.currentFreeAddress
	log.Debugf("Current free address: %x", addr)
	smm.currentFreeAddress = smm.currentFreeAddress + 0x1000

	new_addr_ptr, _, err := mremap(addr)
	if err != 0 {
		log.Debugf("Can not mremap memory for smm, reason: %s", err.Error())
		fmt.Println(err)
		return
	}
	smm.memoryChunks[pid] = new_addr_ptr
	log.Infof("new_addr=%x for pid=%d", new_addr_ptr, pid)

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
	smm.currentFreeAddress = InitFreeAddress

	//fmt.Fprintf(os.Stdout, "Creating memory...")
	addr := PidMemoryAddress
	shmAddr, _, err := anonMmap(addr, size)
	if err != 0 {
		log.Debugf("Can not map memory for smm, reason: %s", err.Error())
		return err
	}
	smm.pidPointer = unsafe.Pointer(shmAddr)
	smm.resultPointer = unsafe.Pointer(shmAddr + unsafe.Sizeof(int(0)))
	return nil
}

func mremap(addr uint64) (new_addr_ptr, r2 uintptr, err syscall.Errno) {
	old_addr := InitAddress
	old_size := SizePerProcess
	new_size := SizePerProcess
	flags := MREMAP_MAYMOVE | MREMAP_FIXED
	new_addr := addr
	log.Debugf("parameters for mremap: %x, %x, %d, %d, %d", old_addr, new_addr, old_size, new_size, flags)
	return syscall.Syscall6(syscall.SYS_MREMAP, uintptr(old_addr), uintptr(old_size), uintptr(new_size), uintptr(flags), uintptr(new_addr), uintptr(0x0))
}

func anonMmap(addr int, size int) (shmAddr, r2 uintptr, err syscall.Errno) {
	prots := syscall.PROT_READ | syscall.PROT_WRITE
	flags := syscall.MAP_SHARED | syscall.MAP_ANONYMOUS | syscall.MAP_FIXED
	fd := DefaultMmapFd
	a := 0

	return syscall.Syscall6(syscall.SYS_MMAP, uintptr(addr), uintptr(size), uintptr(prots), uintptr(flags), uintptr(fd), uintptr(a))
}

func attachToCPU() {
	runtime.LockOSThread()

	const __NR_sched_setaffinity = 203
	var mask [1024 / 64]uint8
	mask[1/64] |= 1 << (1 % 64)
	_, _, errno := syscall.RawSyscall(__NR_sched_setaffinity, 0, uintptr(len(mask)*8), uintptr(unsafe.Pointer(&mask)))
	if errno != 0 {
		log.Debugf("Error in attachToCPU")
	}
}

// set flag to end cycle
func (smm *SharedMemoryManager) Destroy() {
	log.Debugf("Destroying smm...")

	smm.isDestroy = true
}
