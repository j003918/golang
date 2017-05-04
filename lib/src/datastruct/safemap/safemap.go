// safemap project safemap.go
package safemap

import (
	"fmt"
	"sync"
)

type SafeMap struct {
	lock *sync.RWMutex
	mp   map[interface{}]interface{}
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		lock: new(sync.RWMutex),
		mp:   make(map[interface{}]interface{}),
	}
}

func (sm *SafeMap) Println() {
	fmt.Println(sm.mp)
}

func (sm *SafeMap) Get(key interface{}) interface{} {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	if _, ok := sm.mp[key]; ok {
		return sm.mp[key]
	} else {
		return nil
	}
}

func (sm *SafeMap) Set(key, val interface{}) {
	sm.lock.Lock()
	sm.mp[key] = val
	sm.lock.Unlock()
}

func (sm *SafeMap) Del(key interface{}) {
	sm.lock.Lock()
	delete(sm.mp, key)
	sm.lock.Unlock()
}

func (sm *SafeMap) LoopDel(isDel func(k, v interface{}) bool) {
	sm.lock.Lock()
	for k, v := range sm.mp {
		if isDel(k, v) {
			sm.Del(k)
		}
	}
	sm.lock.Unlock()
}

func (sm *SafeMap) Check(key interface{}) bool {
	sm.lock.RLock()
	_, ok := sm.mp[key]
	sm.lock.RUnlock()
	return ok
}

func (sm *SafeMap) CheckWithUpdate(key, val interface{}) bool {
	sm.lock.Lock()
	_, ok := sm.mp[key]
	if ok {
		sm.mp[key] = val
	}
	sm.lock.Unlock()
	return ok
}
