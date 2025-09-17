package parser

import (
	"fmt"
	"io"
	"slices"
	"strconv"
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
	followMap[start] = map[*lexer.TokenType]bool{lexer.TextEndType: true}

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
				for i := len(alt.Elements) - 1; i >= 0; i-- {
					if i == len(alt.Elements)-1 {
						changed = changed || addFollow(followMap, alt.Elements[i], p, false)
					} else {
						changed = changed || addFollow(followMap, alt.Elements[i], alt.Elements[i+1], true)
						if alt.Elements[i+1].MatchEmpty() {
							changed = changed || addFollow(followMap, alt.Elements[i], alt.Elements[i+1], false)
						}
					}
				}
			}
		}
	}

	table := make(map[grammar.Element]map[*lexer.TokenType]grammar.Sentence)
	for _, p := range g.Productions {
		empty := false
		table[p] = make(map[*lexer.TokenType]grammar.Sentence)
		for _, alt := range p.Alternates {
			if alt.Empty() {
				empty = true
			} else {
				for t := range alt.First() {
					table[p][t] = alt
				}
			}
		}
		if empty {
			for t := range followMap[p] {
				table[p][t] = grammar.Sentence{
					Elements: []grammar.Element{},
				}
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

func (p *LL1Parser) PrintTable() {
	var tokens []*lexer.TokenType
	for _, tableTokens := range p.follow {
		for t := range tableTokens {
			if !slices.Contains(tokens, t) {
				tokens = append(tokens, t)
			}
		}
	}

	const space = 10
	fmt.Printf("%"+strconv.Itoa(space)+"s\t", "Production")
	for _, t := range tokens {
		fmt.Printf(" %-"+strconv.Itoa(space)+"s\t", t)
	}
	fmt.Println()
	for g := range p.LL1ParseTable {
		fmt.Printf("%"+strconv.Itoa(space)+"s\t", g)
		for _, t := range tokens {
			if s, ok := p.LL1ParseTable[g][t]; ok {
				fmt.Printf(" %-"+strconv.Itoa(space)+"s\t", s)
			} else {
				fmt.Printf(" %-"+strconv.Itoa(space)+"s\t", "")
			}
		}
		fmt.Println()
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

func (p *LL1Parser) Parse(input io.Reader) (*grammar.Tree, error) {
	//for el, tokens := range p.follow {
	//	fmt.Printf("%10s\t", el)
	//	for t := range tokens {
	//		fmt.Printf("%-10s\t", t)
	//	}
	//	fmt.Println()
	//}

	start := grammar.NewParseTree(p.Start)
	stack := []*grammar.Tree{start, grammar.NewParseTree(lexer.TextEndType)}
	for token, err := range p.Grammar.Lexer.LexSeq(input) {
		if err != nil {
			return start, err
		}
		//fmt.Println("Token:", token)
		for len(stack) > 0 && token.Type != stack[0].Node {
			top := stack[0]
			stack = stack[1:]

			//fmt.Println("Top:", top)
			sentence, ok := p.LL1ParseTable[top.Node.(grammar.Element)][token.Type]
			//fmt.Println("Expand to:", sentence)

			if !ok {
				return nil, fmt.Errorf("no production for %q -> %q", top.Node, token.Type.Id)
			}
			top.Children = parseTree(sentence)
			stack = slices.Insert(stack, 0, top.Children...)
		}
		stack[0].Children = append(stack[0].Children, grammar.NewParseTree(token))
		stack = stack[1:]
	}
	return start.Map(grammar.CleanUp), nil
}

func (p *LL1Parser) ParseText(input string) (*grammar.Tree, error) {
	return p.Parse(strings.NewReader(input))
}

func parseTree(sentence grammar.Sentence) []*grammar.Tree {
	var trees []*grammar.Tree
	for _, s := range sentence.Elements {
		trees = append(trees, grammar.NewParseTree(s))
	}
	return trees
}
