package store

type Store interface {
	Get(string) string
	Put(string, string)
	Delete(string)
}
