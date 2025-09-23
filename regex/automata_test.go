package regex

import (
	"fmt"
	"testing"
)

func TestDfaMinimization(t *testing.T) {
	r := NewRegex("fee|fie")
	fmt.Println(r.Dfa.GraphViz("fee|fie"))

	minDfa := r.Dfa.minimize()
	fmt.Println(minDfa.GraphViz("min fee|fie"))
}

func TestDfaMinimization2(t *testing.T) {
	r := NewRegex("a(b|c)*")
	fmt.Println(r.Dfa.GraphViz("a(b|c)*"))

	minDfa := r.Dfa.minimize()
	fmt.Println(minDfa.GraphViz("min a(b|c)*"))
}

func TestDfaMinimization3(t *testing.T) {
	r := NewRegex("(a(b|c)*){10,15}")
	fmt.Println(r.Dfa.GraphViz("(a(b|c)*){10,15}"))

	minDfa := r.Dfa.minimize()
	fmt.Println(minDfa.GraphViz("min (a(b|c)*){10,15}"))
}
