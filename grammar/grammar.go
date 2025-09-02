// package grammar

package grammar

import (
	"fmt"
	"io"
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

		ToString() string
	}

	Parser interface {
		Parse(input io.Reader, start *Production) (*ParseTree, error)
	}

	LL1Parser struct {
		Grammar       *Grammar
		Start         *Production
		LL1ParseTable map[Element]map[*lexer.TokenType]Sentence

		follow map[Element]map[*lexer.TokenType]bool
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

func NewLL1Parser(g *Grammar, start *Production) *LL1Parser {
	followMap := make(map[Element]map[*lexer.TokenType]bool)
	followMap[start] = map[*lexer.TokenType]bool{lexer.EOFType: true}

	// Algorithm to compute FOLLOW, adapted from Aho and Ullman.
	// repeat
	//     for each production X → Y1 Y2 ··· Yk
	//         for i from k to 1
	//             if i = k then
	//                  FOLLOW[Y i] ← FOLLOW[Y i] ∪ FOLLOW[X]
	//             else
	//                  FOLLOW[Y i] ← FOLLOW[Y i] ∪ FIRST[Y i+1]
	//                  if Y i+1 is empty then FOLLOW[Y i] ← FOLLOW[Y i] ∪ FOLLOW[Y i+1]
	// until FOLLOW did not change in this iteration.
	for changed := true; changed; {
		changed = false
		for _, p := range g.Productions {
			for _, alt := range p.Alternates {
				for i := len(alt) - 1; i >= 0; i-- {
					if i == len(alt)-1 {
						changed = changed || addFollow(followMap, alt[i], p, false)
					} else {
						changed = changed || addFollow(followMap, alt[i], alt[i+1], true)
						if alt[i+1].MatchEmpty() {
							changed = changed || addFollow(followMap, alt[i], alt[i+1], false)
						}
					}
				}
			}
		}
	}

	table := make(map[Element]map[*lexer.TokenType]Sentence)
	for _, p := range g.Productions {
		matchEmpty := false
		table[p] = make(map[*lexer.TokenType]Sentence)
		for _, alt := range p.Alternates {
			if alt.MatchEmpty() {
				matchEmpty = true
			} else {
				for t := range alt.First() {
					table[p][t] = alt
				}
			}
		}
		if matchEmpty {
			for t := range followMap[p] {
				table[p][t] = Sentence{}
			}
		}
	}
	return &LL1Parser{g, start, table, followMap}
}

func addFollow(
	followMap map[Element]map[*lexer.TokenType]bool,
	addTo Element,
	followedBy Element,
	addFirst bool) bool {

	if followMap[addTo] == nil {
		followMap[addTo] = make(map[*lexer.TokenType]bool)
	}

	var elements map[*lexer.TokenType]bool
	if addFirst {
		elements = followedBy.First()
	} else {
		if followMap[followedBy] == nil {
			followMap[followedBy] = make(map[*lexer.TokenType]bool)
		}
		elements = followMap[followedBy]
	}
	changed := false
	for k := range elements {
		if !followMap[addTo][k] {
			followMap[addTo][k] = true
			changed = true
		}
	}
	return changed
}

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

func (s Sentence) ToString() string {
	var str strings.Builder
	for i, e := range s {
		if i > 0 {
			str.WriteString(" ")
		}
		str.WriteString(e.ToString())
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

func (p *Production) ToString() string {
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
//func (c *Choice) Recognise(g *Grammar, production LanguageElement, tokens *TokenSeq, cd CycleDetector) (*ParseTree, error) {
//	token, err, _ := tokens.Peek()
//	if err != nil {
//		return nil, err
//	}
//
//	var alternate Sentence
//	for _, a := range c.Alternates {
//		first, err := a.First(g, cd)
//		if err != nil {
//			return nil, err
//		}
//		if _, ok := first[token.Type]; ok {
//			alternate = a
//			break
//		}
//	}
//	if alternate == nil {
//		return nil, fmt.Errorf("no alternates found for choice %q on token %q", c.ToString(), token.Type)
//	}
//	return alternate.Recognise(g, production, tokens, cd)
//}
//
//func (c *Choice) Retention() TreeRetention {
//	return c.TreeRetention
//}
//
//func (c *Choice) SetRetention(tr TreeRetention) {
//	c.TreeRetention = tr
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
//func (s *Sequence) Follow(g *Grammar, production string, cd CycleDetector) (map[string]bool, bool, error) {
//	found := -1
//search:
//	for i, e := range s.Elements {
//		switch p := e.(type) {
//		case *ProductionRef:
//			if p.Ref == production {
//				found = i
//				break search
//			}
//		}
//	}
//
//	follow := make(map[string]bool)
//	emptyTillEnd := true
//	if found != -1 {
//		for _, e := range s.Elements[found+1:] {
//			first, err := e.First(g, cd)
//			if err != nil {
//				return nil, false, err
//			}
//			maps.Insert(follow, maps.All(first))
//			if !e.MatchEmpty(g) {
//				emptyTillEnd = false
//				break
//			}
//		}
//	}
//	return follow, emptyTillEnd, nil
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
//func (s *Sequence) Recognise(g *Grammar, production LanguageElement, tokens *TokenSeq, cd CycleDetector) (*ParseTree, error) {
//	tree := &ParseTree{production, nil}
//	for _, e := range s.Elements {
//		token, err, _ := tokens.Peek()
//		if err != nil {
//			return nil, err
//		}
//		first, err := e.First(g, cd)
//		if err != nil {
//			return nil, err
//		}
//		if _, ok := first[token.Type]; ok {
//			child, err := e.Recognise(g, production, tokens, cd)
//			if err != nil {
//				return nil, err
//			}
//
//			//if child.Node.Retention() > Promoted {
//			//	//    t  		         y
//			//	//  x   y     -->      x a b
//			//	//     a  b
//			//	newTree := *child
//			//	newTree.Node = newTree.Node.Copy()
//			//	newTree.Node.SetRetention(newTree.Node.Retention() - 1)
//			//	newTree.Children = slices.Concat(tree.Children, newTree.Children)
//			//	tree = &newTree
//			//} else if child.Node.Retention() >= Retain {
//			//if child.Node.Retention() != Drop {
//			tree.Children = append(tree.Children, child)
//			//}
//		} else if !e.MatchEmpty(g) {
//			return nil, fmt.Errorf("token %q cannot start %q", token.Type, e.ToString())
//		}
//	}
//	//if len(tree.Children) == 1 {
//	//    return tree.Children[0], nil
//	//} else {
//	//    return tree, nil
//	//}
//	return tree, nil
//}
//
//func (s *Sequence) Retention() TreeRetention {
//	return s.TreeRetention
//}
//
//func (s *Sequence) SetRetention(tr TreeRetention) {
//	s.TreeRetention = tr
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
//func (o *Optional) Recognise(g *Grammar, production LanguageElement, tokens *TokenSeq, cd CycleDetector) (*ParseTree, error) {
//	token, err, _ := tokens.Peek()
//	if err != nil {
//		return nil, err
//	}
//	first, err := o.Sentence.First(g, cd)
//	if err != nil {
//		return nil, err
//	}
//	if _, ok := first[token.Type]; ok {
//		return o.Sentence.Recognise(g, production, tokens, cd)
//	} else if !o.Sentence.MatchEmpty(g) {
//		return nil, fmt.Errorf("token %q cannot start %q", token.Type, o.Sentence.ToString())
//	}
//	return nil, nil
//}
//
//func (o *Optional) Retention() TreeRetention {
//	return o.TreeRetention
//}
//
//func (o *Optional) SetRetention(tr TreeRetention) {
//	o.TreeRetention = tr
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
//func (o *ZeroOrMore) Recognise(g *Grammar, production LanguageElement, tokens *TokenSeq, cd CycleDetector) (*ParseTree, error) {
//	first, err := o.Sentence.First(g, cd)
//	if err != nil {
//		return nil, err
//	}
//	tree := ParseTree{production, nil}
//	matchedOnce := false
//	for {
//		token, err, _ := tokens.Peek()
//		if err != nil {
//			return nil, err
//		}
//		if _, ok := first[token.Type]; ok {
//			matchedOnce = true
//			child, err := o.Sentence.Recognise(g, production, tokens, cd)
//			if err != nil {
//				return nil, err
//			}
//			tree.Children = append(tree.Children, child)
//
//		} else if !matchedOnce && !o.Sentence.MatchEmpty(g) {
//			return nil, fmt.Errorf("token %q cannot start %q", token.Type, o.Sentence.ToString())
//
//		} else {
//			break
//		}
//	}
//	//if len(tree.Children) == 1 {
//	//	return tree.Children[0], nil
//	//} else {
//	//	return &tree, nil
//	//}
//	return &tree, nil
//}
//
//func (o *ZeroOrMore) Retention() TreeRetention {
//	return o.TreeRetention
//}
//
//func (o *ZeroOrMore) SetRetention(tr TreeRetention) {
//	o.TreeRetention = tr
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
//func (o *OneOrMore) Recognise(g *Grammar, production LanguageElement, tokens *TokenSeq, cd CycleDetector) (*ParseTree, error) {
//	first, err := o.Sentence.First(g, cd)
//	if err != nil {
//		return nil, err
//	}
//	tree := ParseTree{production, nil}
//	matchedOnce := false
//	for {
//		token, err, _ := tokens.Peek()
//		if err != nil {
//			return nil, err
//		}
//		if _, ok := first[token.Type]; ok {
//			matchedOnce = true
//			child, err := o.Sentence.Recognise(g, production, tokens, cd)
//			if err != nil {
//				return nil, err
//			}
//			tree.Children = append(tree.Children, child)
//
//		} else if !matchedOnce && !o.Sentence.MatchEmpty(g) {
//			return nil, fmt.Errorf("token %q cannot start %q", token.Type, o.Sentence.ToString())
//
//		} else {
//			break
//		}
//	}
//	//if len(tree.Children) == 1 {
//	//	return tree.Children[0], nil
//	//} else {
//	//	return &tree, nil
//	//}
//	return &tree, nil
//}
//
//func (o *OneOrMore) Retention() TreeRetention {
//	return o.TreeRetention
//}
//
//func (o *OneOrMore) SetRetention(tr TreeRetention) {
//	o.TreeRetention = tr
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
//func (r *Repeat) Recognise(g *Grammar, production LanguageElement, tokens *TokenSeq, cd CycleDetector) (*ParseTree, error) {
//	first, err := r.Sentence.First(g, cd)
//	if err != nil {
//		return nil, err
//	}
//	tree := ParseTree{production, nil}
//
//	for matched := 0; matched < r.Max; matched++ {
//		token, err, _ := tokens.Peek()
//		if err != nil {
//			return nil, err
//		}
//		if _, ok := first[token.Type]; ok {
//			child, err := r.Sentence.Recognise(g, production, tokens, cd)
//			if err != nil {
//				return nil, err
//			}
//			tree.Children = append(tree.Children, child)
//
//		} else if matched < r.Min && !r.Sentence.MatchEmpty(g) {
//			return nil, fmt.Errorf("token %q cannot start %q", token.Type, r.Sentence.ToString())
//
//		} else if matched >= r.Min {
//			break
//		}
//	}
//	//if len(tree.Children) == 1 {
//	//	return tree.Children[0], nil
//	//} else {
//	//	return &tree, nil
//	//}
//	return &tree, nil
//}
//
//func (r *Repeat) Retention() TreeRetention {
//	return r.TreeRetention
//}
//
//func (r *Repeat) SetRetention(tr TreeRetention) {
//	r.TreeRetention = tr
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
	lexer := lexer.NewLexer(slices.Collect(maps.Values(tokenTypes))...)
	if mods != nil {
		lexer.Modulator(mods...)
	}
	return &Grammar{
		name,
		lexer,
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

func (p *LL1Parser) Parse(input io.Reader) (*ParseTree, error) {
	start := NewSyntaxTree(p.Start)
	stack := []*ParseTree{start, NewSyntaxTree(lexer.EOFType)}
	for token := range p.Grammar.Lexer.LexSeq(input) {
		for token.Type != stack[0].Node {
			top := stack[0]
			stack = stack[1:]

			sentence := p.LL1ParseTable[top.Node][token.Type]
			if sentence == nil {
				return nil, fmt.Errorf("no production for %q -> %q", top.Node, token.Type.Id)
			}
			top.Children = parseTree(sentence)
			stack = slices.Insert(stack, 0, top.Children...)
		}
		stack[0].Children = append(stack[0].Children, NewSyntaxTree(token))
		stack = stack[1:]
	}
	return start, nil
}

func (p *LL1Parser) ParseText(input string) (*ParseTree, error) {
	return p.Parse(strings.NewReader(input))
}

func parseTree(sentence Sentence) []*ParseTree {
	var trees []*ParseTree
	for _, s := range sentence {
		trees = append(trees, NewSyntaxTree(s))
	}
	return trees
}
