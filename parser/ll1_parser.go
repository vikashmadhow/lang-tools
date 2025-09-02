package parser

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/vikashmadhow/lang-tools/grammar"
	"github.com/vikashmadhow/lang-tools/lexer"
)

type (
	LL1Parser struct {
		Grammar       *grammar.Grammar
		Start         *grammar.Production
		LL1ParseTable map[grammar.Element]map[*lexer.TokenType]grammar.Sentence

		follow map[grammar.Element]map[*lexer.TokenType]bool
	}
)

func NewLL1Parser(g *grammar.Grammar, start *grammar.Production) *LL1Parser {
	followMap := make(map[grammar.Element]map[*lexer.TokenType]bool)
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

	table := make(map[grammar.Element]map[*lexer.TokenType]grammar.Sentence)
	for _, p := range g.Productions {
		matchEmpty := false
		table[p] = make(map[*lexer.TokenType]grammar.Sentence)
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
				table[p][t] = grammar.Sentence{}
			}
		}
	}
	return &LL1Parser{
		Grammar:       g,
		Start:         start,
		LL1ParseTable: table,
		follow:        followMap,
	}
}

func addFollow(
	followMap map[grammar.Element]map[*lexer.TokenType]bool,
	addTo grammar.Element,
	followedBy grammar.Element,
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

func (p *LL1Parser) Parse(input io.Reader) (*Tree[grammar.Element], error) {
	start := NewParseTree(p.Start)
	stack := []*Tree[grammar.Element]{start, NewParseTree(lexer.EOFType)}
	for token := range p.Grammar.Lexer.LexSeq(input) {
		for token.Type != stack[0].Node {
			top := stack[0]
			stack = stack[1:]

			sentence := p.LL1ParseTable[top.Node.(grammar.Element)][token.Type]
			if sentence == nil {
				return nil, fmt.Errorf("no production for %q -> %q", top.Node, token.Type.Id)
			}
			top.Children = parseTree(sentence)
			stack = slices.Insert(stack, 0, top.Children...)
		}
		stack[0].Children = append(stack[0].Children, NewParseTree(token))
		stack = stack[1:]
	}
	return start, nil
}

func (p *LL1Parser) ParseText(input string) (*Tree[grammar.Element], error) {
	return p.Parse(strings.NewReader(input))
}

func parseTree(sentence grammar.Sentence) []*Tree[grammar.Element] {
	var trees []*Tree[grammar.Element]
	for _, s := range sentence {
		trees = append(trees, NewParseTree(s))
	}
	return trees
}
