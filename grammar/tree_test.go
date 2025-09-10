package grammar

import (
	"fmt"
	"testing"

	"github.com/vikashmadhow/lang-tools/lexer"
)

func TestMatch(t *testing.T) {
	tree := testTree()
	fmt.Println(tree.GraphViz("Test tree"))

	index := tree.BuildIndex()
	result := index.Match([][]string{
		{"f", ANCESTOR, "e'", ANCESTOR, "ID", SIBLING, "PLUS"},
		{"f", PARENT, "t"},
		//{"f"},
	})
	//println(result)

	for _, r := range result {
		fmt.Println(r.Tree.GraphViz(r.Tree.Node.String()))
	}
}

func testTree() *Tree {
	g := testSimpleGrammar()
	return &Tree{
		Node: g.ProdByName["e"],
		Children: []*Tree{
			{
				Node: g.ProdByName["f"],
				Children: []*Tree{
					{
						Node: g.ProdByName["t"],
					},
					{
						Node: g.ProdByName["e'"],
						Children: []*Tree{
							{
								Node: g.Lexer.Type("ID"),
								Children: []*Tree{
									{
										Node: &lexer.Token{Type: g.Lexer.Type("ID"), Text: "x"},
									},
								},
							},
							{
								Node: g.Lexer.Type("PLUS"),
								Children: []*Tree{
									{
										Node: &lexer.Token{Type: g.Lexer.Type("PLUS"), Text: "+"},
									},
								},
							},
							{
								Node: g.Lexer.Type("ID"),
								Children: []*Tree{
									{
										Node: &lexer.Token{Type: g.Lexer.Type("ID"), Text: "y"},
									},
								},
							},
						},
					},
					{
						Node: g.ProdByName["f"],
					},
				},
			},
			{
				Node: g.ProdByName["e'"],
				Children: []*Tree{
					{
						Node: g.ProdByName["f"],
						Children: []*Tree{
							{
								Node: g.ProdByName["t"],
							},
							{
								Node: g.ProdByName["e'"],
								Children: []*Tree{
									{
										Node: g.Lexer.Type("ID"),
										Children: []*Tree{
											{
												Node: &lexer.Token{Type: g.Lexer.Type("PLUS"), Text: "x"},
											},
										},
									},
									{
										Node: g.Lexer.Type("PLUS"),
										Children: []*Tree{
											{
												Node: &lexer.Token{Type: g.Lexer.Type("ID"), Text: "+"},
											},
										},
									},
									{
										Node: g.Lexer.Type("ID"),
										Children: []*Tree{
											{
												Node: &lexer.Token{Type: g.Lexer.Type("ID"), Text: "y"},
											},
										},
									},
								},
							},
							{
								Node: g.ProdByName["f"],
							},
						},
					},
				},
			},
		},
	}
}

func testSimpleGrammar() *Grammar {
	// e  -> t e'
	// e' -> + t e
	//    |
	// t  -> f t'
	// t' -> * f t'
	//    |
	// f  -> ID
	//    |  ( e )
	rules := []Rule{
		{"e", [][]string{{"t", "e'"}}},
		{"e'", [][]string{{"PLUS", "t", "e'"}, {}}},
		{"t", [][]string{{"f", "t'"}}},
		{"t'", [][]string{{"TIME", "f", "t'"}, {}}},
		{"f", [][]string{{"ID"}, {"OPEN", "e", "CLOSE"}}},
		{"PLUS", [][]string{{"\\+"}}},
		{"TIME", [][]string{{"\\*"}}},
		{"OPEN", [][]string{{"\\("}}},
		{"CLOSE", [][]string{{"\\)"}}},
		{"ID", [][]string{{"[_a-zA-Z][_a-zA-Z0-9]*"}}},
		{"SPC", [][]string{{"\\s+"}, {"#Ignore"}}},
	}
	g, err := NewGrammar("test_simple_grammar", rules)
	if err != nil {
		panic(err)
	}
	return g
}
