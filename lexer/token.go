package lexer

import (
	"errors"

	"github.com/vikashmadhow/lang-tools/regex"
	"github.com/vikashmadhow/lang-tools/seq"
)

type (
	Token struct {
		Type   *TokenType
		Text   string
		Line   int
		Column int
	}

	TokenType struct {
		Id       string
		Pattern  string
		Compiled *regex.CompiledRegex
	}

	TokenSeq struct {
		next seq.Seq2[*Token, error]
		stop func()
		//pushedBack []*Token
		pushedBack chan *Token
	}

	TokenMatcher struct {
		def     *TokenType
		matcher *regex.Matcher
	}
)

var (
	Empty     = "∅"
	EOF       = "Ω"
	EmptyType = &TokenType{Empty, "", regex.NewRegex("")}
	EOFType   = &TokenType{EOF, "$", regex.NewRegex("$")}
)

func SimpleTokenType(id string) *TokenType {
	return NewTokenType(id, regex.Escape(id))
}

func NewTokenType(id string, pattern string) *TokenType {
	return &TokenType{id, pattern, regex.NewRegex(pattern)}
}

func (t *TokenSeq) Next() (*Token, error, bool) {
	if len(t.pushedBack) > 0 {
		//token := <- t.pushedBack[len(t.pushedBack)-1]
		//t.pushedBack = t.pushedBack[:len(t.pushedBack)-1]
		return <-t.pushedBack, nil, true
	}
	//return t.next()
	token, err, valid := t.next()
	if err != nil {
		return nil, err, valid
	}
	if !valid {
		return nil, errors.New("lexer returned an invalid token"), valid
	}
	return token, nil, valid
}

func (t *TokenSeq) Peek() (*Token, error, bool) {
	token, err, valid := t.Next()
	if err != nil {
		return nil, err, valid
	}
	t.Pushback(token)
	return token, nil, valid
}

func (t *TokenSeq) Pushback(token *Token) {
	//t.pushedBack = append(t.pushedBack, token)
	t.pushedBack <- token
}

func (t *TokenSeq) Stop() {
	t.stop()
}

func (t *TokenType) Terminal() bool {
	return true
}

func (t *TokenType) Empty() bool {
	return t.Id == Empty
}

func (t *TokenType) MatchEmpty() bool {
	return t.Id == Empty || t.Compiled.MatchEmpty()
}

func (t *TokenType) First() map[*TokenType]bool {
	return map[*TokenType]bool{t: true}
}

func (t *TokenType) String() string {
	return t.Id
}

func (t *Token) Terminal() bool {
	return true
}

func (t *Token) Empty() bool {
	return t.Type.Id == Empty
}

func (t *Token) MatchEmpty() bool {
	return t.Type.Id == Empty || t.Type.Compiled.MatchEmpty()
}

func (t *Token) First() map[*TokenType]bool {
	return map[*TokenType]bool{t.Type: true}
}

func (t *Token) String() string {
	return t.Text
}
