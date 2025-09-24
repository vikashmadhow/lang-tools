package parser

import (
	"fmt"
	"testing"

	"github.com/vikashmadhow/lang-tools/grammar"
)

func TestMinCGrammar(t *testing.T) {
	// Program structure
	g := minLangGrammar()
	parser := NewLL1Parser(g, g.ProdByName["program"])

	//parser.PrintTable()

	tree, err := parser.ParseText(TestProg1)
	if err != nil {
		println(err.Error())
	} else {
		fmt.Println(tree.GraphViz("Test min C prog"))
	}

	pruned := tree.Map(grammar.DropOrphanNonTerminal)
	//println(GraphViz(pruned, expr+" (drop orphan)"))

	//pruned = pruned.Map(Compact[grammar.Element])
	//println(GraphViz(pruned, "x + y (compact)"))

	pruned = pruned.Map(grammar.PromoteSingleChild)
	println(pruned.GraphViz("Pruned"))
}

func TestRestructureTree(t *testing.T) {
	// Program structure
	g := minLangGrammar()
	parser := NewLL1Parser(g, g.ProdByName["program"])

	//parser.PrintTable()

	tree, err := parser.ParseText(TestProg1)
	if err != nil {
		println(err.Error())
	}

	tree = tree.Map(grammar.DropOrphanNonTerminal)
	tree = tree.Map(grammar.PromoteSingleChild)
	println(tree.GraphViz("Pruned"))

	tree = tree.MapMatch(func(t *grammar.Tree, result *grammar.MatchResult) *grammar.Tree {
		return &grammar.Tree{
			Node:     t.Children[0].Node,
			Children: t.Children[1:],
		}
	}, [][]string{{"stmt", grammar.PARENT, "FUNC"}})
	println(tree.GraphViz("Transformed"))

}

const TestProg1 = `
    var global := 1000;

    func main() { 
       var x := 1000;      
       var z := 1000;      
       if z + 5 > x {
          x = x + 1;
       } else {
          x = x - 1;
       }
       return x;   
    }

    func factorial(n) {
        if n == 0 {
            return 1;
        } else {
            return n * call factorial(n - 1);
        }
    }
    `
