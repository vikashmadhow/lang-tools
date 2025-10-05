// author: Vikash Madhow (vikash.madhow@gmail.com)

// Package regex implements a regular expression library that can be
// supplied with a sequence of characters and will return if the characters at
// that point is a prefix matching the regular expression.
package regex

import (
	"maps"
	"math"
	"math/rand"
	"slices"
	"strconv"
	"strings"
)

type (
	// Pattern is the base visible interface of regular expressions
	Pattern interface {
		String() string
		nfa() *Nfa
	}

	Regex struct {
		Pattern Pattern
		Dfa     *Dfa
	}

	// choice represents the regex | regex rule
	choice struct {
		left  Pattern
		right Pattern
	}

	// sequence represents a sequence of regular expressions (a, b, ...)
	sequence struct {
		sequence []Pattern
	}

	// zeroOrOne is for an optional regular expression (re?)
	zeroOrOne struct {
		opt Pattern
	}

	// zeroOrMore is for the Kleene closure (re*)
	zeroOrMore struct {
		re Pattern
	}

	// oneOrMore is for positive closure (re+)
	oneOrMore struct {
		re Pattern
	}

	// oneOrMore is for positive closure (re+)
	repeat struct {
		re       Pattern
		min, max uint8
	}

	// captureGrp is for grouping regular expressions inside brackets, i.e., (re)
	captureGroup struct {
		re Pattern
	}
)

func Escape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	s = strings.ReplaceAll(s, "{", "\\{")
	s = strings.ReplaceAll(s, "}", "\\}")
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "+", "\\+")
	s = strings.ReplaceAll(s, "*", "\\*")
	s = strings.ReplaceAll(s, "?", "\\?")
	return s
}

// NewRegex creates a new regular expression from the input
func NewRegex(input string) *Regex {
	group := 0
	//groups := list.New()
	//groups.PushBack(0)
	groups := set[int]{0: true}
	parser := parser{[]rune(input), 0, &group, groups}
	r := parser.regex(&modifier{caseInsensitive: false, unicode: false})
	n := r.nfa()
	//d := n.Dfa().Minimize()
	d := n.Dfa()
	return &Regex{r, d}
}

func (r *Regex) Matcher() *Matcher {
	return &Matcher{LastMatch: Start, Groups: map[int]*strings.Builder{}, Compiled: r, State: r.Dfa.start}
}

func (r *Regex) Match(input string) bool {
	return r.Matcher().Match(input)
}

func (r *Regex) MatchEmpty() bool {
	return r.Dfa.finalMap[r.Dfa.start]
}

func (r *Regex) Generate() string {
	var s strings.Builder
	state := r.Dfa.start
	trans := r.Dfa.Trans[state]
	for len(trans) > 0 {
		nextStates := len(trans)
		final := r.Dfa.finalMap[state]
		if final {
			nextStates += 1
		}
		n := rand.Intn(nextStates)
		if final && n == nextStates-1 {
			break
		} else {
			t := slices.Collect(maps.Keys(trans))
			c := t[n]
			target := trans[c]

			s.WriteString(target.char.random())
			//s.WriteRune(c.spanSet().random())
			state = target.state
		}
		trans = r.Dfa.Trans[state]
	}
	return s.String()
}

func (r *Regex) String() string {
	return r.Pattern.String()
}

//-----------------Regex interface methods------------//

func (c *choice) String() string {
	return c.left.String() + "|" + c.right.String()
}

// Dfa constructs a finite automaton for the choice (union) of two regular expressions.
//
//	    left
//	    ∧  \
//	   /    v
//	start   final
//	   \    ∧
//	    v  /
//	    right
func (c *choice) nfa() *Nfa {
	final := &stateObj{}
	a := Nfa{
		Trans:    make(NfaTrans),
		start:    &stateObj{},
		final:    []state{final},
		finalMap: map[state]bool{final: true},
	}

	left := c.left.nfa()
	right := c.right.nfa()

	a.merge(left)
	a.merge(right)

	a.addTransitions(a.start, &targetState{"", emptyChar, left.start}, &targetState{"", emptyChar, right.start})
	a.addTransitions(left.final[0], &targetState{"", emptyChar, a.final[0]})
	a.addTransitions(right.final[0], &targetState{"", emptyChar, a.final[0]})

	return &a
}

func (s *sequence) String() string {
	ret := ""
	for _, re := range s.sequence {
		ret += re.String()
	}
	return ret
}

// nfa constructs a finite-state automaton for the sequence of regular expressions.
// It merges the individual Dfa of each regular expression in the sequence, connecting
// the final state of one to the start state of the next. It returns a pointer to the resulting Dfa.
//
//	Start --> re1 in sequence --> re2 --> .... --> final
func (s *sequence) nfa() *Nfa {
	//final := &stateObj{}
	a := Nfa{
		Trans: make(NfaTrans),
		start: &stateObj{},
		//final: []state{final},
		//finalMap: map[state]bool{final: true},
	}

	first := true
	for _, re := range s.sequence {
		reAutomata := re.nfa()
		a.merge(reAutomata)
		if first {
			a.start = reAutomata.start
			first = false
		} else {
			//a.addTransitions(a.final[0], map[char]state{&empty{}: reAutomata.start})
			a.addTransitions(a.final[0], &targetState{"", emptyChar, reAutomata.start})
		}
		a.final = reAutomata.final
		a.finalMap = reAutomata.finalMap
	}
	if first {
		a.final = []state{a.start}
		a.finalMap = map[state]bool{a.start: true}
	}
	return &a
}

