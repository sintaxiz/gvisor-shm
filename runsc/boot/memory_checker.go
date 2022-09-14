package boot
// #include <sys/mman.h>
// #include <unistd.h>
// #include <errno.h>
import "C"
import "time"
import "fmt"
import "unsafe"

// create memory chunk 
func CreateMemory(size int) (uintptr, error) {
	fmt.Println("Creating memory...")
	addr := C.NULL
	size_c := C.size_t(size)
	prots := C.int(C.PROT_READ | C.PROT_WRITE)
	flags := C.int(C.MAP_SHARED)
	return uintptr(C.mmap(addr, size_c, prots, flags, C.int(-1), 0)), nil
}

// If memory has changed ends job
func CheckMemoryContAndQuit(shm_mem uintptr) {
	changed := false
	for ; changed != true ; {
		if *(*int)(unsafe.Pointer(shm_mem)) == 1 {
			changed = true
		}
		fmt.Println("Check, go to sleep")
		time.Sleep(1 * time.Second)
	}
	fmt.Println("Memory is changed, exit...")
}

