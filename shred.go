package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

func Shred(path string) error {
	return shredFile(path, true)
}

func shredFile(path string, deleteAfter bool) error {
	return shredFileWithOptions(path, true, deleteAfter)
}

func shredFileWithOptions(path string, truncateAfter bool, deleteAfter bool) error {

	fmt.Println("Shredding file:", path)

	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		fmt.Println("Error opening file:", path)

		return err
	}
	defer file.Close()

	lstat, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", path)
		return err
	}

	if size := lstat.Size(); size > 0 {
		err = overWriteWithRandomData(file, size, 3)
		if err != nil {
			return err
		}
	}

	if truncateAfter {
		err = file.Truncate(0)
		if err != nil {
			return err
		}

		err = file.Sync()
		if err != nil {
			return err
		}
	}

	file.Close()

	if deleteAfter {
		return os.Remove(path)
	}
	return nil
}

func overWriteWithRandomData(file *os.File, size int64, passes int) error {
	for i := 0; i < passes; i++ {
		_, err := file.Seek(0, 0)
		if err != nil {
			return err
		}

		_, err = io.CopyN(file, rand.Reader, size)
		if err != nil {
			return err
		}

		err = file.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}
