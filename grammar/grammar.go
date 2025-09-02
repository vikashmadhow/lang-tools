// package grammar

package grammar

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/vikashmadhow/lang-tools/lexer"
)

type (
	Grammar struct {
		Id          string
		Lexer       *lexer.Lexer
		Productions []*Production
		ProdByName  map[string]*Production
	}

	Element interface {
		Terminal() bool

		Empty() bool
		MatchEmpty() bool

		First() map[*lexer.TokenType]bool

		fmt.Stringer
	}

	Sentence []Element

	Production struct {
		Name       string
		Alternates []Sentence
		first      map[*lexer.TokenType]bool
	}

	// Rule makes it easier to define grammar rules with self-reference
	// and mutually recursive rules. These rules are resolved to production
	// and token-types on grammar creation.
	//
	// Rule names starting with a lower-case letter are considered non-terminals
	// (productions), while those starting with an upper-case letter are terminals
	// (token types).
	//
	// Example:
	//
	// language := []Rule{
	//   {"e",     [][]string{{"t", "e'"}}},
	//   {"e'",    [][]string{{"PLUS", "t", "e'"}, {}}},
	//   {"t",     [][]string{{"f", "t'"}}},
	//   {"t'",    [][]string{{"TIME", "f", "t'"}, {}}},
	//   {"f",     [][]string{{"ID"}, {"OPEN", "e", "CLOSE"}}},
	//   {"PLUS",  [][]string{{"\\+"}}},
	//   {"TIME",  [][]string{{"\\*"}}},
	//   {"OPEN",  [][]string{{"\\("}}},
	//   {"CLOSE", [][]string{{"\\)"}}},
	//   {"ID",    [][]string{{"[_a-zA-Z][_a-zA-Z0-9]*"}}},
	// }
	Rule struct {
		Name  string
		Match [][]string
	}

	//Sequence struct {
	//	Elements      []Sentence
	//	TreeRetention TreeRetention
	//	first         map[string]bool
	//}
	//
	//Choice struct {
	//	Alternates    []Sentence
	//	TreeRetention TreeRetention
	//	first         map[string]bool
	//}
	//
	//Optional struct {
	//	Sentence      Sentence
	//	TreeRetention TreeRetention
	//}
	//
	//ZeroOrMore struct {
	//	Sentence      Sentence
	//	TreeRetention TreeRetention
	//}
	//
	//OneOrMore struct {
	//	Sentence      Sentence
	//	TreeRetention TreeRetention
	//}
	//
	//Repeat struct {
	//	Min, Max      int
	//	Sentence      Sentence
	//	TreeRetention TreeRetention
	//	first         map[string]bool
	//	follow        map[string]bool
	//}
)

func (s Sentence) Terminal() bool {
	if len(s) == 0 {
		return true
	} else if len(s) == 1 && s[0].Terminal() {
		return true
	}
	return false
}

func (s Sentence) Empty() bool {
	return len(s) == 0 || (len(s) == 1 && s[0].Empty())
}

func (s Sentence) MatchEmpty() bool {
	for _, e := range s {
		if !e.MatchEmpty() {
			return false
		}
	}
	return true
}

func (s Sentence) First() map[*lexer.TokenType]bool {
	first := make(map[*lexer.TokenType]bool)
	for _, e := range s {
		maps.Insert(first, maps.All(e.First()))
		if !e.MatchEmpty() {
			break
		}
	}
	return first
}

func (s Sentence) String() string {
	var str strings.Builder
	for i, e := range s {
		if i > 0 {
			str.WriteString(" ")
		}
		str.WriteString(e.String())
	}
	return str.String()
}

func (p *Production) Terminal() bool {
	return false
}

func (p *Production) Empty() bool {
	for _, alt := range p.Alternates {
		if alt.Empty() {
			return true
		}
	}
	return false
}

func (p *Production) MatchEmpty() bool {
	for _, alt := range p.Alternates {
		if alt.MatchEmpty() {
			return true
		}
	}
	return false
}

func (p *Production) First() map[*lexer.TokenType]bool {
	if p.first == nil {
		p.first = make(map[*lexer.TokenType]bool)
		for _, alt := range p.Alternates {
			maps.Insert(p.first, maps.All(alt.First()))
		}
	}
	return p.first
}

func (p *Production) String() string {
	return p.Name
}

// --- Production  --- //

// s : a b
// x : b s A
// y : s t?
// t : X
// z : y s

// follow(s) = {A, X, V}
// b s A
// s X
// s V

