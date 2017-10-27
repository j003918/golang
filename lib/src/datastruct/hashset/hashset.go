// hashset project hashset.go
package hashset

var nilVal = struct{}{}

type HashSet struct {
	hashMap map[interface{}]struct{}
}

func New() *HashSet {
	return &HashSet{hashMap: make(map[interface{}]struct{}, 1000)}
}

func (set *HashSet) Set(key interface{}) {
	set.hashMap[key] = nilVal
}

func (set *HashSet) Del(key interface{}) {
	delete(set.hashMap, key)
}

func (set *HashSet) Contains(key interface{}) bool {
	_, contains := set.hashMap[key]
	return contains
}
