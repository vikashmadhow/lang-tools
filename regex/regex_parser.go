// author: Vikash Madhow (vikash.madhow@gmail.com)

// Parses regular expression to this grammar:
//
//	re -> re '|' re
//	    | re ('*' | '+' | '?')
//	    | re re
//	    | '(' re ')'
//	    | ch
//
//	ch -> '[' '^'? (c ['-' c])+ ']'
//	    | c
//	    | '\' ('*' | '+' | '?' | '|' | '(' | ')' | '[' | ']')
//
//	Refactored to remove left-recursion and ambiguity:
//	using: A = Aa|B  =>  A  = BA'
//	                     A' = aA'|e
//
//	regex  = term ['|' regex]
//	term   = { factor }
//	factor = base [('*' | '+' | '?')]
//	base   = '(' regex ')'
//	       | ch

package regex

import (
	"maps"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
)

// ----------------Regex top-down parsing----------------//
type (
	parser struct {
		input    []rune
		position int
		group    *int
		groups   set[int]
	}

	modifier struct {
		caseInsensitive bool
		unicode         bool
	}
)

func (r *parser) peek() rune {
	if r.position < len(r.input) {
		return r.input[r.position]
	} else {
		return 0
	}
}

func (r *parser) next() rune {
	c := r.input[r.position]
	r.position++
	return c
}

func (r *parser) hasMore() bool {
	return len(r.input) > r.position
}

func (r *parser) regex(mod *modifier) Pattern {
	term := r.term(mod)
	if r.hasMore() && r.peek() == '|' {
		r.next()
		right := r.regex(mod)
		return &choice{term, right}
	} else {
		return term
	}
}

func (r *parser) term(mod *modifier) Pattern {
	var factors []Pattern
	for r.hasMore() && r.peek() != ')' && r.peek() != '|' {
		f := r.factor(mod)
		if f != nil {
			factors = append(factors, f)
		}
	}
	return &sequence{factors}
}

func (r *parser) factor(mod *modifier) Pattern {
	base := r.base(mod)
	if r.hasMore() {
		switch r.peek() {
		case '*':
			r.next()
			return &zeroOrMore{base}
		case '+':
			r.next()
			return &oneOrMore{base}
		case '?':
			r.next()
			return &zeroOrOne{base}
		case '{':
			r.next()
			m := ""
			n := ""
			first := true
			if r.hasMore() {
				for r.hasMore() {
					c := r.next()
					if c == '}' {
						break
					}
					if c == ',' {
						first = false
					} else if first {
						m += string(c)
					} else {
						n += string(c)
					}
				}
				var mi, ma int32
				if len(strings.TrimSpace(m)) == 0 {
					mi = 0
				} else {
					x, err := strconv.Atoi(m)
					if err != nil {
						mi = 0
					} else {
						mi = int32(x)
					}
				}
				if len(strings.TrimSpace(n)) == 0 {
					ma = math.MaxUint8
				} else {
					x, err := strconv.Atoi(n)
					if err != nil {
						ma = math.MaxUint8
					} else {
						ma = int32(x)
					}
				}
				if first {
					ma = mi
				}
				mi = min(math.MaxUint8, max(0, mi))
				ma = min(math.MaxUint8, max(0, ma))
				if mi > ma {
					ma = atomic.SwapInt32(&mi, ma)
				}
				return &repeat{base, uint8(mi), uint8(ma)}
			} else {
				//return &singleChar{mod, '{', cp(r.groups)}
				return chars(mod, r.groups, false, span{'{', '{'})
			}
		}
	}
	return base
}

func (r *parser) base(mod *modifier) Pattern {
	if r.peek() == '(' {
		r.next()
		if r.peek() == '?' {
			// modifiers
			r.next()
			if r.hasMore() {
				switch r.next() {
				case 'i':
					mod.caseInsensitive = true
				case 'u':
					mod.unicode = true
				}
			}

			// lenient parsing: don't break if no closing bracket, read to the end
			if r.hasMore() {
				r.next()
			}
			return nil
		} else if r.peek() == ':' {
			// list
			r.next()
			var conversion conversion
			var listName strings.Builder
			for r.hasMore() && r.peek() != ')' && r.peek() != ':' {
				listName.WriteRune(r.next())
			}
			if r.peek() == ':' {
				r.next()
				for r.hasMore() && r.peek() != ')' {
					switch r.next() {
					case 'l':
						conversion.lower = true
					case 'u':
						conversion.upper = true
					case 't':
						conversion.title = true
					case 's':
						conversion.singleSpace = true
					case 'm':
						conversion.trim = true
					}
				}
			}
			if r.hasMore() {
				r.next()
			}
			//return &inList{
			//	mod:     mod,
			//	list:    listName.String(),
			//	convert: conversion,
			//}
			return &char{
				modifier: mod,
				list:     listName.String(),
				convert:  &conversion,
			}
		} else {
			*r.group++
			g := *r.group
			r.groups[g] = true

			re := r.regex(mod)
			//r.groups.Remove(r.groups.Back())
			delete(r.groups, g)

			// lenient parsing: don't break if no closing bracket, read to the end
			if r.hasMore() {
				r.next()
			}
			return &captureGroup{re}
		}
	} else {
		return r.ch(mod)
	}
}

