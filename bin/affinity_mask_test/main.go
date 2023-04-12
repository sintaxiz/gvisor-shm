package main

import (
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

func pinToCPU(cpu uint) error {
	const __NR_sched_setaffinity = 203
	var mask [1024 / 64]uint8
	runtime.LockOSThread()
	mask[cpu/64] |= 1 << (cpu % 64)
	_, _, errno := syscall.RawSyscall(__NR_sched_setaffinity, 0, uintptr(len(mask)*8), uintptr(unsafe.Pointer(&mask)))
	if errno != 0 {
		return errno
	}
	return nil
}

func main() {
	//fmt.Println("value: ", pinToCPU(4))
	id, _, _ := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)

	var mask [1024 / 64]uintptr
	pid, _, _ := syscall.RawSyscall(39, 0, 0, 0)
	v1, _, err := syscall.RawSyscall(204, pid, uintptr(len(mask)*8), uintptr(unsafe.Pointer(&mask[0])))
	if err != 0 {
		fmt.Println("get affinity error")
		return
	}
	nmask := mask[:v1/4]
	var ret = make([]int, 0)
	idx := 0
	for _, v := range nmask {
		for i := 0; i < 64; i++ {
			ct := int32(v & 1)
			v >>= 1
			if ct > 0 {
				ret = append(ret, idx)
			}
			idx++
		}
	}

	if id == 0 {
		fmt.Println("In child:", id, ret)
	} else {
		fmt.Println("In parent:", id, ret)
	}
	time.Sleep(100000 * time.Second)
}
