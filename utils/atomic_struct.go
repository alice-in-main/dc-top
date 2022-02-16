package utils

import "sync"

type Mutator func(*Lockable) error

type Lockable interface {
	Alter(Mutator)
}

type Locked struct {
	lock sync.Mutex
	obj  Lockable
}

func NewLocked(obj Lockable) Locked {
	return Locked{
		lock: sync.Mutex{},
		obj:  obj,
	}
}

func (l *Locked) Get() Lockable {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.obj
}

func (l *Locked) Set(new_obj *Lockable) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.obj = *new_obj
}

func (l *Locked) Modify(f Mutator) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.obj.Alter(f)
}
