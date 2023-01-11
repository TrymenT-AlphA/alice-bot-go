package model

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Task struct {
	Name string
	Url  string
	Type string
}

func (task *Task) Download(dir string) error {
	out, err := os.OpenFile(
		filepath.Join(dir, fmt.Sprintf("%s.%s", task.Name, task.Type)),
		os.O_WRONLY|os.O_CREATE, 0666,
	)
	if err != nil {
		return err
	}

	response, err := http.Get(task.Url)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = response.Body.Close()
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(out)

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
