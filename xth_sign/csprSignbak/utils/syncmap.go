package utils

import "sync"
var syncMap *SyncMap

func init(){
	syncMap = NewSyncMap()
}
func Get(addr string)uint64{
	return syncMap.Get(addr)
}
func Set(addr string,v uint64){
	syncMap.Set(addr,v)
}
func Add(addr string,n uint64){
	syncMap.Add(addr,n)
}

func NewSyncMap() *SyncMap  {
	return &SyncMap{
		M:make(map[string]uint64),
		L:sync.RWMutex{},
	}
}
type SyncMap struct {
	M map[string]uint64
	L sync.RWMutex
}
func (s *SyncMap) Set(key string,v uint64){
	s.L.Lock()
	defer s.L.Unlock()
	s.M[key] += v
}
func (s *SyncMap) Add(key string, n uint64) {
	s.L.Lock()
	defer s.L.Unlock()
	s.M[key] += n
}

func (s *SyncMap) Get(key string) uint64 {
	s.L.RLock()
	defer s.L.RUnlock()
	return s.M[key]
}
