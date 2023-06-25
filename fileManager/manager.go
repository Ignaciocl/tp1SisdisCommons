package fileManager

type Manager[T any] interface {
	Write(data T) error
	ReadLine() (T, error)
	Read() ([]T, error)
	Close()
	Clear()
}

type Transformer[T any, G any] interface {
	ToWritable(data T) G
	FromWritable(d G) T
}