func (r *zeroOrOne) String() string {
	return r.opt.String() + "?"
	//return "?(" + r.opt.Pattern() + ")"
}

// Dfa constructs and returns an NFA for an optional subpattern.
//
//	    _______________
//	   /               \
//	  /                 v
//	start --> ... --> final
func (r *zeroOrOne) nfa() *Nfa {
	opt := r.opt.nfa()
	//opt.addTransitions(opt.start, map[char]state{&empty{}: opt.final[0]})
	opt.addTransitions(opt.start, &targetState{"", emptyChar, opt.final[0]})
	return opt
}

func (r *zeroOrMore) String() string {
	return r.re.String() + "*"
	//return "*(" + r.re.Pattern() + ")"
}

// Dfa generates a finite automaton for a zero-or-more repetition (Kleene closure) of the Pattern.
//
//	    ______________
//	   ^              \
//	  /                v
//	start --> ... --> final
//	  ^                /
//	   \              v
//	    --------------
func (r *zeroOrMore) nfa() *Nfa {
	re := r.re.nfa()
	//re.addTransitions(re.start, map[char]state{&empty{}: re.final[0]})
	//re.addTransitions(re.final[0], map[char]state{&empty{}: re.start})
	re.addTransitions(re.start, &targetState{"", emptyChar, re.final[0]})
	re.addTransitions(re.final[0], &targetState{"", emptyChar, re.start})
	return re
}

func (r *oneOrMore) String() string {
	return r.re.String() + "+"
	//return "+(" + r.re.Pattern() + ")"
}

// Dfa generates a finite automaton for a one-or-more repetition of the Pattern.
//
//	start --> ... --> final
//	 ^                  /
//	  \                v
//	    ---------------
func (r *oneOrMore) nfa() *Nfa {
	re := r.re.nfa()
	//re.addTransitions(re.final[0], map[char]state{&empty{}: re.start})
	re.addTransitions(re.final[0], &targetState{"", emptyChar, re.start})
	return re
}

func (r *repeat) String() string {
	s := r.re.String() + "{"
	if r.min == r.max {
		s += strconv.Itoa(int(r.min))
	} else {
		if r.min != 0 {
			s += strconv.Itoa(int(r.min))
		}
		s += ","
		if r.max != math.MaxUint8 {
			s += strconv.Itoa(int(r.max))
		}
	}
	return s + "}"
}

// Dfa generates a finite automaton for a range (m,n) repetition of the Pattern.
//
//	                              ___________________
//							     /   _______________ \
//		    			  	    /   /           ___ \ \
//	         +-m times--+      /   /           /   \ \ \
//	         |          |     ^   ^           ^     v v v
//	start -> r -> ...-> r -> r -> r -> ... -> r ->  final
//	                         |                |
//	                         +---n-m times----+
func (r *repeat) nfa() *Nfa {
	//final := &stateObj{}
	a := &Nfa{
		Trans: make(NfaTrans),
		start: &stateObj{},
		//final:    []state{final},
		//finalMap: map[state]bool{final: true},
	}
	first := true
	if r.min > 0 {
		s := &sequence{slices.Repeat([]Pattern{r.re}, int(r.min))}
		a = s.nfa()
		first = false
	}
	if r.max > r.min {
		if r.max == 255 {
			re := r.re.nfa()
			a.merge(re)
			//a.addTransitions(re.start, map[char]state{&empty{}: re.final[0]})
			//a.addTransitions(re.final[0], map[char]state{&empty{}: re.start})
			a.addTransitions(re.start, &targetState{"", emptyChar, re.final[0]})
			a.addTransitions(re.final[0], &targetState{"()", emptyChar, re.start})
			if first {
				a.start = re.start
				first = false
			} else {
				//a.addTransitions(a.final[0], map[char]state{&empty{}: re.start})
				a.addTransitions(a.final[0], &targetState{"", emptyChar, re.start})
			}
			a.final = re.final
			a.finalMap = re.finalMap
		} else {
			for i := r.min; i < r.max; i++ {
				re := r.re.nfa()
				a.merge(re)
				//a.addTransitions(re.start, map[char]state{&empty{}: re.final[0]})
				a.addTransitions(re.start, &targetState{"", emptyChar, re.final[0]})
				if first {
					a.start = re.start
					first = false
				} else {
					//a.addTransitions(a.final[0], map[char]state{&empty{}: re.start})
					a.addTransitions(a.final[0], &targetState{"", emptyChar, re.start})
				}
				a.final = re.final
				a.finalMap = re.finalMap
			}
		}
	}
	if first {
		a.final = []state{a.start}
		a.finalMap = map[state]bool{a.start: true}
	}
	return a
}

func (r *captureGroup) String() string {
	return "(" + r.re.String() + ")"
}

func (r *captureGroup) nfa() *Nfa {
	return r.re.nfa()
}
