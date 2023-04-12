package main

import (
	"flag"
	"fmt"
	"syscall"
	"unsafe"
)

func write_to_shm_mem() {

	addr := uintptr(0x7f45221f6000)
	//*(*int)(unsafe.Pointer(addr + unsafe.Sizeof(int(0)))) = 666
	*(*int)(unsafe.Pointer(addr)) = 666
	for !(*(*int)(unsafe.Pointer(addr)) == 0) {

	}
	//fmt.Println("pid = ", *(*int)(unsafe.Pointer(addr + unsafe.Sizeof(int(0)))))
}

func getpid() int64 {
	write_to_shm_mem()
	return 0
}

func direct_test(it int) {
	for i := 0; i < it; i++ {
		syscall.Getpid()
	}
}

func main() {

	defaultIterations := 10000

	shmPtr := flag.Bool("s", false, "shared memory syscalls")
	iterPtr := flag.Int("i", defaultIterations, "number of iterations")
	flag.Parse()
	fmt.Println("shared mode on? :", *shmPtr)
	fmt.Println("iterations:", *iterPtr)

	iterations := *iterPtr

	if !*shmPtr {
		direct_test(iterations)
	} else {
		for i := 0; i < iterations; i++ {
			getpid()
		}
	}

	//addr := uintptr(0x7f45221fa000)
	//*(*int)(unsafe.Pointer(addr)) = 666
}
