package lexer

import (
	"errors"
	"unicode"

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
	Empty     = rune(unicode.Co.R16[0].Lo + 125)
	Unknown   = rune(unicode.Co.R16[0].Lo + 126)
	TextStart = rune(unicode.Co.R16[0].Lo + 127) // Emitted once at the start of matching the text
	TextEnd   = rune(unicode.Co.R16[0].Lo + 128) // Emitted at the end of matching the text
	LineStart = rune(unicode.Co.R16[0].Lo + 129) // Emitted at the start of matching a line
	LineEnd   = rune(unicode.Co.R16[0].Lo + 130) // Emitted at the end of matching a line

	//Whitespace = NewTokenType("WS", "[ \t\n\r]+")
	//Identifier = NewTokenType("IDENTIFIER", "[a-zA-Z_][a-zA-Z0-9_]*")
	//Number     = NewTokenType("NUMBER", "[0-9]+")

	EmptyType   = &TokenType{string(Empty), "", regex.NewRegex(string(Empty))}
	UnknownType = &TokenType{string(Unknown), "", regex.NewRegex(string(Unknown))}
	TextEndType = &TokenType{string(TextEnd), "$", regex.NewRegex(string(TextEnd))}
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
	return t == EmptyType
}

func (t *TokenType) MatchEmpty() bool {
	return t == EmptyType || t.Compiled.MatchEmpty()
}

func (t *TokenType) First() map[*TokenType]bool {
	return map[*TokenType]bool{t: true}
}

func (t *TokenType) TreePattern() string {
	return t.Id
}

func (t *TokenType) String() string {
	return t.Id
}

func (t *Token) Terminal() bool {
	return true
}

func (t *Token) Empty() bool {
	return t.Type == EmptyType
}

func (t *Token) MatchEmpty() bool {
	return t.Type == EmptyType || t.Type.Compiled.MatchEmpty()
}

func (t *Token) First() map[*TokenType]bool {
	return map[*TokenType]bool{t.Type: true}
}

func (t *Token) TreePattern() string {
	return t.Type.Id
}

func (t *Token) String() string {
	return t.Text
}
