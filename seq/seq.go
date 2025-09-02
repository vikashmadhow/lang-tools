// Package seq implements a set of functions to work with the iter.Pull and iter.Pull2
// versions of iter.Seq and iter.Seq2, respectively.
package seq

import "iter"

// Seq is a function that returns a value of type V and a bool which is true if the
// sequence has more values. This is compatible with the next value function of the
// iter.Pull version of an iter.Seq.
type Seq[V any] func() (V, bool)

// Seq2 is a function that returns a key-value pair of type (K, V) and a bool which is
// true if the sequence has more values. This is compatible with the next value function
// of the iter.Pull2 version of an iter.Seq2.
type Seq2[K, V any] func() (K, V, bool)

// FilterFunc is a function that can be used to filter a sequence through seq.Filter.
// It takes a value of type V and returns true if the value must be preserved or false
// to remove the value from the sequence.
type FilterFunc[V any] func(V) bool

// Filter2Func is a function that can be used to filter a sequence through seq.Filter2.
// It takes a key-value of type (K, V) and returns true if the value must be preserved
// or false to remove the value from the sequence.
type Filter2Func[K, V any] func(K, V) bool

// MapFunc is a function to convert values of type U to values of type V. It can be applied
// to values of a sequence through the function seq.Map.
type MapFunc[U, V any] func(U) V

// Map2Func is a function to convert key-value pairs of type (K, V) to values of type (MK, MV).
// It can be applied to values of a sequence through the function seq.Map2.
type Map2Func[K, V, MK, MV any] func(K, V) (MK, MV)

// Pair is a struct holding a key-value pair. This is used by the FlatMap2Func to transform
// a key-value to a sequence of key-value pairs (which are then flattened).
type Pair[A, B any] struct {
	A A
	B B
}

// FlatMapFunc is a function provided to seq.FlatMap to transform a value of type U to a sequence
// of type V ([]V). The sequence is then flattened into a Seq[V].
type FlatMapFunc[U, V any] func(U) []V

// FlatMap2Func is a function provided to seq.FlatMap2 to transform a key-value pair of type (K, V)
// to a sequence of type (MK, MV) ([]seq.KeyPair[MK, MV], to be exact). The sequence is then flattened
// into a Seq2[MK, MV].
type FlatMap2Func[K, V, MK, MV any] func(K, V) []Pair[MK, MV]

// Push converts an iter.Pull style iterator to the default iterator (push) style.
func Push[V any](seq Seq[V], stop func()) iter.Seq[V] {
	return func(yield func(value V) bool) {
		defer stop()
		for v, ok := seq(); ok; v, ok = seq() {
			if !yield(v) {
				break
			}
		}
	}
}

// Push2 converts an iter.Pull2 style iterator to the default iterator (push) style.
func Push2[K, V any](seq Seq2[K, V], stop func()) iter.Seq2[K, V] {
	return func(yield func(key K, value V) bool) {
		defer stop()
		for k, v, ok := seq(); ok; k, v, ok = seq() {
			if !yield(k, v) {
				break
			}
		}
	}
}
