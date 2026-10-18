package main

import (
	"fmt"
	"slices"
)

func cb(newVal int) {
	fmt.Println("callback babay", newVal)
}

func main() {
	a := Atom[int]{}
	id := a.Subscribe(cb)
	a.Set(3)
	a.Set(4)
	a.Unsubscribe(id)
	val := a.Get()
	fmt.Println(val)
}

type Subscriber[T any] struct {
	id int
	cb func(value T)
}

type Atom[T any] struct {
	value       T
	subscribers []Subscriber[T]
	id          int
}

func (a *Atom[T]) Subscribe(cb func(val T)) int {
	currentId := a.id
	a.id++
	a.subscribers = append(a.subscribers, Subscriber[T]{id: currentId, cb: cb})
	return currentId
}

func (a *Atom[T]) Set(val T) {
	a.value = val
	for _, s := range a.subscribers {
		s.cb(a.value)
	}
}

func (a *Atom[T]) Get() T {
	return a.value
}

func (a *Atom[T]) Unsubscribe(id int) {
	for index, subscriber := range a.subscribers {
		if id == subscriber.id {
			a.subscribers = slices.Delete(a.subscribers, index, index+1)
			return
		}
	}
}
