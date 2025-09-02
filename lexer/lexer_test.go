// author: Vikash Madhow (vikash.madhow@gmail.com)

package lexer

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/vikashmadhow/lang-tools/seq"
)

func TestBasicLexer(t *testing.T) {
	l := NewLexer(
		&TokenType{Id: "LET", Pattern: "let"},
		&TokenType{Id: "INT", Pattern: "[0-9]+"},
		&TokenType{Id: "ID", Pattern: "[_a-zA-Z][_a-zA-Z0-9]*"},
		&TokenType{Id: "EQ", Pattern: "="},
		&TokenType{Id: "SPC", Pattern: "\\s+"},
	)

	var tokens []*Token
	//tokenSeq := l.LexText("let x =  1000")
	//defer tokenSeq.Stop()
	//for token := range seq.Push2(tokenSeq.Next, tokenSeq.Stop) {
	//	tokens = append(tokens, token)
	//}

	for token := range l.LexTextSeq("let x =  1000") {
		tokens = append(tokens, token)
	}

	_, err := matchTokens(tokens, []*Token{
		{l.Type("LET"), "let", 1, 1},
		{l.Type("SPC"), " ", 1, 4},
		{l.Type("ID"), "x", 1, 5},
		{l.Type("SPC"), " ", 1, 6},
		{l.Type("EQ"), "=", 1, 7},
		{l.Type("SPC"), "  ", 1, 8},
		{l.Type("INT"), "1000", 1, 10},
		{EOFType, "$", 1, 14},
	})

	if err != nil {
		t.Error(err)
	}
}

func TestLexerError(t *testing.T) {
	l := NewLexer(
		&TokenType{Id: "LET", Pattern: "let"},
		&TokenType{Id: "INT", Pattern: "[0-9]+"},
		&TokenType{Id: "ID", Pattern: "[_a-zA-Z][_a-zA-Z0-9]*"},
		&TokenType{Id: "EQ", Pattern: "="},
		&TokenType{Id: "SPC", Pattern: "\\s+"},
	)

	var tokens []*Token
	tokenSeq := l.LexText("let x? =  1000")
	for token, e := range seq.Push2(tokenSeq.Next, tokenSeq.Stop) {
		if e != nil {
			println(e.Error())
			return
		}
		fmt.Println(token)
		tokens = append(tokens, token)
	}
}

func TestMultiline(t *testing.T) {
	l := NewLexer(
		&TokenType{Id: "LET", Pattern: "let"},
		&TokenType{Id: "INT", Pattern: "\\d+"},
		&TokenType{Id: "ID", Pattern: "[_a-zA-Z][_a-zA-Z0-9]*"},
		&TokenType{Id: "EQ", Pattern: "="},
		&TokenType{Id: "PLUS", Pattern: "\\+|-"},
		&TokenType{Id: "TIME", Pattern: "\\*|/"},
		&TokenType{Id: "SPC", Pattern: "\\s+"},
	)
	//l.Modulator(func(token regex.Token, err error) []seq.Pair[regex.Token, error] {
	//	if token.Type == "SPC" {
	//		return nil
	//	} else {
	//		return []seq.Pair[regex.Token, error]{{token, err}}
	//	}
	//})

	l.Modulator(Ignore(l.Type("SPC")))

	var tokens []*Token
	//tokenSeq := l.LexText(`let x = 1000
	//						 let y =x+y*-2000`)
	//for token := range seq.Push2(tokenSeq.Next, tokenSeq.Stop) {
	//	//if token.Type != "SPC" {
	//	tokens = append(tokens, token)
	//	//}
	//}

	for token := range l.LexTextSeq(`let x = 1000
							 let y =x+y*-2000`) {
		tokens = append(tokens, token)
	}

	//fmt.Println(tokens)
	_, err := matchTokens(tokens, []*Token{
		{l.Type("LET"), "let", 1, 1},
		//{"SPC", " ", 1, 4},
		{l.Type("ID"), "x", 1, 5},
		//{"SPC", " ", 1, 6},
		{l.Type("EQ"), "=", 1, 7},
		//{"SPC", " ", 1, 8},
		{l.Type("INT"), "1000", 1, 9},
		//{"SPC", "\n\t\t\t\t\t\t\t ", 2, 0},
		{l.Type("LET"), "let", 2, 9},
		//{"SPC", " ", 2, 12},
		{l.Type("ID"), "y", 2, 13},
		//{"SPC", " ", 2, 14},
		{l.Type("EQ"), "=", 2, 15},
		{l.Type("ID"), "x", 2, 16},
		{l.Type("PLUS"), "+", 2, 17},
		{l.Type("ID"), "y", 2, 18},
		{l.Type("TIME"), "*", 2, 19},
		{l.Type("PLUS"), "-", 2, 20},
		{l.Type("INT"), "2000", 2, 21},
		{EOFType, "$", 2, 25},
	})

	if err != nil {
		t.Error(err)
	}
}

func TestUnicode(t *testing.T) {
	l := NewLexer(
		&TokenType{Id: "LET", Pattern: "let"},
		&TokenType{Id: "INT", Pattern: "\\d+"},
		&TokenType{Id: "ID", Pattern: "[_a-zA-Z]\\S*"},
		&TokenType{Id: "EQ", Pattern: "="},
		&TokenType{Id: "PLUS", Pattern: "\\+|-"},
		&TokenType{Id: "TIME", Pattern: "\\*|/"},
		&TokenType{Id: "SPC", Pattern: "\\s+"},
	)
	l.Buffer(3)
	l.Modulator(Ignore(l.Type("SPC")))

	var tokens []*Token
	for token := range l.LexTextSeq(`let A日本語 = 1000`) {
		tokens = append(tokens, token)
	}

	fmt.Println(tokens)
	_, err := matchTokens(tokens, []*Token{
		{l.Type("LET"), "let", 1, 1},
		{l.Type("ID"), "A日本語", 1, 5},
		{l.Type("EQ"), "=", 1, 10},
		{l.Type("INT"), "1000", 1, 12},
		{EOFType, "$", 1, 16},
	})

	if err != nil {
		t.Error(err)
	}
}

