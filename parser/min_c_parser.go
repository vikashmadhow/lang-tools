package parser

import (
	"github.com/vikashmadhow/lang-tools/grammar"
)

func minCGrammar() *grammar.Grammar {
	rules := []grammar.Rule{
		{"program", [][]string{{"stmtList"}}},
		{"stmtList", [][]string{{"stmt", "stmtList"}, {}}},
		{"stmt", [][]string{
			{"ID", "ASSIGN", "expr", "SEMICOLON"},
			{"IF", "OPEN", "expr", "CLOSE", "LBRACE", "stmtList", "RBRACE", "elsePart"},
			{"WHILE", "OPEN", "expr", "CLOSE", "LBRACE", "stmtList", "RBRACE"},
		}},
		{"elsePart", [][]string{{"ELSE", "LBRACE", "stmtList", "RBRACE"}, {}}},
		{"expr", [][]string{{"term", "exprTail"}}},
		{"exprTail", [][]string{
			{"PLUS", "term", "exprTail"},
			{"MINUS", "term", "exprTail"},
			{},
		}},
		{"term", [][]string{{"factor", "termTail"}}},
		{"termTail", [][]string{
			{"MUL", "factor", "termTail"},
			{"DIV", "factor", "termTail"},
			{},
		}},
		{"factor", [][]string{
			{"ID"},
			{"NUM"},
			{"OPEN", "expr", "CLOSE"}}},

		{"IF", [][]string{{"if"}}},
		{"ELSE", [][]string{{"else"}}},
		{"WHILE", [][]string{{"while"}}},

		{"ASSIGN", [][]string{{"="}}},
		{"PLUS", [][]string{{"\\+"}}},
		{"MINUS", [][]string{{"\\-"}}},
		{"MUL", [][]string{{"\\*"}}},
		{"DIV", [][]string{{"/"}}},
		{"OPEN", [][]string{{"\\("}}},
		{"CLOSE", [][]string{{"\\)"}}},
		{"LBRACE", [][]string{{"\\{"}}},
		{"RBRACE", [][]string{{"\\}"}}},
		{"SEMICOLON", [][]string{{";"}}},
		{"NUM", [][]string{{"[0-9]+"}}},
		{"ID", [][]string{{"[a-zA-Z_][a-zA-Z0-9_]*"}}},

		{"SPC", [][]string{{"\\s+"}, {"#Ignore"}}},
	}

	// Grammar rules
	g, err := grammar.NewGrammar("min_c_grammar", rules)
	if err != nil {
		panic(err)
	} else {
		return g
	}
	// Test cases
	//tests := []struct {
	//    input   string
	//    wantErr bool
	//}{
	//    {
	//        input: `main() {
	//                    x = 1;
	//                    y = 2 * 3;
	//                    z = (x + y) / 2;
	//                }`,
	//        wantErr: false,
	//    },
	//}
	//
	//for _, tt := range tests {
	//    _, err := parser.ParseText(tt.input)
	//    if (err != nil) != tt.wantErr {
	//        t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
	//    }
	//}
}
