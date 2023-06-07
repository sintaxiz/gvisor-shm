package kernel

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
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
	ElfAddr          = 0x7f49221f6000
	SelectorAddress  = 0x7f45221f9000

	// mremap syscall constants
	MREMAP_FIXED   = 0x2
	MREMAP_MAYMOVE = 0x1

	// For flag "syscall.MAP_ANONYMOUS"
	DefaultMmapFd = -1

	// for debugging
	TestResult = 1337

	KernelIdForSmm = 0

	SupportedSyscalls = []int{
		4, // stat
		5, // fstat
		6, // lstat
		7, // poll
		8, // lseek
		39 // getpid
	}
)

func spinlock()

// continuously check memory
func (smm *SharedMemoryManager) Start() error {
	attachToCPU()
	log.Debugf("Start smm thread.")
	log.Debugf("Address for reading pid = %x", smm.pidPointer)

	for {
		p := smm.getPid()
		if p != 0 {
			syscallNum := *(*int)(unsafe.Pointer(smm.pidPointer))
			if !isSupportedSyscall(syscallNum) {
				failToPtrace();
			}
			log.Debugf("return before syscall: %x", uint64(smm.tasks[p].Arch().Return()))

			smm.tasks[p].Arch().SetSyscallNo(uint64(syscallNum))
			smm.tasks[p].completeSleep()

			smm.tasks[p].runState = smm.tasks[p].doSyscall()
			smm.tasks[p].prepareSleep()
			log.Debugf("return after syscall: %x", uint64(smm.tasks[p].Arch().Return()))
			smm.setResult(uint64(smm.tasks[666].Arch().Return())
			smm.endCommunication()

		}
		spinlock()
	}
	return nil
}

// set flag to end cycle
func (smm *SharedMemoryManager) Destroy() error {
	log.Debugf("Destroying smm...")
	smm.isDestroy = true
	return nil
}

func (smm *SharedMemoryManager) AddProcess(pid int, t *Task) error {
	log.Debugf("Adding process to smm...")

	size := SizePerProcess
	addr := InitAddress

	proc_mem, _, err := anonMmap(addr, size)
	if err != 0 {
		log.Warningf("Can not map memory, reason: %s", err.Error())
		return fmt.Errorf("Can not map memory, reason: %s", err.Error())
	}
	log.Infof("proc_mem=%x for pid=%d", proc_mem, pid)
	smm.tasks[666] = t
	return nil
}

func (smm *SharedMemoryManager) CreateAddr(pid int) (uintptr, error) {
	log.Debugf("Creating address in smm for pid=%d...", pid)

	new_addr_ptr := uintptr(InitFreeAddress + SizePerProcess*pid)
	smm.memoryChunks[pid] = new_addr_ptr
	return new_addr_ptr, nil
}

func (smm *SharedMemoryManager) AfterAddingProcess(pid int) error {
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
		return fmt.Errorf("Can not mremap memory for smm, reason: %s", err.Error())
	}
	smm.memoryChunks[pid] = new_addr_ptr
	log.Infof("new_addr=%x for pid=%d", new_addr_ptr, pid)
	return nil
}

func (smm *SharedMemoryManager) DeleteProcess(pid int) error {
	log.Debugf("Deleting process from smm...")

	delete(smm.memoryChunks, pid)
	return nil
}

// create memory chunk
func (smm *SharedMemoryManager) CreateMemory(size int) error {
	log.Debugf("Creating memory for smm...")

	smm.memoryChunks = make(map[int]uintptr)
	smm.tasks = make(map[int]*Task)
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

	// for SUD
	selectorAddr, _, err := anonMmap(SelectorAddress, size)
	if err != 0 {
		log.Debugf("Can not map memory for selectorAddr, reason: %s", err.Error())
		return err
	}
	log.Debugf("Selector addr: %x", selectorAddr)
	*(*int)(unsafe.Pointer(uintptr(selectorAddr))) = 0

	fd, err2 := syscall.Open("/tmp/syscalldispatcher.o", syscall.O_RDONLY, 0777)
	if err2 != nil {
		log.Debugf("Can not open file for smm, reason: %s", err2)
		return err2
	}
	log.Debugf("Fd: %d", fd)

	elf_mem, _, err := mmapElf(ElfAddr, fd)
	if err != 0 {
		log.Warningf("Can not map elf, reason: %s", err.Error())
		return nil
	}
	log.Debugf("Proc mem for elf file: %x", elf_mem)

	return nil
}

func (smm *SharedMemoryManager) getPid() int {
	return *(*int)(smm.pidPointer)
	//return spinlock(PidMemoryAddress)
}

func (smm *SharedMemoryManager) setResult(res uint64) {
	*(*uint64)(unsafe.Pointer(uintptr(smm.pidPointer) + unsafe.Sizeof(int(0)))) = res
}

func (smm *SharedMemoryManager) endCommunication() {
	*(*int)(smm.pidPointer) = 0
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

func mmapElf(addr int, fd int) (elfAddr, r2 uintptr, err syscall.Errno) {
	prots := syscall.PROT_READ | syscall.PROT_EXEC
	flags := syscall.MAP_SHARED | syscall.MAP_FIXED
	a := 0
	size := 0x3fa0

	return syscall.Syscall6(syscall.SYS_MMAP, uintptr(addr), uintptr(size), uintptr(prots), uintptr(flags), uintptr(fd), uintptr(a))
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

	var mask unix.CPUSet
	mask.Set(KernelIdForSmm)
	err := unix.SchedSetaffinity(0, &mask)
	if err != nil {
		log.Debugf("Error in attachToCPU in SMM")
	}

	// const __NR_sched_setaffinity = 203
	// var mask [1024 / 64]uint8
	// mask[1/64] |= 1 << (1 % 64)
	// _, _, errno := syscall.RawSyscall(__NR_sched_setaffinity, 0, uintptr(len(mask)*8), uintptr(unsafe.Pointer(&mask)))
	// if errno != 0 {
	// 	log.Debugf("Error in attachToCPU")
	// }
}

func isSupportedSyscall(sysno int) bool {
	for i := range SupportedSyscalls {
		if sysno == i {
			return true
		}
	}
	return false
}

func failToPtrace() {
	*(*int)(unsafe.Pointer(uintptr(SelectorAddress))) = 1
}