func TestReverse(t *testing.T) {
	l := NewLexer(
		&TokenType{Id: "LET", Pattern: "let"},
		&TokenType{Id: "INT", Pattern: "\\d+"},
		&TokenType{Id: "ID", Pattern: "[_a-zA-Z]\\S*"},
		&TokenType{Id: "EQ", Pattern: "="},
		&TokenType{Id: "PLUS", Pattern: "\\+|-"},
		&TokenType{Id: "TIME", Pattern: "\\*|/"},
		&TokenType{Id: "SPC", Pattern: "\\s+"},
	)
	//l.Buffer(3)
	l.Modulator(Ignore(l.Type("SPC")), Reverse())

	var tokens []*Token
	for token := range l.LexTextSeq(`let A日本語 = 1000`) {
		tokens = append(tokens, token)
	}

	fmt.Println(tokens)
	_, err := matchTokens(tokens, []*Token{
		{l.Type("INT"), "1000", 1, 12},
		{l.Type("EQ"), "=", 1, 10},
		{l.Type("ID"), "A日本語", 1, 5},
		{l.Type("LET"), "let", 1, 1},
	})

	if err != nil {
		t.Error(err)
	}
}

func TestReverseAlternate(t *testing.T) {
	l := NewLexer(
		&TokenType{Id: "LET", Pattern: "let"},
		&TokenType{Id: "INT", Pattern: "\\d+"},
		&TokenType{Id: "ID", Pattern: "[_a-zA-Z]\\S*"},
		&TokenType{Id: "EQ", Pattern: "="},
		&TokenType{Id: "PLUS", Pattern: "\\+|-"},
		&TokenType{Id: "TIME", Pattern: "\\*|/"},
		&TokenType{Id: "SPC", Pattern: "\\s+"},
	)
	//l.Buffer(3)
	l.Modulator(Ignore(l.Type("SPC")))

	l.Modulator(func() Modulator {
		var stream []seq.Pair[*Token, error] = nil
		return func(t *Token, err error) []seq.Pair[*Token, error] {
			if t.Type == EOFType {
				slices.Reverse(stream)
				return stream
			} else {
				stream = append(stream, seq.Pair[*Token, error]{A: t, B: err})
				if len(stream) >= 2 {
					ret := make([]seq.Pair[*Token, error], len(stream))
					copy(ret, stream)
					slices.Reverse(ret)
					stream = nil
					return ret
				}
				return nil
			}
		}
	}())

	var tokens []*Token
	for token := range l.LexTextSeq(`let A日本語 = 1000 +`) {
		tokens = append(tokens, token)
	}

	fmt.Println(tokens)
	_, err := matchTokens(tokens, []*Token{
		{l.Type("ID"), "A日本語", 1, 5},
		{l.Type("LET"), "let", 1, 1},
		{l.Type("INT"), "1000", 1, 12},
		{l.Type("EQ"), "=", 1, 10},
		{l.Type("PLUS"), "+", 1, 17},
	})

	if err != nil {
		t.Error(err)
	}
}

func TestEndError(t *testing.T) {
	l := NewLexer(
		&TokenType{Id: "LET", Pattern: "let"},
		&TokenType{Id: "INT", Pattern: "\\d+"},
		&TokenType{Id: "ID", Pattern: "[_a-zA-Z][_a-zA-Z0-9]*"},
		&TokenType{Id: "EQ", Pattern: ":="},
		&TokenType{Id: "EQ_PLUS", Pattern: ":\\+"},
		&TokenType{Id: "PLUS", Pattern: "\\+|-"},
		&TokenType{Id: "TIME", Pattern: "\\*|/"},
		&TokenType{Id: "SPC", Pattern: "\\s+"},
	)

	var tokens []*Token
	tokenSeq := l.LexText(`let x : 1000 :`)
	for token, e := range seq.Push2(tokenSeq.Next, tokenSeq.Stop) {
		if e != nil {
			println(e.Error())
			return
		}
		if token.Type != l.Type("SPC") {
			tokens = append(tokens, token)
		}
	}

	_, err := matchTokens(tokens, []*Token{
		{l.Type("LET"), "let", 1, 1},
		{l.Type("ID"), "x", 1, 5},
		{l.Type("EQ"), ":=", 1, 7},
		{l.Type("INT"), "1000", 1, 10},
		{EOFType, "$", 1, 14},
	})

	if err != nil {
		t.Error(err)
	}
}

func matchTokens(t1 []*Token, t2 []*Token) (bool, error) {
	if len(t1) != len(t2) {
		return false, errors.New(fmt.Sprint("comparing different number of tokens:", len(t1), ",", len(t2)))
	}
	for i, token := range t1 {
		if *t2[i] != *token {
			return false, errors.New(fmt.Sprint("failed at position:", i, ",", token, "!=", t2[i]))
		}
	}
	return true, nil
}

func String(tokens []*Token) string {
	var s strings.Builder
	for _, token := range tokens {
		//if i > 0 {
		//	s.WriteString(" ")
		//}
		s.WriteString(token.String())
	}
	return s.String()
}
