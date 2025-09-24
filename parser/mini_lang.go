package parser

import (
	"github.com/vikashmadhow/lang-tools/grammar"
)

func minLangGrammar() *grammar.Grammar {
	rules := []grammar.Rule{
		{"program", [][]string{{"stmtList"}}},
		{"stmtList", [][]string{{"stmt", "stmtList"}, {}}},
		{"stmt", [][]string{
			{"RETURN", "expr", "SEMICOLON"},
			{"VAR", "ID", "DECLARE", "expr", "SEMICOLON"},
			{"ID", "ASSIGN", "expr", "SEMICOLON"},
			{"FUNC", "ID", "OPEN", "argList", "CLOSE", "LBRACE", "stmtList", "RBRACE"},
			{"IF", "expr", "LBRACE", "stmtList", "RBRACE", "elsePart"},
			{"WHILE", "expr", "LBRACE", "stmtList", "RBRACE"},
		}},
		{"argList", [][]string{{"ID", "argList"}, {}}},
		{"exprList", [][]string{{"expr", "exprList"}, {}}},
		{"elsePart", [][]string{{"ELSE", "LBRACE", "stmtList", "RBRACE"}, {}}},

		{"expr", [][]string{{"term", "exprTail"}}},
		{"exprTail", [][]string{
			{"PLUS", "term", "exprTail"},
			{"MINUS", "term", "exprTail"},
			{},
		}},

		{"expr", [][]string{{"eq", "exprTail"}}},
		{"exprTail", [][]string{
			{"EQUAL", "eq", "exprTail"},
			{},
		}},
		{"eq", [][]string{{"comp", "eqTail"}}},
		{"eqTail", [][]string{
			{"GREATER", "comp", "eqTail"},
			{"GREATER_OR_EQUAL", "comp", "eqTail"},
			{"LESS", "comp", "eqTail"},
			{"LESS_OR_EQUAL", "comp", "eqTail"},
			{},
		}},

		{"comp", [][]string{{"term", "compTail"}}},
		{"compTail", [][]string{
			{"PLUS", "term", "compTail"},
			{"MINUS", "term", "compTail"},
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
			{"CALL", "ID", "OPEN", "exprList", "CLOSE"},
			{"OPEN", "expr", "CLOSE"}}},

		{"IF", [][]string{{"if"}}},
		{"ELSE", [][]string{{"else"}}},
		{"WHILE", [][]string{{"while"}}},
		{"VAR", [][]string{{"var"}}},
		{"FUNC", [][]string{{"func"}}},
		{"RETURN", [][]string{{"return"}}},
		{"CALL", [][]string{{"call"}}},

		{"ASSIGN", [][]string{{"="}}},
		{"DECLARE", [][]string{{":="}}},
		{"EQUAL", [][]string{{"=="}}},
		{"GREATER", [][]string{{">"}}},
		{"GREATER_OR_EQUAL", [][]string{{">="}}},
		{"LESS", [][]string{{"<"}}},
		{"LESS_OR_EQUAL", [][]string{{"<="}}},
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
