package parser

import (
	"context"
	"testing"

	"github.com/goccy/go-graphviz"
	"github.com/vikashmadhow/lang-tools/grammar"
)

// Seq[characters] -> Lexer -> Seq[Token] -> Modulator... -> Seq[Token] -> Syntax Analyser -> ST -> Semantic Processors... -> AT -> Translators... -> Translation

func TestSimpleParser(t *testing.T) {
	g := testSimpleGrammar()
	parser := NewLL1Parser(g, g.ProdByName["e"])
	parser.PrintTable()

	expr := "x * y + z"
	tree, err := parser.ParseText(expr)
	if err != nil {
		println(err)
	}
	println(tree.GraphViz(expr))

	//pruned := tree.Map(Compose(
	//	Compact[grammar.Element],
	//	PromoteSingleChild[grammar.Element],
	//	DropOrphanNonTerminal[grammar.Element]))
	pruned := tree.Map(grammar.DropOrphanNonTerminal)
	//println(GraphViz(pruned, expr+" (drop orphan)"))

	//pruned = pruned.Map(Compact[grammar.Element])
	//println(GraphViz(pruned, "x + y (compact)"))

	pruned = pruned.Map(grammar.PromoteSingleChild)
	println(pruned.GraphViz(expr + " (promote single child)"))

	expr = "x + y * z"
	tree, err = parser.ParseText(expr)
	if err != nil {
		println(err)
	}
	//println(GraphViz(tree, expr))

	//pruned := tree.Map(Compose(
	//	Compact[grammar.Element],
	//	PromoteSingleChild[grammar.Element],
	//	DropOrphanNonTerminal[grammar.Element]))
	pruned = tree.Map(grammar.DropOrphanNonTerminal)
	//println(GraphViz(pruned, expr+" (drop orphan)"))

	//pruned = pruned.Map(Compact[grammar.Element])
	//println(GraphViz(pruned, "x + y (compact)"))

	pruned = pruned.Map(grammar.PromoteSingleChild)
	println(pruned.GraphViz(expr + " (promote single child)"))

	ctx := context.Background()
	viz, err := graphviz.New(ctx)
	if err != nil {
		panic(err)
	}
	defer viz.Close()

	dot := pruned.GraphViz(expr + " (promote single child)")
	graph, err := graphviz.ParseBytes([]byte(dot))
	if err != nil {
		panic(err)
	}

	//var buf bytes.Buffer
	//if err := viz.Render(ctx, graph, graphviz.PNG, &buf); err != nil {
	//	panic(err)
	//}
	//
	//// 2. get as image.Image instance
	//image, err := viz.RenderImage(ctx, graph)
	//if err != nil {
	//	panic(err)
	//}

	// 3. write to file directly
	if err := viz.RenderFilename(ctx, graph, graphviz.PNG, "c:/temp/lang/graph.png"); err != nil {
		panic(err)
	}

	//tree = tree.Map(DropOrphanNonTerminal[grammar.Element])
	//println(GraphViz(tree, "x + y (drop orphan)"))

	//tree, err := g.ParseTextFromStart(
	//	`let x := 1000;
	//	 let y := 2000;
	//    x = x + 5 * (4 + y / 2);
	//	 y = y + x;`,
	//)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Printf(tree.GraphViz("first program"))
}

//func TestSimpleParser(t *testing.T) {
//	g := testGrammar()
//	tree, err := g.ParseTextFromStart(
//		`let x := 1000;
//		 let y := 2000;
//	    x = x + 5 * (4 + y / 2);
//		 y = y + x;`,
//	)
//	if err != nil {
//		t.Fatal(err)
//	}
//	fmt.Printf(tree.GraphViz("first program"))
//}

//func testGrammar() *Grammar {
//	lex := NewLexer(
//		NewTokenType("LET", "let"),
//		NewTokenType("INT", "\\d+"),
//		NewTokenType("ID", "[_a-zA-Z][_a-zA-Z0-9]*"),
//		SimpleTokenType("="),
//		SimpleTokenType(":="),
//		SimpleTokenType(">"),
//		SimpleTokenType(">="),
//		SimpleTokenType("<"),
//		SimpleTokenType("<="),
//		SimpleTokenType("("),
//		SimpleTokenType(")"),
//		SimpleTokenType(";"),
//		NewTokenType("ADD", "\\+|-"),
//		NewTokenType("MUL", "\\*|/"),
//		NewTokenType("SPC", "\\s+"),
//	)
//	lex.Modulator(Ignore("SPC"))
//	return NewLexer(
//		"test_language",
//		lex,
//		[]*Production{
//			{
//				Name:     "Program",
//				Sentence: &OneOrMore{&ProductionRef{"Stmt", Retain}, Retain},
//			},
//			{
//				Name: "Stmt",
//				Sentence: &Choice{
//					Alternates: []Sentence{
//						&Sequence{Elements: []Sentence{
//							&TokenRef{"LET", Promote1},
//							&TokenRef{"ID", Retain},
//							&TokenRef{":=", Drop},
//							&ProductionRef{"Expr", Retain},
//							&TokenRef{";", Drop},
//						}},
//						&Sequence{Elements: []Sentence{
//							&TokenRef{"ID", Retain},
//							&TokenRef{"=", Promote1},
//							&ProductionRef{"Expr", Retain},
//							&TokenRef{";", Drop},
//						}},
//					},
//				},
//			},
//			{
//				Name: "Expr",
//				Sentence: &Sequence{Elements: []Sentence{
//					&ProductionRef{"Term", Retain},
//					&Optional{&Sequence{
//						Elements: []Sentence{
//							&TokenRef{"ADD", Promote2},
//							&ProductionRef{"Expr", Retain},
//						}, TreeRetention: Retain,
//					}, Retain},
//				}},
//			},
//			{
//				Name:     "Term",
//				Sentence: &OneOrMore{&ProductionRef{"Factor", Retain}, Retain},
//			},
//			{
//				Name: "Factor",
//				Sentence: &Sequence{Elements: []Sentence{
//					&ProductionRef{"Base", Retain},
//					&Optional{&Sequence{
//						Elements: []Sentence{
//							&TokenRef{"MUL", Promote2},
//							&ProductionRef{"Expr", Retain},
//						}, TreeRetention: Retain,
//					}, Retain},
//				}},
//			},
//			{
//				Name: "Base",
//				Sentence: &Choice{Alternates: []Sentence{
//					&Sequence{Elements: []Sentence{
//						&TokenRef{"(", Promote1},
//						&ProductionRef{"Expr", Retain},
//						&TokenRef{")", Drop},
//					}},
//					&TokenRef{"INT", Retain},
//					&TokenRef{"ID", Retain},
//				}},
//			},
//		},
//	)
//}

func testSimpleGrammar() *grammar.Grammar {
	// e  -> t e'
	// e' -> + t e
	//    |
	// t  -> f t'
	// t' -> * f t'
	//    |
	// f  -> ID
	//    |  ( e )
	rules := []grammar.Rule{
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
	g, err := grammar.NewGrammar("test_simple_grammar", rules)
	if err != nil {
		panic(err)
	}
	return g
}
