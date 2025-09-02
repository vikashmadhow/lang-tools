package seq

import "iter"

// Map is a function that maps elements of type U in the pull-type version of
// an iter.Seq to elements of type V using a function U -> V.
func Map[U, V any, F ~func(U) V](seq Seq[U], mapper F) Seq[V] {
	return func() (V, bool) {
		value, valid := seq()
		return mapper(value), valid
	}
}

// MapSeq is the Map function for push-style iter.Seq.
func MapSeq[U, V any, F ~func(U) V](seq iter.Seq[U], mapper F) iter.Seq[V] {
	return func(yield func(V) bool) {
		for value := range seq {
			if !yield(mapper(value)) {
				break
			}
		}
	}
}

// Map2 is the Map function for the 2-elements version of the pull-style version of iter.Seq2.
func Map2[K, V, MK, MV any, F ~func(K, V) (MK, MV)](seq Seq2[K, V], mapper F) Seq2[MK, MV] {
	return func() (MK, MV, bool) {
		key, value, valid := seq()
		mk, mv := mapper(key, value)
		return mk, mv, valid
	}
}

// MapSeq2 is the Map2 function for push-style version of iter.Seq2.
func MapSeq2[K, V, MK, MV any, F ~func(K, V) (MK, MV)](seq iter.Seq2[K, V], mapper F) iter.Seq2[MK, MV] {
	return func(yield func(MK, MV) bool) {
		for key, value := range seq {
			mk, mv := mapper(key, value)
			if !yield(mk, mv) {
				break
			}
		}
	}
}

// FlatMap is a function that maps elements of type U in the pull-type version of
// an iter.Seq to elements of type []V using a function U -> []V, and then flattening
// the result slice into a sequence of V. Thus, similar to Map, FlatMap takes a Seq[U]
// and produces a Seq[V], but the mapper function can produce zero of more V values for
// each original U value, unlike Map where each U can only be mapped to only one V.
func FlatMap[U, V any, F ~func(U) []V](seq Seq[U], mapper F) Seq[V] {
	remaining := make(chan V, 100)
	return func() (V, bool) {
		for len(remaining) == 0 {
			value, valid := seq()
			if valid {
				mapped := mapper(value)
				for _, v := range mapped {
					remaining <- v
				}
			} else {
				break
			}
		}
		for len(remaining) > 0 {
			return <-remaining, true
		}
		close(remaining)
		return *new(V), false
	}
}

// FlatMapSeq is the FlatMap function for push-style iter.Seq.
func FlatMapSeq[U, V any, F ~func(U) []V](seq iter.Seq[U], mapper F) iter.Seq[V] {
	remaining := make(chan V, 100)
	return func(yield func(V) bool) {
		for value := range seq {
			mapped := mapper(value)
			for _, v := range mapped {
				remaining <- v
			}
			if len(remaining) > 0 {
				if !yield(<-remaining) {
					break
				}
			}
		}
		for len(remaining) > 0 {
			if !yield(<-remaining) {
				break
			}
		}
		close(remaining)
	}
}

// FlatMap2 is the FlatMap function for the 2-elements version of the pull-style version of iter.Seq2.
// The mapper for FlatMap2 is a function F(A, B) -> []Pair[MA, MB] where the pair (A, B) is mapped to
// zero or more pairs of (MA, MB).
func FlatMap2[A, B, MA, MB any, F ~func(A, B) []Pair[MA, MB]](seq Seq2[A, B], mapper F) Seq2[MA, MB] {
	remaining := make(chan Pair[MA, MB], 100)
	return func() (MA, MB, bool) {
		for len(remaining) == 0 {
			a, b, valid := seq()
			if valid {
				mapped := mapper(a, b)
				for _, v := range mapped {
					remaining <- v
				}
			} else {
				break
			}
		}
		for len(remaining) > 0 {
			pair := <-remaining
			return pair.A, pair.B, true
		}
		close(remaining)
		return *new(MA), *new(MB), false
	}
}

// FlatMapSeq2 is the FlatMap2 function for push-style version of iter.Seq2.
func FlatMapSeq2[A, B, MA, MB any, F ~func(A, B) []Pair[MA, MB]](seq iter.Seq2[A, B], mapper F) iter.Seq2[MA, MB] {
	remaining := make(chan Pair[MA, MB], 100)
	return func(yield func(MA, MB) bool) {
		for key, value := range seq {
			mapped := mapper(key, value)
			for _, v := range mapped {
				remaining <- v
			}
			if len(remaining) > 0 {
				pair := <-remaining
				if !yield(pair.A, pair.B) {
					break
				}
			}
		}
		for len(remaining) > 0 {
			pair := <-remaining
			if !yield(pair.A, pair.B) {
				break
			}
		}
		close(remaining)
	}
}
