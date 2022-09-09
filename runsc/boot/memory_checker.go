package boot
// #include <sys/mman.h>
// #include <unistd.h>
// #include <errno.h>
import "C"
import "time"
import "fmt"

// create memory chunk 
func CreateMemory(size int) (uintptr, err) {
	addr = C.NULL
	size = 1
	prots = C.PROT_READ | C.PROT_WRITE
	flags = C.MAP_SHARED
	return uintptr(C.mmap(addr, size, prots, flags, -1, 0)), nil
}

// If memory has changed ends job
func CheckMemoryContAndQuit(shm_mem uintptr) {
	changed := false
	for ; changed != true ; {
		if shm_mem[0] == 1 {
			changed := true
		}
		fmt.Println("Check, go to sleep")
		time.Sleep(1 * time.Second)
	}
	fmt.Println("Memory is changed, exit...")
}

