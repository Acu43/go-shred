package main

import (
	"fmt"
	"os"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go-shred <file>")
		os.Exit(1)
	}

	arg := os.Args[1]

	err := Shred(arg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Shred error:", err)

		if pathErr, ok := err.(*os.PathError); ok {
			if errno, ok := pathErr.Err.(syscall.Errno); ok {
				os.Exit(int(errno))
			}
		}
		os.Exit(1)

	}
}
