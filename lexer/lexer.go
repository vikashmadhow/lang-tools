// author: Vikash Madhow (vikash.madhow@gmail.com)

// Package lexer implements a simple fast lexer using the prefix regular expression matcher.
package lexer

import (
	"bufio"
	"errors"
	"io"
	"iter"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/vikashmadhow/lang-tools/regex"
	"github.com/vikashmadhow/lang-tools/seq"
)

type Lexer struct {
	Definition []*TokenType
	TokenTypes map[string]*TokenType
	matchers   []*TokenMatcher
	modulators []Modulator
	bufferSize int
}

const minBufferSize = 8

func NewLexer(definition ...*TokenType) *Lexer {
	var matchers []*TokenMatcher
	for _, d := range definition {
		if d.Compiled == nil {
			d.Compiled = regex.NewRegex(d.Pattern)
		}
		matchers = append(matchers, &TokenMatcher{d, d.Compiled.Matcher()})
	}
	tokenTypes := make(map[string]*TokenType)
	for _, d := range definition {
		tokenTypes[d.Id] = d
	}
	return &Lexer{definition, tokenTypes, matchers, nil, 1024}
}

func NewLexerFromPatterns(patterns ...string) *Lexer {
	var tokens []*TokenType
	for i, r := range patterns {
		tokens = append(tokens, NewTokenType("p"+strconv.Itoa(i), r))
	}
	return NewLexer(tokens...)
}

func (lexer *Lexer) Buffer(size int) {
	lexer.bufferSize = size
}

func (lexer *Lexer) Modulator(modulator ...Modulator) {
	lexer.modulators = append(lexer.modulators, modulator...)
}

func (lexer *Lexer) Type(tokenType string) *TokenType {
	return lexer.TokenTypes[tokenType]
}

func (lexer *Lexer) LexText(input string) *TokenSeq {
	return lexer.Lex(strings.NewReader(input))
}

func (lexer *Lexer) LexTextSeq(input string) iter.Seq2[*Token, error] {
	return lexer.LexSeq(strings.NewReader(input))
}

func (lexer *Lexer) Lex(in io.Reader) *TokenSeq {
	next, stop := iter.Pull2(lexer.lex(in, false))
	if lexer.modulators != nil {
		for _, m := range lexer.modulators {
			next = seq.FlatMap2(next, m)
		}
	}
	return &TokenSeq{
		next:       next,
		stop:       stop,
		pushedBack: make(chan *Token, 64),
	}
}

func (lexer *Lexer) LexSeq(in io.Reader) iter.Seq2[*Token, error] {
	next := lexer.lex(in, false)
	if lexer.modulators != nil {
		for _, m := range lexer.modulators {
			next = seq.FlatMapSeq2(next, m)
		}
	}
	return next
}

