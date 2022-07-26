package store

type Store interface {
	Get(key, contextId string) (string, error)
	Put(key, value, contextId string) error
	Delete(key, contextId string) error
	CheckIfKeyDeleted(key, contextId string) (bool, error)
}
