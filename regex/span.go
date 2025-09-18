package regex

import (
	"math/rand"
	"slices"
	"unicode/utf8"
)

type (
	span struct {
		from rune
		to   rune
	}

	spanSet []span
)

var (
	allUnicode     = spanSet{span{0, utf8.MaxRune}}
	asciiPrintable = spanSet{span{32, 126}}
)

func (r span) len() int {
	return int(r.to) - int(r.from) + 1
}

func (r span) random() rune {
	return rune(int(r.from) + rand.Intn(r.len()))
}

func (r span) intersect(other span) bool {
	return r.to >= other.from && other.to >= r.from
}

func (r span) match(c rune) bool {
	return r.from <= c && c <= r.to
}

func (r spanSet) len() int {
	l := 0
	for _, s := range r {
		l += s.len()
	}
	return l
}

func (r spanSet) random() rune {
	n := rand.Intn(r.len())
	for _, s := range r {
		count := s.len()
		if n < count {
			return s.random()
		}
		n -= count
	}
	return 0
}

func (r spanSet) invertUnicode() spanSet {
	return allUnicode.minus(r)
}

func (r spanSet) invertAsciiPrintable() spanSet {
	return asciiPrintable.minus(r)
}

func (r spanSet) minus(other spanSet) spanSet {
	var result spanSet
	r1 := r.compact()
	r2 := other.compact()

	j := 0
	for _, left := range r1 {
		for j < len(r2) && left.from > r2[j].to {
			j++
		}
		if j == len(r2) || left.to < r2[j].from {
			result = append(result, left)
		} else {
			reachedEnd := false
			for j < len(r2) && left.to >= r2[j].from {
				if left.from < r2[j].from {
					result = append(result, span{left.from, r2[j].from - 1})
				}
				if left.to <= r2[j].to {
					reachedEnd = true
				} else {
					left.from = r2[j].to + 1
				}
				j++
			}
			if !reachedEnd {
				result = append(result, left)
			}
		}
	}

	return result
}

func (r spanSet) compact() spanSet {
	if len(r) <= 1 {
		return r[:]
	}
	r.sort()
	result := spanSet{r[0]}
	for i := 1; i < len(r); i++ {
		last := &result[len(result)-1]
		if last.intersect(r[i]) {
			if last.to < r[i].to {
				last.to = r[i].to
			}
		} else {
			result = append(result, r[i])
		}
	}
	return result
}

func (r spanSet) sort() {
	slices.SortFunc(r, func(a, b span) int {
		return int(a.from) - int(b.from)
	})
}

func (r spanSet) match(c rune) bool {
	if len(r) < 5 {
		for _, s := range r {
			if s.match(c) {
				return true
			}
		}
	} else {
		lo, hi := 0, len(r)-1
		for lo <= hi {
			mid := (lo + hi) / 2
			if r[mid].match(c) {
				return true
			}
			if r[mid].from < c {
				lo = mid + 1
			} else {
				hi = mid
			}
		}
	}
	return false
}