func (r *parser) ch(mod *modifier) Pattern {
	if r.peek() == '[' {
		r.next()

		exclude := false
		if r.peek() == '^' {
			r.next()
			exclude = true
		}

		//charSets := list.New()
		var charSets spanSet
		for r.hasMore() && r.peek() != ']' {
			from := r.next()
			if r.peek() == '-' {
				r.next()
				if r.hasMore() && r.peek() != ']' {
					to := r.next()
					//charSets.PushBack(&charRange{mod, from, to, cp(r.groups)})
					charSets = append(charSets, span{from, to})
				} else {
					//charSets.PushBack(&charRange{mod, from, math.MaxUint8, cp(r.groups)})
					charSets = append(charSets, span{from, math.MaxUint8})
				}
			} else {
				//charSets.PushBack(&singleChar{mod, from, cp(r.groups)})
				charSets = append(charSets, span{from, from})
			}
		}
		// lenient parsing: don't break if no closing square bracket, read to the end
		if r.hasMore() {
			r.next()
		}
		//return ch{mod, exclude, *charSets, cp(r.groups), nil}
		return chars(mod, r.groups, exclude, charSets...)

	} else if r.peek() == '\\' {
		r.next()
		// lenient parsing: a single backlash at the end is interpreted as escaping itself
		if r.hasMore() {
			switch c := r.next(); c {
			case 'd':
				//return &charRange{mod, '0', '9', cp(r.groups)}
				return chars(mod, r.groups, false, span{'0', '9'})
			case 'D':
				//cs := list.New()
				//cs.PushBack(&charRange{mod, '0', '9', cp(r.groups)})
				//return &charSet{mod, true, *cs, cp(r.groups), nil}
				return chars(mod, r.groups, true, span{'0', '9'})

			case 's':
				//cs := list.New()
				//cs.PushBack(&singleChar{mod, ' ', cp(r.groups)})
				//cs.PushBack(&singleChar{mod, '\t', cp(r.groups)})
				//cs.PushBack(&singleChar{mod, '\n', cp(r.groups)})
				//cs.PushBack(&singleChar{mod, '\f', cp(r.groups)})
				//cs.PushBack(&singleChar{mod, '\r', cp(r.groups)})
				//return &charSet{mod, false, *cs, cp(r.groups), nil}
				return chars(mod, r.groups, false,
					span{' ', ' '},
					span{'\t', '\t'},
					span{'\n', '\n'},
					span{'\f', '\f'},
					span{'\r', '\r'})

			case 'S':
				//cs := list.New()
				//cs.PushBack(&singleChar{mod, ' ', cp(r.groups)})
				//cs.PushBack(&singleChar{mod, '\t', cp(r.groups)})
				//cs.PushBack(&singleChar{mod, '\n', cp(r.groups)})
				//cs.PushBack(&singleChar{mod, '\f', cp(r.groups)})
				//cs.PushBack(&singleChar{mod, '\r', cp(r.groups)})
				//return &charSet{mod, true, *cs, cp(r.groups), nil}
				return chars(mod, r.groups, true,
					span{' ', ' '},
					span{'\t', '\t'},
					span{'\n', '\n'},
					span{'\f', '\f'},
					span{'\r', '\r'})

			case 'w':
				//cs := list.New()
				//cs.PushBack(&charRange{mod, '0', '9', cp(r.groups)})
				//cs.PushBack(&charRange{mod, 'a', 'z', cp(r.groups)})
				//cs.PushBack(&charRange{mod, 'A', 'Z', cp(r.groups)})
				//cs.PushBack(&singleChar{mod, '_', cp(r.groups)})
				//return &charSet{mod, false, *cs, cp(r.groups), nil}
				return chars(mod, r.groups, false,
					span{'0', '9'},
					span{'a', 'z'},
					span{'A', 'Z'},
					span{'_', '_'})

			case 'W':
				//cs := list.New()
				//cs.PushBack(&charRange{mod, '0', '9', cp(r.groups)})
				//cs.PushBack(&charRange{mod, 'a', 'z', cp(r.groups)})
				//cs.PushBack(&charRange{mod, 'A', 'Z', cp(r.groups)})
				//cs.PushBack(&singleChar{mod, '_', cp(r.groups)})
				//return &charSet{mod, true, *cs, cp(r.groups), nil}
				return chars(mod, r.groups, true,
					span{'0', '9'},
					span{'a', 'z'},
					span{'A', 'Z'},
					span{'_', '_'})

			default:
				//return &singleChar{mod, c, cp(r.groups)}
				return chars(mod, r.groups, false, span{c, c})
			}
		} else {
			//return &singleChar{mod, '\\', cp(r.groups)}
			return chars(mod, r.groups, false, span{'\\', '\\'})
		}
	} else if r.peek() == '.' {
		r.next()
		//return &anyChar{mod: mod}
		return chars(mod, r.groups, false, allUnicode...)

	} else {
		//return &singleChar{mod, r.next(), cp(r.groups)}
		c := r.next()
		return chars(mod, r.groups, false, span{c, c})
	}
}

func chars(mod *modifier, groups set[int], exclude bool, spans ...span) *char {
	if exclude {
		spans = spanSet(spans).invertUnicode()
	} else {
		spans = spanSet(spans).compact()
	}
	return &char{
		set:      spans,
		modifier: mod,
		groups:   maps.Clone(groups),
		exclude:  exclude,
	}
}

//func cp(groups *list.List) *list.List {
//	cp := list.New()
//	for g := groups.Front(); g != nil; g = g.Next() {
//		cp.PushBack(g.Value)
//	}
//	return cp
//}
