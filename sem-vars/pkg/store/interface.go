package store

import "fmt"

type Store interface {
	Get(string) (string, error)
	Put(string, string) error
}

var NotFoundErr = fmt.Errorf("key not found")
