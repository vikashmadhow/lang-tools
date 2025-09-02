package seq

import "iter"

// Filter is a function that filters elements of type V in the pull-type version of
// an iter.Seq using a function V -> bool (the predicate). Elements for which the
// predicate returns false are filtered out of the sequence.
func Filter[V any, F ~func(V) bool](seq Seq[V], predicate F) Seq[V] {
	return func() (V, bool) {
		for {
			value, valid := seq()
			if !valid {
				return value, valid
			}
			if predicate(value) {
				return value, true
			}
		}
	}
}

// FilterSeq is the Filter function designed to work on the push-style iter.Seq.
func FilterSeq[V any, F ~func(V) bool](seq iter.Seq[V], predicate F) iter.Seq[V] {
	return func(yield func(value V) bool) {
		for v := range seq {
			if predicate(v) {
				if !yield(v) {
					break
				}
			}
		}
	}
}

// Filter2 is the Filter function designed to work on the 2-elements version of the
// pull-style version of iter.Seq2.
func Filter2[K, V any, F ~func(K, V) bool](seq Seq2[K, V], predicate F) Seq2[K, V] {
	return func() (K, V, bool) {
		for {
			key, value, valid := seq()
			if !valid {
				return key, value, valid
			}
			if predicate(key, value) {
				return key, value, true
			}
		}
	}
}

// FilterSeq2 is the Filter2 function designed to work on the push-style iter.Seq.
func FilterSeq2[K, V any, F ~func(K, V) bool](seq iter.Seq2[K, V], predicate F) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range seq {
			if predicate(k, v) {
				if !yield(k, v) {
					break
				}
			}
		}
	}
}
