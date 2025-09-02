package lexer

import (
	"slices"

	"github.com/vikashmadhow/lang-tools/seq"
)

// Modulator is a function that can modify the token stream produced by a lexer. It
// takes a token and maps it to a sequence of tokens. Mapping to an empty sequence
// (or nil) removes the token from the stream. Otherwise, the mapped sequence of tokens
// is merged into the token stream. This is similar to a flat-map operation on the token
// stream. Multiple modulators can be set on a lexer, and they are invoked in the same
// order that they were installed.
type Modulator func(*Token, error) []seq.Pair[*Token, error]

// Ignore is a Modulator that removes the specified token types from the token stream.
// It is useful to remove syntactically useless tokens in some languages such as whitespace.
func Ignore(types ...*TokenType) Modulator {
	ignore := map[*TokenType]bool{}
	for _, t := range types {
		ignore[t] = true
	}
	return func(t *Token, e error) []seq.Pair[*Token, error] {
		if _, ok := ignore[t.Type]; ok {
			return nil
		}
		return []seq.Pair[*Token, error]{{t, e}}
	}
}

// Reverse is an example Modulator that reverses the token stream. It works by holding
// all tokens from the stream in a slice which it then reverses when the Lexer sends the
// EOF token at the end of lexing.
func Reverse() Modulator {
	var stream []seq.Pair[*Token, error] = nil
	return func(t *Token, err error) []seq.Pair[*Token, error] {
		if t.Type == EOFType {
			slices.Reverse(stream)
			return stream
		} else {
			stream = append(stream, seq.Pair[*Token, error]{A: t, B: err})
			return nil
		}
	}
}
