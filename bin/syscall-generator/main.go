package main

import (
	"syscall"
)

func getpid() int {
	return syscall.Getpid()
}

func main() {
	iterations := 10000
	for i := 0; i < iterations; i++ {
		getpid()
	}
}
