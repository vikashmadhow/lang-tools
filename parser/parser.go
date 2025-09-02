package parser

import (
	"io"

	"github.com/vikashmadhow/lang-tools/grammar"
)

type (
	Parser interface {
		Parse(input io.Reader) (*Tree[grammar.Element], error)
		ParseText(input string) (*Tree[grammar.Element], error)
	}
)
