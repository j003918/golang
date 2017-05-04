// list_stack project liststack.go
package stack

import (
	"container/list"
	"sync"
)

type Stack struct {
	list.List
	sync.Mutex
}

func New() *Stack {
	return new(Stack)

}

func (s *Stack) Empty() bool {
	return s.Len() == 0
}

func (s *Stack) Top() interface{} {
	e := s.Back()
	if e != nil {
		return e.Value
	}
	return nil
}

func (s *Stack) Pop() interface{} {
	e := s.Back()
	if e != nil {
		s.Remove(e)
		return e.Value
	}
	return nil
}

func (s *Stack) Push(v interface{}) {
	s.PushBack(v)
}

func (s *Stack) Clean() {
	//add "n" for modify list.remove bug??
	var n *list.Element
	for e := s.Front(); e != nil; e = n {
		n = e.Next()
		s.Remove(e)
	}
}