func (lexer *Lexer) lex(in io.Reader, matchUnknown bool) iter.Seq2[*Token, error] {
	return func(yield func(t *Token, e error) bool) {
		column, line := 1, 1
		scanner := bufio.NewReader(in)

		var noMatch strings.Builder

		lastFullMatchPosition := -1
		var lastFullMatchToken *TokenType
		var lastFullMatch strings.Builder

		bufferSize := lexer.bufferSize
		if bufferSize < minBufferSize {
			bufferSize = minBufferSize
		}
		input := make([]byte, bufferSize)

		var emitted, matched, read int
		var err error
		for err == nil {
			if emitted > 0 {
				// anything before emitted can be discarded, which will free space
				// for reading more characters
				copy(input, input[emitted:])
				matched -= emitted
				read -= emitted
				if lastFullMatchPosition != -1 {
					lastFullMatchPosition -= emitted
				}
				emitted = 0
			}
			if read > int(float32(len(input))*0.75) {
				newInput := make([]byte, int(float32(len(input))*1.5))
				copy(newInput, input[:read])
				input = newInput
			}

			var readCount int
			readCount, err = scanner.Read(input[read:])
			read += readCount
			for matched < read {
				r, n := utf8.DecodeRune(input[matched:read])
				if r == utf8.RuneError {
					break
				}
				matched += n
				noneMatch := true
				for _, m := range lexer.matchers {
					if m.matcher.LastMatch != regex.NoMatch {
						match := m.matcher.MatchNext(r)
						if match == regex.FullMatch {
							if m.matcher.FullMatch.Len() > lastFullMatch.Len() {
								lastFullMatchPosition = matched
								lastFullMatch = m.matcher.FullMatch
								lastFullMatchToken = m.def
							}
						}
						if match != regex.NoMatch {
							noneMatch = false
						}
					}
				}

				if lastFullMatchPosition == -1 {
					noMatch.WriteRune(r)
					if noneMatch {
						lexer.reset()
					}
				} else {
					if noMatch.Len() > 0 {
						// Current character fails to produce but previous character did:
						// We can emit a token ending with the previous character and start
						// a new token with the current character.
						if noMatch.Len() > lastFullMatch.Len()-n {
							unknown := noMatch.String()[0:(noMatch.Len() - lastFullMatch.Len() + n)]
							t, e := lexer.produceErrorToken(unknown, !matchUnknown, line, column)
							if !yield(t, e) || (e != nil && !matchUnknown) {
								return
							}
							emitted += len(unknown)
						}
						noMatch.Reset()
					} else if noneMatch {
						t := lexer.produceToken(lastFullMatchToken, lastFullMatch.String(), line, column)
						if !yield(t, nil) {
							return
						}
						emitted = lastFullMatchPosition
						matched = lastFullMatchPosition

						lastFullMatchPosition = -1
						lastFullMatch.Reset()
						lastFullMatchToken = nil
					}
				}
				if !noneMatch {
					if r == '\n' {
						line++
						column = 1
					} else {
						column++
					}
				}
			}
		}
		if lastFullMatch.Len() == 0 {
			if noMatch.Len() > 0 {
				unknown := noMatch.String()
				t, e := lexer.produceErrorToken(unknown, !matchUnknown, line, column)
				if !yield(t, e) || (e != nil && !matchUnknown) {
					return
				}
			}
		} else {
			t := lexer.produceToken(lastFullMatchToken, lastFullMatch.String(), line, column)
			if !yield(t, nil) {
				return
			}
		}
		yield(&Token{Type: TextEndType, Text: "", Line: line, Column: column}, nil)
	}
}

func Tokenize(s string, regex ...string) iter.Seq[string] {
	return func(yield func(t string) bool) {
		lexer := NewLexerFromPatterns(regex...)
		for t := range lexer.lex(strings.NewReader(s), true) {
			if t.Type != TextEndType {
				if !yield(t.Text) {
					return
				}
			}
		}
	}
}

func Split(s string, regex ...string) iter.Seq[string] {
	return func(yield func(t string) bool) {
		lexer := NewLexerFromPatterns(regex...)
		for t, e := range lexer.lex(strings.NewReader(s), true) {
			if e != nil {
				if !yield(t.Text) {
					return
				}
			}
		}
	}
}

func (lexer *Lexer) produceToken(token *TokenType, text string, line, column int) *Token {
	lexer.reset()
	return &Token{token, text, line, column - utf8.RuneCountInString(text)}
}

func (lexer *Lexer) produceErrorToken(noMatch string, showErrorMessage bool, line, column int) (*Token, error) {
	var msg string
	if showErrorMessage {
		msg = lexer.errorMessage(noMatch, line, column)
	}
	lexer.reset()
	return &Token{
			Type:   UnknownType,
			Text:   noMatch,
			Line:   line,
			Column: column,
		},
		errors.New(msg)
}

func (lexer *Lexer) errorMessage(noMatch string, line, column int) string {
	var msg strings.Builder
	first := true
	if len(noMatch) > 0 {
		msg.WriteString("error at [" + strconv.Itoa(line) + ":" + strconv.Itoa(column-len(noMatch)) + "]")
		msg.WriteString(": unmatched text: " + noMatch)
		first = false
	} else {
		msg.WriteString("error at " + strconv.Itoa(line) + ":" + strconv.Itoa(column))
	}
	for _, m := range lexer.matchers {
		if m.matcher.LastMatch == regex.PartialMatch {
			if first {
				msg.WriteString(": potential partial match(es): ")
				first = false
			} else {
				msg.WriteString(", ")
			}
			trans := m.matcher.Compiled.Dfa.Trans[m.matcher.State]
			msg.WriteString(m.def.Id + " (next expected character(s): ")
			f := true
			for k := range trans {
				if f {
					f = false
				} else {
					msg.WriteString(", ")
				}
				msg.WriteString(k.Pattern())
			}
			msg.WriteRune(')')
		}
	}
	return msg.String()
}

func (lexer *Lexer) reset() {
	for _, m := range lexer.matchers {
		m.matcher.Reset()
	}
}
