// author: Vikash Madhow (vikash.madhow@gmail.com)

package regex

// The char interface represents the different type of character and
// character sets used in regular expressions. They become the label
// in the NFA and DFA generated to recognize strings with the regular
// expressions.

import (
	"container/list"
	"embed"
	"math"
	"math/rand"
	"strings"
	"unicode"
	"unicode/utf8"
)

//go:embed lists/*
var lists embed.FS

// -------------Character and character sets parsing-------------//
type (
	char interface {
		match(c rune) bool
		isEmpty() bool

		groups() *list.List // [int]
		setGroups(g *list.List)

		modifier() *modifier

		// spanSet returns the range of characters that can be matched by this char.
		spanSet() spanSet

		random() string

		Pattern
	}

	empty struct{ _ uint8 }

	anyChar struct {
		mod   *modifier
		group list.List
	}

	singleChar struct {
		mod   *modifier
		char  rune
		group list.List
	}

	charRange struct {
		mod   *modifier
		from  rune
		to    rune
		group list.List
	}

	charSet struct {
		mod     *modifier
		exclude bool
		sets    list.List // [char]
		group   list.List // [int]

		span spanSet
	}

	// Matches with a list of strings. This is only used for random generation
	// from the list of strings.
	conversion struct {
		lower       bool
		upper       bool
		title       bool
		trim        bool
		singleSpace bool
	}

	inList struct {
		mod     *modifier
		list    string
		words   []string
		convert conversion
		group   list.List // [int]
	}
)

//------------- The empty character -------------//

func (c *empty) String() string {
	return ""
}

func (c *empty) isEmpty() bool {
	return true
}

func (c *empty) groups() *list.List {
	return nil
}

func (c *empty) setGroups(g *list.List) {}

func (c *empty) nfa() *automata {
	return nil
}

func (c *empty) match(rune) bool {
	return false
}

func (c *empty) spanSet() spanSet {
	return nil
}

func (c *empty) random() string {
	return ""
}

func (c *empty) modifier() *modifier {
	return nil
}

//------------- Any character -------------//

func (c *anyChar) String() string {
	return "."
	//return ".:" + label(c.groups())
}

func (c *anyChar) isEmpty() bool {
	return false
}

func (c *anyChar) groups() *list.List {
	return &c.group
}

func (c *anyChar) setGroups(g *list.List) {
	c.group = *g
}

func (c *anyChar) nfa() *automata {
	return charNfa(c)
}

func (c *anyChar) match(rune) bool {
	return true
}

func (c *anyChar) spanSet() spanSet {
	//if c.mod.unicode {
	return allUnicode
	//} else {
	//    return asciiPrintable
	//}
}

func (c *anyChar) random() string {
	return string(c.spanSet().random())
}

func (c *anyChar) modifier() *modifier {
	return c.mod
}

//------------- A single character match -------------//

func (c *singleChar) String() string {
	return string(c.char)
	//return string(c.char) + ":" + label(c.groups())
}

func (c *singleChar) isEmpty() bool {
	return false
}

func (c *singleChar) groups() *list.List {
	return &c.group
}

func (c *singleChar) setGroups(g *list.List) {
	c.group = *g
}

func (c *singleChar) nfa() *automata {
	return charNfa(c)
}

func (c *singleChar) match(char rune) bool {
	if c.mod.caseInsensitive {
		return unicode.ToLower(char) == unicode.ToLower(c.char)
	} else {
		return c.char == char
	}
}

func (c *singleChar) spanSet() spanSet {
	if c.mod.caseInsensitive {
		l := unicode.ToLower(c.char)
		u := unicode.ToUpper(c.char)
		if l != u {
			return spanSet{
				{l, l},
				{u, u},
			}
		}
	}
	return spanSet{
		{c.char, c.char},
	}
}

func (c *singleChar) random() string {
	return string(c.spanSet().random())
}

func (c *singleChar) modifier() *modifier {
	return c.mod
}

//------------- A character range match -------------//

func (c *charRange) String() string {
	if c.to < math.MaxUint8 {
		return string(c.from) + "-" + string(c.to)
	} else {
		return string(c.from) + "-"
	}
}

func (c *charRange) isEmpty() bool {
	return false
}

func (c *charRange) groups() *list.List {
	return &c.group
}

func (c *charRange) setGroups(g *list.List) {
	c.group = *g
}

func (c *charRange) nfa() *automata {
	return charNfa(c)
}

