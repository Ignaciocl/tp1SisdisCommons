package keyChecker

type Checker interface {
	IsKey(key string) bool
	AddKey(key string) error
	Clear()
	Close()
}
