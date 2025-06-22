package store

import (
	"fmt"
	"sync"
)

var s = make(map[string]string)
var m = sync.RWMutex{}

func Add(key, value string) error {
	m.Lock()
	defer m.Unlock()

	s[key] = value

	return nil
}

func Get(key string) (string, error) {
	m.Lock()
	defer m.Unlock()

	val, exists := s[key]

	if !exists {
		return "", fmt.Errorf("element doesn't exist")
	}

	return val, nil
}

func Delete(key string) error {
	m.Lock()
	defer m.Unlock()

	_, exists := s[key]
	if !exists {
		return fmt.Errorf("element doesn't exist")
	}
	delete(s, key)
	return nil
}
