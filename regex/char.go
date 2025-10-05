// author: Vikash Madhow (vikash.madhow@gmail.com)

package regex

// The char interface represents the different type of character and
// character sets used in regular expressions. They become the label
// in the NFA and DFA generated to recognize strings with the regular
// expressions.

import (
	"embed"
	"math/rand"
	"strings"
	"unicode"
	"unicode/utf8"
)

//go:embed lists/*
var lists embed.FS

// -------------Character and character sets parsing-------------//
type (
	char struct {
		set     spanSet
		exclude bool
		//groups   *list.List
		groups   set[int]
		modifier *modifier

		list    string
		words   []string
		convert *conversion
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
)

// The empty character
var emptyChar = &char{}

func (c *char) IsEmpty() bool {
	return len(c.set) == 0
}

func (c *char) Copy() *char {
	cp := *c
	return &cp
}

func (c *char) nfa() *Nfa {
	return charNfa(c)
}

func (c *char) match(ch rune) bool {
	return c.set.match(ch)
}

func (c *char) random() string {
	if c.list != "" {
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
	} else {
		return string(c.set.random())
	}
}

func (c *char) String() string {
	var str strings.Builder
	str.WriteString("[")
	s := c.set
	if c.exclude {
		str.WriteString("^")
		s = s.invertUnicode()
	}
	for _, s := range s {
		str.WriteString(s.String())
	}
	str.WriteString("]")
	return str.String()
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
