package utils

import "sync"

var maplock map[string]*sync.Mutex
var lock sync.Mutex
func init(){
	maplock = make(map[string]*sync.Mutex)
}
func Lock(addr string){
	lock.Lock()

	addlock,ok := maplock[addr]
	if !ok {
		maplock[addr]= new(sync.Mutex)
	}
	addlock = maplock[addr]

	lock.Unlock()

	addlock.Lock()
}
func Unlock(addr string){
	lock.Lock()
	addlock := maplock[addr]
	lock.Unlock()
	addlock.Unlock()
}