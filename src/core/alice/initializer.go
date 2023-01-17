package alice

import (
	"container/heap"
)

var Initializer = &InitializerType{
	fs:  nil,
	len: 0,
}

type InitializeWork struct {
	f func()
	p int
}

type InitializerType struct {
	fs  []InitializeWork
	len int
}

func (initializer *InitializerType) Len() int {
	return initializer.len
}

func (initializer *InitializerType) Less(i, j int) bool {
	return initializer.fs[i].p > initializer.fs[j].p
}

func (initializer *InitializerType) Swap(i, j int) {
	initializer.fs[i], initializer.fs[j] = initializer.fs[j], initializer.fs[i]
}

func (initializer *InitializerType) Pop() any {
	initializer.len--
	tail := &initializer.fs[initializer.len]
	initializer.fs = initializer.fs[:initializer.len]
	return tail
}

func (initializer *InitializerType) Push(i any) {
	initializer.fs = append(initializer.fs, i.(InitializeWork))
	initializer.len++
}

func (initializer *InitializerType) Register(f func(), p int) {
	heap.Push(initializer, InitializeWork{
		f: f,
		p: p,
	})
}

func (initializer *InitializerType) Initialize() {
	for initializer.Len() > 0 {
		f := initializer.Pop()
		f.(*InitializeWork).f()
	}
}
