package store

import (
	"fmt"
	"sync"
)

var ErrNotFound = fmt.Errorf("element not found")

func InitStore() Store {
	return &storeStruct{
		s: make(map[string]string),
		m: sync.RWMutex{},
	}
}

type Store interface {
	Add(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}

type storeStruct struct {
	s map[string]string
	m sync.RWMutex
}

func (s storeStruct) Add(key, value string) error {
	s.m.Lock()
	defer s.m.Unlock()

	s.s[key] = value

	return nil
}

func (s storeStruct) Get(key string) (string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	val, exists := s.s[key]

	if !exists {
		return "", ErrNotFound
	}

	return val, nil
}

func (s storeStruct) Delete(key string) error {
	s.m.Lock()
	defer s.m.Unlock()

	_, exists := s.s[key]
	if !exists {
		return ErrNotFound
	}
	delete(s.s, key)
	return nil
}
