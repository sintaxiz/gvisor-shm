package main

import (
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

func main() {
	iterations := 10000
	for i := 0; i < iterations; i++ {
		getpid()
	}

	//addr := uintptr(0x7f45221fa000)
	//*(*int)(unsafe.Pointer(addr)) = 666
}
