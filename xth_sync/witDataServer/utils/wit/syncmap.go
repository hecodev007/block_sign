package wit

import "sync"

type SyncMap struct {
	M map[int64]chan []byte
	L sync.RWMutex
}

func NewSyncMap() *SyncMap {
	return &SyncMap{
		M: make(map[int64]chan []byte),
		L: sync.RWMutex{},
	}
}
func (s *SyncMap) Del(key int64) {
	s.L.Lock()
	defer s.L.Unlock()
	delete(s.M, key)
}

func (s *SyncMap) Store(key int64, value chan []byte) {
	s.L.Lock()
	defer s.L.Unlock()
	s.M[key] = value
}

func (s *SyncMap) Load(key int64) (value chan []byte) {
	s.L.RLock()
	defer s.L.RUnlock()
	return s.M[key]
}
