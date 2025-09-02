package regex

import (
	"testing"
)

func TestGenerator_Next1(t *testing.T) {
	r := NewRegex("a*")
	for i := 0; i < 10; i++ {
		println(r.Generate())
	}
}

func TestGenerator_Next2(t *testing.T) {
	r := NewRegex("a+")
	for i := 0; i < 10; i++ {
		println(r.Generate())
	}
}

func TestGenerator_Next3(t *testing.T) {
	r := NewRegex("([A-Z][0-9]){3}|([A-Z][0-9][A-Z] [0-9][A-Z][0-9])")
	for i := 0; i < 10; i++ {
		println(r.Generate())
	}
}

func TestGenerator_Next4(t *testing.T) {
	r := NewRegex("\\d{4}-[A-Z]+")
	for i := 0; i < 10; i++ {
		println(r.Generate())
	}
}

func TestGenerator_Next5(t *testing.T) {
	r := NewRegex("(?i)[A-Z]{5}-[^ A-Za-z!#$#$#$#]{20}")
	for i := 0; i < 10; i++ {
		println(r.Generate())
	}
}

func TestGenerator_Next6(t *testing.T) {
	r := NewRegex("(20[012]|19[7-9])\\d-[0-9]{5}(-[1-4])?")
	for i := 0; i < 1000; i++ {
		println(r.Generate())
	}
}

func TestGenerator_EmailGeneration(t *testing.T) {
	r := NewRegex("(:word_en:l)(\\.(:word_en:ls))?@(:word_en:l)(\\.(:word_en:l))?\\.(ca|com|net|org|edu)")
	for i := 0; i < 1000; i++ {
		println(r.Generate())
	}
}

func TestGenerator_WordGeneration(t *testing.T) {
	r := NewRegex("(:word_fr:l)( (:word_fr)){1,10}")
	for i := 0; i < 1000; i++ {
		println(r.Generate())
	}
}