//// --- Choice --- //
//
//func (c *Choice) Terminal() bool {
//	return false
//}
//
//func (c *Choice) First(g *Grammar, cd CycleDetector) (map[string]bool, error) {
//	if c.first != nil {
//		return c.first, nil
//	}
//	err := cd.add(c)
//	if err != nil {
//		return nil, err
//	}
//	c.first = map[string]bool{}
//	for _, a := range c.Alternates {
//		f, e := a.First(g, cd)
//		if e != nil {
//			return nil, e
//		}
//		maps.Insert(c.first, maps.All(f))
//	}
//	cd.remove(c)
//	return c.first, nil
//}
//
//func (c *Choice) Follow(g *Grammar, production string, cd CycleDetector) (map[string]bool, bool, error) {
//	follow := make(map[string]bool)
//	emptyTillEnd := false
//	for _, a := range c.Alternates {
//		f, empty, err := a.Follow(g, production, cd)
//		if err != nil {
//			return nil, false, err
//		}
//		maps.Insert(follow, maps.All(f))
//		emptyTillEnd = emptyTillEnd || empty
//	}
//	return follow, emptyTillEnd, nil
//}
//
//func (c *Choice) MatchEmpty(g *Grammar) bool {
//	for _, a := range c.Alternates {
//		if a.MatchEmpty(g) {
//			return true
//		}
//	}
//	return false
//}
//
//func (c *Choice) Copy() LanguageElement {
//	altCopy := make([]Sentence, len(c.Alternates))
//	for i, a := range c.Alternates {
//		altCopy[i] = a.Copy().(Sentence)
//	}
//	return &Choice{altCopy, c.TreeRetention, c.first}
//}
//
//func (c *Choice) ToString() string {
//	s := ""
//	first := true
//	for _, a := range c.Alternates {
//		if first {
//			first = false
//		} else {
//			s += " | "
//		}
//		s += a.ToString()
//	}
//	return s
//}
//
//// --- SEQUENCE --- //
//
//func (s *Sequence) Terminal() bool {
//	return false
//}
//
//func (s *Sequence) First(g *Grammar, cd CycleDetector) (map[string]bool, error) {
//	if s.first != nil {
//		return s.first, nil
//	}
//	err := cd.add(s)
//	if err != nil {
//		return nil, err
//	}
//	s.first = map[string]bool{}
//	for _, e := range s.Elements {
//		f, er := e.First(g, cd)
//		if er != nil {
//			return nil, er
//		}
//		maps.Insert(s.first, maps.All(f))
//		if !e.MatchEmpty(g) {
//			break
//		}
//	}
//	cd.remove(s)
//	return s.first, nil
//}
//
//func (s *Sequence) MatchEmpty(g *Grammar) bool {
//	for _, e := range s.Elements {
//		if !e.MatchEmpty(g) {
//			return false
//		}
//	}
//	return true
//}
//
//func (s *Sequence) Copy() LanguageElement {
//	elCopy := make([]Sentence, len(s.Elements))
//	for i, e := range s.Elements {
//		elCopy[i] = e.Copy().(Sentence)
//	}
//	return &Sequence{elCopy, s.TreeRetention, s.first}
//}
//
//func (s *Sequence) ToString() string {
//	text := ""
//	first := true
//	for _, e := range s.Elements {
//		if first {
//			first = false
//		} else {
//			text += " "
//		}
//		text += e.ToString()
//	}
//	return text
//}
//
//// --- Optional (?) --- //
//
//func (o *Optional) Terminal() bool {
//	return false
//}
//
//func (o *Optional) First(g *Grammar, cd CycleDetector) (map[string]bool, error) {
//	return o.Sentence.First(g, cd)
//}
//
//func (o *Optional) Follow(g *Grammar, production string, cd CycleDetector) (map[string]bool, bool, error) {
//	f, _, err := o.Sentence.Follow(g, production, cd)
//	if err != nil {
//		return nil, false, err
//	}
//	return f, true, nil
//}
//
//func (o *Optional) MatchEmpty(*Grammar) bool {
//	return true
//}
//
//func (o *Optional) Copy() LanguageElement {
//	return &Optional{o.Sentence.Copy().(Sentence), o.TreeRetention}
//}
//
//func (o *Optional) ToString() string {
//	return "(" + o.Sentence.ToString() + ")?"
//}
//
//// --- Zero or more (*) --- //
//
//func (o *ZeroOrMore) Terminal() bool {
//	return false
//}
//
//func (o *ZeroOrMore) First(g *Grammar, cd CycleDetector) (map[string]bool, error) {
//	return o.Sentence.First(g, cd)
//}
//
//func (o *ZeroOrMore) Follow(g *Grammar, production string, cd CycleDetector) (map[string]bool, bool, error) {
//	return o.Sentence.Follow(g, production, cd)
//}
//
//func (o *ZeroOrMore) MatchEmpty(*Grammar) bool {
//	return true
//}
//
//func (o *ZeroOrMore) Copy() LanguageElement {
//	return &ZeroOrMore{o.Sentence.Copy().(Sentence), o.TreeRetention}
//}
//
//func (o *ZeroOrMore) ToString() string {
//	return "(" + o.Sentence.ToString() + ")*"
//}
//
//// --- One or more (*) --- //
//
//func (o *OneOrMore) Terminal() bool {
//	return false
//}
//
//func (o *OneOrMore) First(g *Grammar, cd CycleDetector) (map[string]bool, error) {
//	return o.Sentence.First(g, cd)
//}
//
//func (o *OneOrMore) Follow(g *Grammar, production string, cd CycleDetector) (map[string]bool, bool, error) {
//	return o.Sentence.Follow(g, production, cd)
//}
//
//func (o *OneOrMore) MatchEmpty(g *Grammar) bool {
//	return o.Sentence.MatchEmpty(g)
//}
//
//func (o *OneOrMore) Copy() LanguageElement {
//	return &OneOrMore{o.Sentence.Copy().(Sentence), o.TreeRetention}
//}
//
//func (o *OneOrMore) ToString() string {
//	return "(" + o.Sentence.ToString() + ")+"
//}
//
//// --- Repeat match {m,n} --- //
//
//func (r *Repeat) Terminal() bool {
//	return false
//}
//
//func (r *Repeat) First(g *Grammar, cd CycleDetector) (map[string]bool, error) {
//	return r.Sentence.First(g, cd)
//}
//
//func (r *Repeat) Follow(g *Grammar, production string, cd CycleDetector) (map[string]bool, bool, error) {
//	return r.Sentence.Follow(g, production, cd)
//}
//
//func (r *Repeat) MatchEmpty(g *Grammar) bool {
//	return r.Min == 0 || r.Sentence.MatchEmpty(g)
//}
//
//func (r *Repeat) Copy() LanguageElement {
//	return &Repeat{r.Min, r.Max, r.Sentence.Copy().(Sentence), r.TreeRetention, r.first, r.follow}
//}
//
//func (r *Repeat) ToString() string {
//	return "(" + r.Sentence.ToString() + "){" + strconv.Itoa(r.Min) + "," + strconv.Itoa(r.Max) + "}"
//}

