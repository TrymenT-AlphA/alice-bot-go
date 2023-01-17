package util

import (
	"bufio"
	"os"
)

func Write(filename string, data []byte) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	_, err = writer.Write(data)
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
