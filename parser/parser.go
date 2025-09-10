package parser

import (
	"io"

	"github.com/vikashmadhow/lang-tools/grammar"
)

type (
	Parser interface {
		Parse(input io.Reader) (*grammar.Tree, error)
		ParseText(input string) (*grammar.Tree, error)
	}
)
