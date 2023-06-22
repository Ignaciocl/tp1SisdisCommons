package fileManager

import (
	"errors"
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/utils"
	"io"
	"os"
)

type Storable interface {
	GetId() int64
	SetId(id int64)
}

type db[T Storable] struct {
	writer   io.WriteSeeker
	reader   io.Reader
	closer   io.Closer
	agus     Transformer[T, []byte]
	sizeLine int64
	saved    int64
	endLine  []byte
}

func createByteArray(endLine []byte, size int64) []byte {
	b := append(endLine, make([]byte, size-1-int64(len(endLine)))...)
	return append(b, []byte("\n")...)
}
func (d *db[T]) Write(data T) error {
	isNew := true
	if data.GetId() > 0 {
		_, err := d.writer.Seek((data.GetId()-1)*d.sizeLine, io.SeekStart)
		if err != nil {
			return err
		}
		isNew = false
	}
	dataToSave := d.agus.ToWritable(data)
	msgSize := int64(len(dataToSave))
	if endLineSize := int64(len(d.endLine)); msgSize > d.sizeLine-endLineSize-1 {
		return fmt.Errorf("size of data to stored far exceeds the limit set at the constructor which is: %d, size of message: %d, end batch size: %d", d.sizeLine-1, msgSize, endLineSize)
	}
	dataToSave = append(dataToSave, createByteArray(d.endLine, d.sizeLine-msgSize)...)
	if i, err := d.writer.Write(dataToSave); err != nil {
		utils.LogError(err, fmt.Sprintf("could not save data to db, bytes stored: %d, if the number is greater than 0 stop execution", i))
		return err
	}
	if isNew {
		d.saved += 1
		data.SetId(d.saved)
	} else {
		_, err := d.writer.Seek(0, io.SeekEnd)
		utils.FailOnError(err, "an impossibilitie happened, and licha told me to catch all and all errors")
	}
	return nil
}

func (d *db[T]) ReadLine() (T, error) {
	var dataParsed T
	data := make([]byte, d.sizeLine)
	if i, err := d.reader.Read(data); err != nil {
		if errors.Is(err, io.EOF) {
			return dataParsed, err
		}
		return dataParsed, fmt.Errorf("could not read a line from the file, err is: %v, amount read: %d", err, i)
	} else {
		return d.agus.FromWritable(data), nil
	}
}

func (d *db[T]) Read() ([]T, error) {
	data := make([]T, 0)
	counter := int64(0)
	for {
		line, err := d.ReadLine()
		if err != nil {
			return data, err
		}
		line.SetId(counter)
		counter += 1
		data = append(data, line)
	}
}

func (d *db[T]) Close() {
	d.closer.Close()
}

func CreateDB[T Storable](trans Transformer[T, []byte], fileName string, sizeLine int64, endLine string) (Manager[T], error) {
	if sizeLine <= 1 {
		return nil, fmt.Errorf("invalid amount of size of line, should be bigger than 1, %d received", sizeLine)
	}
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("error while openning file: %v", err)
	}
	d := db[T]{
		agus:     trans,
		sizeLine: sizeLine,
		saved:    0,
		reader:   f,
		writer:   f,
		closer:   f,
		endLine:  []byte(endLine),
	}
	return &d, nil
}