func (c *charRange) match(char rune) bool {
	if c.mod.caseInsensitive {
		lf := unicode.ToLower(c.from)
		uf := unicode.ToUpper(c.from)

		lt := unicode.ToLower(c.to)
		ut := unicode.ToUpper(c.to)

		if lf != uf || lt != ut {
			return (lf <= char && char <= lt) || (uf <= char && char <= ut)
		}
	}
	return c.from <= char && char <= c.to
}

func (c *charRange) spanSet() spanSet {
	if c.mod.caseInsensitive {
		lf := unicode.ToLower(c.from)
		uf := unicode.ToUpper(c.from)

		lt := unicode.ToLower(c.to)
		ut := unicode.ToUpper(c.to)

		if lf != uf && lt != ut {
			return spanSet{
				{lf, lt},
				{uf, ut},
			}.compact()
		}
	}
	return spanSet{
		{c.from, c.to},
	}
}

func (c *charRange) random() string {
	return string(c.spanSet().random())
}

func (c *charRange) modifier() *modifier {
	return c.mod
}

//------------- A character set combines different characters (and ranges) -------------//

func (c *charSet) String() string {
	ret := "["
	if c.exclude {
		ret += "^"
	}
	first := true
	for cs := c.sets.Front(); cs != nil; cs = cs.Next() {
		if first {
			first = false
		} else {
			ret += "|"
		}
		ret += cs.Value.(char).String()
	}
	ret += "]"
	return ret
}

func (c *charSet) isEmpty() bool {
	return false
}

func (c *charSet) groups() *list.List {
	return &c.group
}

func (c *charSet) setGroups(g *list.List) {
	c.group = *g
}

func (c *charSet) nfa() *automata {
	return charNfa(c)
}

func (c *charSet) match(ch rune) bool {
	return c.spanSet().match(ch)
	//matched := false
	//for cs := c.sets.Front(); cs != nil; cs = cs.Next() {
	//	if cs.Value.(char).match(ch) {
	//		matched = true
	//		break
	//	}
	//}
	//if c.exclude {
	//	return !matched
	//}
	//return matched
}

func (c *charSet) spanSet() spanSet {
	if c.span == nil {
		for cs := c.sets.Front(); cs != nil; cs = cs.Next() {
			c.span = append(c.span, cs.Value.(char).spanSet()...)
		}
		if c.exclude {
			//if c.mod.unicode {
			c.span = c.span.invertUnicode()
			//} else {
			//	c.span = c.span.invertAsciiPrintable()
			//}
		} else {
			c.span = c.span.compact()
		}
	}
	return c.span
}

func (c *charSet) random() string {
	return string(c.spanSet().random())
}

func (c *charSet) modifier() *modifier {
	return c.mod
}

//----------------- In list ----------------//

func (c *inList) String() string {
	return "(:" + c.list + ")"
}

func (c *inList) isEmpty() bool {
	return false
}

func (c *inList) groups() *list.List {
	return &c.group
}

func (c *inList) setGroups(g *list.List) {
	c.group = *g
}

func (c *inList) nfa() *automata {
	return charNfa(c)
}

func (c *inList) match(char rune) bool {
	return false
}

func (c *inList) spanSet() spanSet {
	return nil
}

func (c *inList) random() string {
	if c.words == nil {
		bytes, err := lists.ReadFile("lists/" + c.list)
		if err != nil {
			panic(err)
		}
		content := string(bytes)
		c.words = strings.Split(content, "\n")
		for i, w := range c.words {
			c.words[i] = strings.TrimSpace(w)
		}
	}
	word := c.words[rand.Intn(len(c.words))]
	if c.convert.trim {
		word = strings.TrimSpace(word)
	}
	if c.convert.singleSpace {
		word = singleSpace(word)
	}
	if c.convert.lower {
		word = strings.ToLower(word)
	} else if c.convert.upper {
		word = strings.ToUpper(word)
	} else if c.convert.title {
		r, s := utf8.DecodeRuneInString(word)
		word = string(unicode.ToUpper(r)) + strings.ToLower(word[s:])
	}
	return word
}

func (c *inList) modifier() *modifier {
	return c.mod
}

func singleSpace(s string) string {
	replaced := strings.Builder{}
	var spaceFound bool
	for _, c := range s {
		if unicode.IsSpace(c) {
			if !spaceFound {
				spaceFound = true
				replaced.WriteRune(c)
			}
		} else {
			spaceFound = false
			replaced.WriteRune(c)
		}
	}
	return replaced.String()
}
