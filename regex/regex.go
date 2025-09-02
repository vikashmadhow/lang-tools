// author: Vikash Madhow (vikash.madhow@gmail.com)

// Package regex implements a regular expression library that can be
// supplied with a sequence of characters and will return if the characters at
// that point is a prefix matching the regular expression.
package regex

import (
    "container/list"
    "maps"
    "math"
    "math/rand"
    "slices"
    "strconv"
    "strings"
)

type (
    // Regex is the base visible interface of regular expressions
    Regex interface {
        Pattern() string
        nfa() *automata
    }

    CompiledRegex struct {
        Regex Regex
        Nfa   *automata
        Dfa   *automata
    }

    // choice represents the regex | regex rule
    choice struct {
        left  Regex
        right Regex
    }

    // sequence represents a sequence of regular expressions (a, b, ...)
    sequence struct {
        sequence []Regex
    }

    // zeroOrOne is for an optional regular expression (re?)
    zeroOrOne struct {
        opt Regex
    }

    // zeroOrMore is for the Kleene closure (re*)
    zeroOrMore struct {
        re Regex
    }

    // oneOrMore is for positive closure (re+)
    oneOrMore struct {
        re Regex
    }

    // oneOrMore is for positive closure (re+)
    repeat struct {
        re       Regex
        min, max uint8
    }

    // captureGrp is for grouping regular expressions inside brackets, i.e., (re)
    captureGroup struct {
        re Regex
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
func NewRegex(input string) *CompiledRegex {
    group := 0
    groups := list.New()
    groups.PushBack(0)
    parser := parser{[]rune(input), 0, &group, groups}
    r := parser.regex(&modifier{caseInsensitive: false, unicode: false})
    n := r.nfa()
    d := n.dfa()
    return &CompiledRegex{r, n, d}
}

func (r *CompiledRegex) Matcher() *Matcher {
    return &Matcher{Start, "", map[int]string{}, r, r.Dfa.start}
}

func (r *CompiledRegex) Match(input string) bool {
    m := r.Matcher()
    for _, c := range input {
        if m.MatchNext(c) == NoMatch {
            return false
        }
    }
    return slices.Index(r.Dfa.final, m.State) != -1
}

func (r *CompiledRegex) MatchEmpty() bool {
    return slices.Index(r.Dfa.final, r.Dfa.start) != -1
}

func (r *CompiledRegex) Generate() string {
    var s strings.Builder
    state := r.Dfa.start
    trans := r.Dfa.Trans[state]
    for len(trans) > 0 {
        nextStates := len(trans)
        final := slices.Index(r.Dfa.final, state) != -1
        if final {
            nextStates += 1
        }
        n := rand.Intn(nextStates)
        if final && n == nextStates-1 {
            break
        } else {
            t := slices.Collect(maps.Keys(trans))
            c := t[n]
            s.WriteString(c.random())
            //s.WriteRune(c.spanSet().random())
            state = trans[c]
        }
        trans = r.Dfa.Trans[state]
    }
    return s.String()
}

//-----------------Regex interface methods------------//

func (c *choice) Pattern() string {
    return c.left.Pattern() + "|" + c.right.Pattern()
    //return "Or(" + c.left.Pattern() + ", " + c.right.Pattern() + ")"
}

// automata constructs a finite automaton for the choice (union) of two regular expressions.
//
//	    left
//	    ∧  \
//	   /    v
//	start   final
//	   \    ∧
//	    v  /
//	    right
func (c *choice) nfa() *automata {
    a := automata{
        Trans: make(transitions),
        start: &stateObj{},
        final: []state{&stateObj{}},
    }

    left := c.left.nfa()
    right := c.right.nfa()

    a.merge(left)
    a.merge(right)

    a.addTransitions(a.start, map[char]state{&empty{}: left.start, &empty{}: right.start})
    a.addTransitions(left.final[0], map[char]state{&empty{}: a.final[0]})
    a.addTransitions(right.final[0], map[char]state{&empty{}: a.final[0]})

    return &a
}

func (s *sequence) Pattern() string {
    ret := ""
    //ret := "Seq("
    //first := true
    for _, re := range s.sequence {
        //if first {
        //	first = false
        //} else {
        //ret += ", "
        //	ret += ""
        //}
        ret += re.Pattern()
    }
    //ret += ")"
    return ret
}

// automata constructs a finite-State automaton for the sequence of regular expressions.
// It merges the individual automata of each regular expression in the sequence, connecting
// the final state of one to the start state of the next. It returns a pointer to the resulting automata.
//
//	start --> re1 in sequence --> re2 --> .... --> final
func (s *sequence) nfa() *automata {
    a := automata{
        Trans: make(transitions),
        start: &stateObj{},
        final: []state{&stateObj{}},
    }

    first := true
    for _, re := range s.sequence {
        reAutomata := re.nfa()
        a.merge(reAutomata)
        if first {
            a.start = reAutomata.start
            first = false
        } else {
            a.addTransitions(a.final[0], map[char]state{&empty{}: reAutomata.start})
        }
        a.final = reAutomata.final
    }
    if first {
        a.final[0] = a.start
    }
    return &a
}

func (r *zeroOrOne) Pattern() string {
    return r.opt.Pattern() + "?"
    //return "?(" + r.opt.Pattern() + ")"
}

// automata constructs and returns an NFA for an optional subpattern.
//
//	    _______________
//	   /               \
//	  /                 v
//	start --> ... --> final
func (r *zeroOrOne) nfa() *automata {
    opt := r.opt.nfa()
    opt.addTransitions(opt.start, map[char]state{&empty{}: opt.final[0]})
    return opt
}

func (r *zeroOrMore) Pattern() string {
    return r.re.Pattern() + "*"
    //return "*(" + r.re.Pattern() + ")"
}

// automata generates a finite automaton for a zero-or-more repetition (Kleene closure) of the pattern.
//
//	    ______________
//	   ^              \
//	  /                v
//	start --> ... --> final
//	  ^                /
//	   \              v
//	    --------------
func (r *zeroOrMore) nfa() *automata {
    re := r.re.nfa()
    re.addTransitions(re.start, map[char]state{&empty{}: re.final[0]})
    re.addTransitions(re.final[0], map[char]state{&empty{}: re.start})
    return re
}

func (r *oneOrMore) Pattern() string {
    return r.re.Pattern() + "+"
    //return "+(" + r.re.Pattern() + ")"
}

// automata generates a finite automaton for a one-or-more repetition of the pattern.
//
//	start --> ... --> final
//	 ^                  /
//	  \                v
//	    ---------------
func (r *oneOrMore) nfa() *automata {
    re := r.re.nfa()
    re.addTransitions(re.final[0], map[char]state{&empty{}: re.start})
    return re
}

func (r *repeat) Pattern() string {
    s := r.re.Pattern() + "{"
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
    //return "*(" + r.re.Pattern() + ")"
}

// automata generates a finite automaton for a range (m,n) repetition of the pattern.
//
//	                              ___________________
//							     /   _______________ \
//		    			  	    /   /           ___ \ \
//	         +-m times--+      /   /           /   \ \ \
//	         |          |     ^   ^           ^     v v v
//	start -> r -> ...-> r -> r -> r -> ... -> r ->  final
//	                         |                |
//	                         +---n-m times----+
func (r *repeat) nfa() *automata {
    a := &automata{
        Trans: make(transitions),
        start: &stateObj{},
        final: []state{&stateObj{}},
    }
    first := true
    if r.min > 0 {
        s := &sequence{slices.Repeat([]Regex{r.re}, int(r.min))}
        a = s.nfa()
        first = false
    }
    if r.max > r.min {
        if r.max == 255 {
            re := r.re.nfa()
            a.merge(re)
            a.addTransitions(re.start, map[char]state{&empty{}: re.final[0]})
            a.addTransitions(re.final[0], map[char]state{&empty{}: re.start})
            if first {
                a.start = re.start
                first = false
            } else {
                a.addTransitions(a.final[0], map[char]state{&empty{}: re.start})
            }
            a.final = re.final
        } else {
            for i := r.min; i < r.max; i++ {
                re := r.re.nfa()
                a.merge(re)
                a.addTransitions(re.start, map[char]state{&empty{}: re.final[0]})
                if first {
                    a.start = re.start
                    first = false
                } else {
                    a.addTransitions(a.final[0], map[char]state{&empty{}: re.start})
                }
                a.final = re.final
            }
        }
    }
    if first {
        a.final[0] = a.start
    }
    return a
}

func (r *captureGroup) Pattern() string {
    return "(" + r.re.Pattern() + ")"
    //return "Grp(" + r.re.Pattern() + ")"
}

func (r *captureGroup) nfa() *automata {
    return r.re.nfa()
}
