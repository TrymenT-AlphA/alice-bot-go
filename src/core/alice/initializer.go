package alice

import (
	"container/heap"
)

var Init = &Initializer{
	fs:  nil,
	len: 0,
}

type InitElem struct {
	f func()
	p int
}

type Initializer struct {
	fs  []InitElem
	len int
}

func (initializer *Initializer) Len() int {
	return initializer.len
}

func (initializer *Initializer) Less(i, j int) bool {
	return initializer.fs[i].p > initializer.fs[j].p
}

func (initializer *Initializer) Swap(i, j int) {
	initializer.fs[i], initializer.fs[j] = initializer.fs[j], initializer.fs[i]
}

func (initializer *Initializer) Pop() any {
	initializer.len--
	tail := &initializer.fs[initializer.len]
	initializer.fs = initializer.fs[:initializer.len]
	return tail
}

func (initializer *Initializer) Push(x any) {
	initializer.fs = append(initializer.fs, x.(InitElem))
	initializer.len++
}

func (initializer *Initializer) Register(f func(), p int) {
	heap.Push(initializer, InitElem{
		f: f,
		p: p,
	})
}

func (initializer *Initializer) Initialize() {
	for initializer.Len() > 0 {
		f := initializer.Pop()
		f.(*InitElem).f()
	}
}