func NewGrammar(name string, rules []Rule) *Grammar {
	tokenTypes, prods, mods := resolve(rules)
	lex := lexer.NewLexer(slices.Collect(maps.Values(tokenTypes))...)
	if mods != nil {
		lex.Modulator(mods...)
	}
	return &Grammar{
		name,
		lex,
		slices.Collect(maps.Values(prods)),
		prods,
	}
}

func resolve(rules []Rule) (map[string]*lexer.TokenType, map[string]*Production, []lexer.Modulator) {
	var tokens = make(map[string]*lexer.TokenType)
	var prods = make(map[string]*Production)
	var modulators []lexer.Modulator
	for _, r := range rules {
		if startsWithUpper(r.Name) {
			tokens[r.Name] = lexer.NewTokenType(r.Name, r.Match[0][0])
			if len(r.Match) > 1 && r.Match[1][0] == "#Ignore" {
				modulators = append(modulators, lexer.Ignore(tokens[r.Name]))
			}
		} else {
			prods[r.Name] = &Production{Name: r.Name, Alternates: nil}
		}
	}

	for _, r := range rules {
		if !startsWithUpper(r.Name) {
			var alternates []Sentence
			for _, alt := range r.Match {
				var sentence Sentence
				for _, symbol := range alt {
					if startsWithUpper(symbol) {
						sentence = append(sentence, tokens[symbol])
					} else {
						sentence = append(sentence, prods[symbol])
					}
				}
				alternates = append(alternates, sentence)
			}
			prods[r.Name].Alternates = alternates
		}
	}
	return tokens, prods, modulators
}

func startsWithUpper(s string) bool {
	first, _ := utf8.DecodeRuneInString(s)
	return unicode.IsUpper(first)
}
