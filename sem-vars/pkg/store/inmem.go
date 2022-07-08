package store

type InMem struct {
	storage map[string]string
}

var _ Store = NewInMem()

func NewInMem() *InMem {
	s := &InMem{}
	s.storage = make(map[string]string)

	return s
}

func (s *InMem) Put(key string, value string) error {
	s.storage[key] = value
	return nil
}

func (s *InMem) Get(key string) (string, error) {
	val, ok := s.storage[key]

	if !ok {
		return "", NotFoundErr
	}

	return val, nil
}
