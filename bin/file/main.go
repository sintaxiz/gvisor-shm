package main

import (
	"fmt"
	"unsafe"
)

func open() int {

	addr := uintptr(0x7f45221f6000)
	//*(*int)(unsafe.Pointer(addr + unsafe.Sizeof(int(0)))) = 666
	filepath := []byte("/home/smollfile\x00")
	*(*int)(unsafe.Pointer(addr)) = 2
	fmt.Println("file addr = ", uint64(uintptr(unsafe.Pointer(&filepath[0]))))
	*(*uint64)(unsafe.Pointer(addr + unsafe.Sizeof(int(0)))) = uint64(uintptr(unsafe.Pointer(&filepath[0])))

	for !(*(*int)(unsafe.Pointer(addr)) == 0) {

	}
	res := *(*int)(unsafe.Pointer(addr + unsafe.Sizeof(int(0))))
	return res
	//fmt.Println("pid = ", *(*int)(unsafe.Pointer(addr + unsafe.Sizeof(int(0)))))
}

func read(fd int) int {
	buf := []byte{91, 91, 91, 91, 91}
	addr := uintptr(0x7f45221f6000)
	//*(*int)(unsafe.Pointer(addr + unsafe.Sizeof(int(0)))) = 666
	fmt.Println("buf addr = ", uint64(uintptr(unsafe.Pointer(&buf[0]))))
	*(*uint64)(unsafe.Pointer(addr + unsafe.Sizeof(int(0)))) = uint64(fd)
	*(*uint64)(unsafe.Pointer(addr + 2*unsafe.Sizeof(int(0)))) = uint64(uintptr(unsafe.Pointer(&buf[0])))
	*(*int)(unsafe.Pointer(addr)) = 666

	for !(*(*int)(unsafe.Pointer(addr)) == 0) {

	}
	res := *(*int)(unsafe.Pointer(addr + unsafe.Sizeof(int(0))))
	fmt.Println("Read: ", buf)
	return res
	//fmt.Println("pid = ", *(*int)(unsafe.Pointer(addr + unsafe.Sizeof(int(0)))))
}

func main() {
	fd := open()
	if fd == -1 {
		fmt.Println("Can not open/create file")
		return
	}
	fmt.Println("fd1 = ", fd)

	// file := os.NewFile(4, "ff")
	// _, err := file.Write([]byte(`my data`))
	// if err != nil {
	// 	panic(err)
	// }

	//_, err := syscall.Write(fd, buf)
	n := read(fd)
	// if err != nil {
	// 	fmt.Println("Can not write to file ", err)
	// }
	fmt.Println("Read: ", n)

	//time.Sleep(100 * time.Second)

}
