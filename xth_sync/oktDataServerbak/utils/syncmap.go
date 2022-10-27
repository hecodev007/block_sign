package utils

import "sync"

type SyncMap struct {
	M map[string]string
	L sync.RWMutex
}

func (s *SyncMap) Del(key string) {
	s.L.Lock()
	defer s.L.Unlock()
	delete(s.M, key)
}

func (s *SyncMap) Store(key string, value string) {
	s.L.Lock()
	defer s.L.Unlock()
	s.M[key] = value
}

func (s *SyncMap) Load(key string) (value string) {
	s.L.RLock()
	defer s.L.RUnlock()
	return s.M[key]
}
