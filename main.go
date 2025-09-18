package main

import (
	"github.com/vikashmadhow/lang-tools/regex"
)

func main() {
	var r *regex.Regex

	r = regex.NewRegex("a")
	println(r.Pattern.String())
	println(r.Dfa.GraphViz(r.Pattern.String()))

	r = regex.NewRegex("a|b")
	println(r)

	r = regex.NewRegex("ab|cd")
	println(r)

	r = regex.NewRegex("(a|b)(c|d)")
	println(r)

	r = regex.NewRegex("a|bc")
	println(r)

	r = regex.NewRegex("a*")
	println(r)

	r = regex.NewRegex("a*|(bc)*")
	println(r)

	r = regex.NewRegex("a+|(abc)*")
	println(r)

	r = regex.NewRegex("x*(fc)*")
	println(r)

	r = regex.NewRegex("ab(cd|ef)?|e(fc)*\\*[a-z0-9ABC---]+")
	println(r)

	m := r.Matcher()
	println(m.MatchNext('a'))
	println(m.MatchNext('b'))
	println(m.MatchNext('c'))
	println(m.MatchNext('d'))

	r = regex.NewRegex("ab(cd|ef)?|a(fc)*\\*[a-z0-9ABC---]+")
	println(r.String())
	println(r.Dfa.GraphViz(r.String()))

	m = r.Matcher()
	println(m.MatchNext('a'))
	println(m.MatchNext('b'))
	println(m.MatchNext('c'))
	println(m.MatchNext('d'))

	//s := "abcd\xbd\xb2=\xbc\u2318日本語"
	//fmt.Println(s)
	//for i, c := range s {
	//	println(string(c), " is at position ", strconv.Itoa(i))
	//}
	//fmt.Println("---")
	//
	//ru := []rune(s)
	//for i, c := range ru {
	//	println(string(c), " is at position ", strconv.Itoa(i))
	//}
}
