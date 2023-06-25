package fileManager

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
)

type csvManager[T any] struct {
	t      Transformer[T, []string]
	reader *csv.Reader
	writer *csv.Writer
	file   fs.File
}

func (c *csvManager[T]) Write(data T) error {
	writable := c.t.ToWritable(data)
	err := c.writer.Write(writable)
	if err == nil {
		c.writer.Flush()
	}
	return err
}

func (c *csvManager[T]) ReadLine() (T, error) {
	row, err := c.reader.Read()
	if err != nil {
		var dummy T
		return dummy, err
	}
	return c.t.FromWritable(row), nil
}

func (c *csvManager[T]) Read() ([]T, error) {
	returnable := make([]T, 0)
	for {
		row, err := c.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return returnable, err
		}
		returnable = append(returnable, c.t.FromWritable(row))
	}
	return returnable, nil
}

func (c *csvManager[T]) Close() {
	c.file.Close()
}

func (c *csvManager[T]) Clear() {
	data, _ := c.file.Stat()
	os.Truncate(data.Name(), 0)
}

// CreateCSVFileManager the transformer has to transform from T to string with the string being a comma separated string
func CreateCSVFileManager[T any](transformer Transformer[T, []string], nameFile string) (Manager[T], error) {
	f, err := os.OpenFile(nameFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("error while openning file: %v", err)
	}
	r := csv.NewReader(f)
	w := csv.NewWriter(f)
	return &csvManager[T]{
		t:      transformer,
		reader: r,
		writer: w,
		file:   f,
	}, nil
}
