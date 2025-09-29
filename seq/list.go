package seq

import "iter"

type List[T any] struct {
	Head T
	Tail *List[T]
}

func NewList[T any](head T, tail *List[T]) *List[T] {
	return &List[T]{
		Head: head,
		Tail: tail,
	}
}

func (l *List[T]) Push(head ...T) *List[T] {
	list := l
	for i := len(head) - 1; i >= 0; i-- {
		list = NewList(head[i], list)
	}
	return list
}

func (l *List[T]) Append(value T) *List[T] {
	if l.Tail == nil {
		return NewList(l.Head, NewList(value, nil))
	} else {
		return NewList(l.Head, l.Tail.Append(value))
	}
}

func (l *List[T]) Replace(newHead T) *List[T] {
	if l == nil {
		return &List[T]{Head: newHead, Tail: nil}
	} else {
		return &List[T]{newHead, l.Tail}
	}
}

func (l *List[T]) Empty() bool {
	return l == nil
}

func (l *List[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		for lst := l; lst != nil; lst = lst.Tail {
			if !yield(lst.Head) {
				break
			}
		}
	}
}

func (l *List[T]) Len() int {
	n := 0
	for lst := l; lst != nil; lst = lst.Tail {
		n++
	}
	return n
}

func (l *List[T]) Reverse() *List[T] {
	if l.Tail == nil {
		return l
	} else {
		return l.Tail.Reverse().Append(l.Head)
	}
}
