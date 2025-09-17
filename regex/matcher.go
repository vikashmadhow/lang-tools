// author: Vikash Madhow (vikash.madhow@gmail.com)

package regex

import (
	"slices"
	"strings"
)

type MatchType int

const (
	// NoMatch Last match was unsuccessful, the matcher must be reset for valid subsequent matching
	NoMatch MatchType = iota

	// PartialMatch Text matched a prefix of the regular expression, but has not reached a final State
	// to be considered a full match.
	PartialMatch

	// FullMatch A full match was achieved; subsequent supplied characters can still result in a full
	// match if the longer string is part of the regular language of the regular expression.
	FullMatch

	// Start Matching has not started yet. The matcher is set to this State on creation or reset.
	Start
)

type Matcher struct {
	LastMatch    MatchType
	FullMatch    strings.Builder
	PartialMatch strings.Builder
	Groups       map[int]string
	Compiled     *CompiledRegex
	State        state
}

func (m *Matcher) Reset() {
	m.LastMatch = Start
	m.FullMatch.Reset()
	m.PartialMatch.Reset()
	m.Groups = make(map[int]string)
	m.State = m.Compiled.Dfa.start
}

func (m *Matcher) Match(input string) bool {
	for _, c := range input {
		if m.MatchNext(c) == NoMatch {
			return false
		}
	}
	return slices.Index(m.Compiled.Dfa.final, m.State) != -1
}

func (m *Matcher) FindNext(input string) bool {
	for _, c := range input {
		if m.MatchNext(c) == NoMatch {
			return false
		}
	}
	return slices.Index(m.Compiled.Dfa.final, m.State) != -1
}

func (m *Matcher) MatchNext(r rune) MatchType {
	if m.LastMatch != NoMatch {
		trans := m.Compiled.Dfa.Trans[m.State]
		for c, t := range trans {
			if c.match(r) {
				m.State = t
				if slices.Index(m.Compiled.Dfa.final, t) == -1 {
					if m.LastMatch == FullMatch {
						m.PartialMatch.Reset()
						m.PartialMatch.WriteString(m.FullMatch.String())
					}
					m.PartialMatch.WriteRune(r)
					m.LastMatch = PartialMatch
				} else {
					if m.LastMatch == PartialMatch {
						m.FullMatch.Reset()
						m.FullMatch.WriteString(m.PartialMatch.String())
					}
					m.FullMatch.WriteRune(r)
					m.LastMatch = FullMatch
				}
				groupSet := set[int]{}
				groups := c.groups()
				for g := groups.Front(); g != nil; g = g.Next() {
					group := g.Value.(int)
					if group != 0 {
						groupSet[group] = true
					}
					_, ok := m.Groups[g.Value.(int)]
					if !ok {
						m.Groups[group] = ""
					}
					m.Groups[group] += string(r)
				}
				return m.LastMatch
			}
		}
		m.LastMatch = NoMatch
	}
	return m.LastMatch
}

// TestMatchNext matches the next character but does not update the matcher's state.
// It is akin to peeking ahead in the input string.
//func (m *Matcher) TestMatchNext(r rune) MatchType {
//	trans := m.Compiled.Dfa.Trans[m.State]
//	for c, t := range trans {
//		if c.match(r) {
//			if slices.Index(m.Compiled.Dfa.final, t) == -1 {
//				return PartialMatch
//			} else {
//				return FullMatch
//			}
//		}
//	}
//	return NoMatch
//}
