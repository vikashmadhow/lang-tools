package regex

import (
	"cmp"
	"slices"
)

type (
	Range[E cmp.Ordered] struct {
		from E
		to   E
	}

	RangeSet[E cmp.Ordered] []Range[E]
)

//func (r RangeSet[E]) invert() charSetRange {
//    return all.minus(r)
//}

//func (r RangeSet[E]) minus(other RangeSet[E]) RangeSet[E] {
//    var result RangeSet[E]
//    r1 := r.compact()
//    r2 := other.compact()
//
//    j := 0
//    for _, left := range r1 {
//        for j < len(r2) && left.from > r2[j].to {
//            j++
//        }
//        if j == len(r2) || left.to < r2[j].from {
//        } else {
//            for j < len(r2) && left.to >= r2[j].from {
//                if left.from < r2[j].from {
//                    result = append(result, Range[E]{left.from, r2[j].from - 1})
//                }
//                left.from = min(left.to, r2[j].to+1)
//                j++
//            }
//            result = append(result, left)
//        }
//    }
//
//    return result
//}

func (r RangeSet[E]) compact() RangeSet[E] {
	if len(r) <= 1 {
		return r[:]
	}
	r.sort()
	result := RangeSet[E]{r[0]}
	for i := 1; i < len(r); i++ {
		last := result[len(result)-1]
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

func (r Range[E]) intersect(other Range[E]) bool {
	return r.to >= other.from && other.to >= r.from
}

func (r RangeSet[E]) sort() {
	slices.SortFunc(r, func(a, b Range[E]) int {
		if a.from < b.from {
			return -1
		} else if a.from > b.from {
			return 1
		} else {
			return 0
		}
	})
}
