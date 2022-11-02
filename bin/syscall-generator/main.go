package main

import (
	"syscall"
	"fmt"
)

func getpid() int64 {
	syscall.Getpid()
	return 0
}

func main() {
	for i := 0; i < 1; i++{
		getpid()
	}
	fmt.Println("END!");
}
