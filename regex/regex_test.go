// author: Vikash Madhow (vikash.madhow@gmail.com)

package regex

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestX(t *testing.T) {
	r := NewRegex("x(ab(vw(cd)|(ef))?)|(a(fc)*\\*[a-z0-9ABC---]+)")
	println(r.Pattern.nfa().GraphViz("x(ab(vw(cd)|(ef))?)|(a(fc)*\\*[a-z0-9ABC---]+)"))
	println(r.Dfa.GraphViz("x(ab(vw(cd)|(ef))?)|(a(fc)*\\*[a-z0-9ABC---]+)"))
	//if !r.Match("") {
	//	t.Error("'' did not match ''")
	//}
	//if r.Match("a") {
	//	t.Error("'' matched 'a'")
	//}
}

func TestMultipleRange(t *testing.T) {
	var re strings.Builder
	re.WriteRune('[')
	var s strings.Builder
	for i := 10000; i > 10; i -= 4 {
		//start, end := rand.Intn(utf8.MaxRune), rand.Intn(utf8.MaxRune)
		//if start > end {
		//	start, end = end, start
		//}
		//re.WriteRune(rune(start))
		//re.WriteRune('-')
		//re.WriteRune(rune(end))

		re.WriteRune(rune(i - 2))
		re.WriteRune('-')
		re.WriteRune(rune(i))

		//s.WriteRune(rune(start + rand.Intn(end-start+1)))
		s.WriteRune(rune(i - 1))
	}
	re.WriteString("]*")
	regex := NewRegex(re.String())
	text := s.String()
	for i := 0; i < 1000; i++ {
		if !regex.Match(text) {
			t.Error("Multiple range did not match")
		}
	}
}

func TestMultipleRangeInternal(t *testing.T) {
	var re strings.Builder
	re.WriteRune('[')
	var s strings.Builder
	for i := 10000; i > 10; i -= 4 {
		re.WriteString("\\x{" + strconv.FormatInt(int64(i-2), 16) + "}")
		re.WriteRune('-')
		re.WriteString("\\x{" + strconv.FormatInt(int64(i), 16) + "}")
		s.WriteRune(rune(i - 1))
	}
	re.WriteString("]*")
	regex, e := regexp.Compile(re.String())
	if e != nil {
		println(e.Error())
		return
	}
	text := s.String()
	for i := 0; i < 1000; i++ {
		if !regex.MatchString(text) {
			t.Error("Multiple range did not match")
		}
	}
}

func Test2(t *testing.T) {
	r := NewRegex("(ab)|(ac)")
	println(r.Pattern.nfa().GraphViz("(ab)|(ac)"))
	println(r.Dfa.GraphViz("(ab)|(ac)"))

	//if !r.Match("") {
	//	t.Error("'' did not match ''")
	//}
	//if r.Match("a") {
	//	t.Error("'' matched 'a'")
	//}
}

func TestCaptureGroup(t *testing.T) {
	r := NewRegex("(aab)|(aac)")
	m := r.Matcher()
	println(r.Dfa.GraphViz("(aab)|(aac)"))

	m.Match("aab")
	for k, v := range m.Groups {
		println(k, v)
	}

	m.Reset()
	m.Match("aac")
	for k, v := range m.Groups {
		println(k, v)
	}

	//if !r.Match("") {
	//	t.Error("'' did not match ''")
	//}
	//if r.Match("a") {
	//	t.Error("'' matched 'a'")
	//}
}

func TestEmpty(t *testing.T) {
	r := NewRegex("")
	//fmt.Println(r.Dfa.GraphViz(""))
	//fmt.Println(r.Dfa.minimize().GraphViz("min"))
	if !r.Match("") {
		t.Error("'' did not match ''")
	}
	if r.Match("a") {
		t.Error("'' matched 'a'")
	}
	if !r.MatchEmpty() {
		t.Error("'' MatchEmpty returns false")
	}
}

func TestSingleChar(t *testing.T) {
	r := NewRegex("a")
	if !r.Match("a") {
		t.Error("'a' did not match 'a'")
	}
	if r.Match("b") {
		t.Error("'a' matched 'b'")
	}
	if r.Match("aa") {
		t.Error("'a' matched 'aa'")
	}
	if r.Match("") {
		t.Error("'a' matched ''")
	}
	if r.MatchEmpty() {
		t.Error("'a' MatchEmpty returns true")
	}
}

func TestSequence(t *testing.T) {
	r := NewRegex("abc")
	if !r.Match("abc") {
		t.Error("'abc' did not match 'abc'")
	}
	if r.Match("ab") {
		t.Error("'abc' matched 'ab'")
	}
	if r.Match("abcabc") {
		t.Error("'abc' matched 'abcabc'")
	}
	if r.Match("") {
		t.Error("'abc' matched ''")
	}
}

func TestChoice(t *testing.T) {
	r := NewRegex("a|b")
	if !r.Match("a") {
		t.Error("'a|b' did not match 'a'")
	}
	if !r.Match("b") {
		t.Error("'a|b' did not match 'b'")
	}
	if r.Match("ab") {
		t.Error("'a|b' matched 'ab'")
	}
	if r.MatchEmpty() {
		t.Error("'a|b' MatchEmpty returns true")
	}
}

func TestSequenceChoice(t *testing.T) {
	r := NewRegex("ab|ac")
	if !r.Match("ab") {
		t.Error("'ab|ac' did not match 'ab'")
	}
	if !r.Match("ac") {
		t.Error("'ab|ac' did not match 'ac'")
	}
	if r.Match("abac") {
		t.Error("'ab|ac' matched 'abac'")
	}
}

func TestOpt(t *testing.T) {
	r := NewRegex("a?")
	if !r.Match("a") {
		t.Error("'a?' did not match 'a'")
	}
	if !r.Match("") {
		t.Error("'a?' did not match ''")
	}
	if r.Match("aa") {
		t.Error("'a?' matched 'aa'")
	}
}

func TestSequenceOpt(t *testing.T) {
	r := NewRegex("(ab)?")
	if !r.Match("ab") {
		t.Error("'(ab)?' did not match 'ab'")
	}
	if !r.Match("") {
		t.Error("'(ab)?' did not match ''")
	}
	if r.Match("abab") {
		t.Error("'(ab)?' matched 'abab'")
	}
}

func TestSequenceOpt2(t *testing.T) {
	r := NewRegex("ab?")
	if !r.Match("ab") {
		t.Error("'ab?' did not match 'ab'")
	}
	if !r.Match("a") {
		t.Error("'ab?' did not match 'a'")
	}
	if r.Match("") {
		t.Error("'ab?' matched ''")
	}
	if r.Match("abab") {
		t.Error("'(ab)?' matched 'abab'")
	}
}

func TestZeroOrMore(t *testing.T) {
	r := NewRegex("a*")
	if !r.Match("a") {
		t.Error("'a*' did not match 'a'")
	}
	if !r.Match("aa") {
		t.Error("'a*' did not match 'aa'")
	}
	if !r.Match("aaa") {
		t.Error("'a*' did not match 'aaa'")
	}
	if !r.Match("") {
		t.Error("'a*' did not match ''")
	}
}

func TestZeroOrMoreSequence(t *testing.T) {
	r := NewRegex("(ab)*")
	if !r.Match("ab") {
		t.Error("'(ab)*' did not match 'ab'")
	}
	if !r.Match("abab") {
		t.Error("'(ab)*' did not match 'abab'")
	}
	if !r.Match("ababab") {
		t.Error("'(ab)*' did not match 'ababab'")
	}
	if !r.Match("") {
		t.Error("'(ab)*' did not match ''")
	}
	if r.Match("a") {
		t.Error("'(ab)*' matched 'a'")
	}
	if r.Match("b") {
		t.Error("'(ab)*' matched 'b'")
	}
	if r.Match("aba") {
		t.Error("'(ab)*' matched 'aba'")
	}
}

func TestOneOrMore(t *testing.T) {
	r := NewRegex("a+")
	if !r.Match("a") {
		t.Error("'a+' did not match 'a'")
	}
	if !r.Match("aa") {
		t.Error("'a+' did not match 'aa'")
	}
	if !r.Match("aaa") {
		t.Error("'a+' did not match 'aaa'")
	}
	if r.Match("") {
		t.Error("'a+' matched ''")
	}
}

func TestOneOrMoreSequence(t *testing.T) {
	r := NewRegex("(ab)+")
	if !r.Match("ab") {
		t.Error("'(ab)+' did not match 'ab'")
	}
	if !r.Match("abab") {
		t.Error("'(ab)+' did not match 'abab'")
	}
	if !r.Match("ababab") {
		t.Error("'(ab)+' did not match 'ababab'")
	}
	if r.Match("") {
		t.Error("'(ab)+' matched ''")
	}
	if r.Match("a") {
		t.Error("'(ab)+' matched 'a'")
	}
	if r.Match("b") {
		t.Error("'(ab)+' matched 'b'")
	}
	if r.Match("aba") {
		t.Error("'(ab)+' matched 'aba'")
	}
}

func TestRepeat(t *testing.T) {
	r := NewRegex("(ab|ac){5,3}")
	if !r.Match("abacab") {
		t.Error("'(ab|ac){3,5}' did not match 'abacab'")
	}
	if !r.Match("abacabab") {
		t.Error("'(ab|ac){3,5}' did not match 'abacabab'")
	}
	if !r.Match("abacababac") {
		t.Error("'(ab|ac){3,5}' did not match 'abacababac'")
	}
	if r.Match("abacababacab") {
		t.Error("'(ab|ac){3,5}' matched 'abacababacab'")
	}
	if r.Match("") {
		t.Error("'(ab|ac){3,5}' matched ''")
	}
}

func TestRepeatExact(t *testing.T) {
	r := NewRegex("(ab|ac){3}")
	if !r.Match("abacab") {
		t.Error("'(ab|ac){3}' did not match 'abacab'")
	}
	if r.Match("abacabab") {
		t.Error("'(ab|ac){3}' matched 'abacabab'")
	}
	if r.Match("abacababac") {
		t.Error("'(ab|ac){3}' matched 'abacababac'")
	}
	if r.Match("abacababacab") {
		t.Error("'(ab|ac){3}' matched 'abacababacab'")
	}
	if r.Match("") {
		t.Error("'(ab|ac){3}' matched ''")
	}
}

func TestRepeatNoUpperLimit(t *testing.T) {
	r := NewRegex("(ab|ac){3,}")
	if !r.Match("abacab") {
		t.Error("'(ab|ac){3,}' did not match 'abacab'")
	}
	if !r.Match("abacabab") {
		t.Error("'(ab|ac){3,}' did not match 'abacabab'")
	}
	if !r.Match("abacababac") {
		t.Error("'(ab|ac){3,}' did not match 'abacababac'")
	}
	if !r.Match("abacababacab") {
		t.Error("'(ab|ac){3,}' did not match 'abacababacab'")
	}
	if r.Match("") {
		t.Error("'(ab|ac){3,}' matched ''")
	}
}

func TestRepeatNoLowerLimit(t *testing.T) {
	r := NewRegex("(ab|ac){,3}")
	if !r.Match("") {
		t.Error("'(ab|ac){,3}' did not match ''")
	}
	if !r.Match("abac") {
		t.Error("'(ab|ac){,3}' did not match 'abac'")
	}
	if !r.Match("abacab") {
		t.Error("'(ab|ac){,3}' did not match 'abacab'")
	}
	if r.Match("abacabab") {
		t.Error("'(ab|ac){,3}' matched 'abacabab'")
	}
	if r.Match("abacababac") {
		t.Error("'(ab|ac){,3}' matched 'abacababac'")
	}
}

func TestDigits(t *testing.T) {
	r := NewRegex("\\d{3,5}")
	if !r.Match("569") {
		t.Error("'\\d{3,5}' did not match '569'")
	}
	if !r.Match("5697") {
		t.Error("'\\d{3,5}' did not match '5697'")
	}
	if !r.Match("56975") {
		t.Error("'\\d{3,5}' did not match '56975'")
	}
	if r.Match("569751") {
		t.Error("'\\d{3,5}' matched '569751'")
	}
	if r.Match("5bc") {
		t.Error("'\\d{3,5}' matched '5bc'")
	}
}

func TestNonDigits(t *testing.T) {
	r := NewRegex("\\D{3,5}")
	if !r.Match("abF") {
		t.Error("'\\D{3,5}' did not match 'abF'")
	}
	if !r.Match("abFs") {
		t.Error("'\\D{3,5}' did not match 'abFs'")
	}
	if !r.Match("abFs?") {
		t.Error("'\\D{3,5}' did not match 'abFs?'")
	}
	if r.Match("abFs?;") {
		t.Error("'\\D{3,5}' matched 'abFs?;'")
	}
	if r.Match("5bc") {
		t.Error("'\\D{3,5}' matched '5bc'")
	}
}

func TestSpaces(t *testing.T) {
	r := NewRegex("\\s{3,5}")
	if !r.Match(" 	 ") {
		t.Error("'\\s{3,5}' did not match ' 	 '")
	}
	if !r.Match("  		") {
		t.Error("'\\s{3,5}' did not match '  		'")
	}
	if !r.Match("  		 ") {
		t.Error("'\\s{3,5}' did not match '  		 '")
	}
	if r.Match("  		  ") {
		t.Error("'\\s{3,5}' matched '  		  '")
	}
	if r.Match("5  ") {
		t.Error("'\\s{3,5}' matched '5  '")
	}
}

func TestNonSpaces(t *testing.T) {
	r := NewRegex("\\S{3,5}")
	if !r.Match("abc") {
		t.Error("'\\S{3,5}' did not match 'abc'")
	}
	if !r.Match("abcd") {
		t.Error("'\\S{3,5}' did not match 'abcd'")
	}
	if !r.Match("abcde") {
		t.Error("'\\S{3,5}' did not match 'abcde'")
	}
	if r.Match("abcdef") {
		t.Error("'\\S{3,5}' matched 'abcdef'")
	}
	if r.Match("   ") {
		t.Error("'\\S{3,5}' matched '   '")
	}
}

func TestWords(t *testing.T) {
	r := NewRegex("\\w{3,5}")
	if !r.Match("ab0") {
		t.Error("'\\w{3,5}' did not match 'ab0'")
	}
	if !r.Match("ab01") {
		t.Error("'\\w{3,5}' did not match 'ab01'")
	}
	if !r.Match("abc01") {
		t.Error("'\\w{3,5}' did not match 'abc01'")
	}
	if r.Match("abc012") {
		t.Error("'\\w{3,5}' matched 'abc012'")
	}
	if r.Match("?bc") {
		t.Error("'\\w{3,5}' matched '?bc'")
	}
}

func TestNonWords(t *testing.T) {
	r := NewRegex("\\W{3,5}")
	if !r.Match("<>?") {
		t.Error("'\\W{3,5}' did not match '<>?'")
	}
	if !r.Match("<>?,") {
		t.Error("'\\W{3,5}' did not match '<>?,'")
	}
	if !r.Match("<>?,.") {
		t.Error("'\\W{3,5}' did not match '<>?,.'")
	}
	if r.Match("<>?,./") {
		t.Error("'\\W{3,5}' matched '<>?,./'")
	}
	if r.Match("A<>") {
		t.Error("'\\W{3,5}' matched 'A<>'")
	}
}

func TestDot(t *testing.T) {
	r := NewRegex(".{3,5}")
	if !r.Match("^*k") {
		t.Error("'.{3,5}' did not match '^*k'")
	}
	if !r.Match("^*k)") {
		t.Error("'.{3,5}' did not match '^*k'")
	}
	if !r.Match("^*k)$") {
		t.Error("'.{3,5}' did not match '^*k)$'")
	}
	if r.Match("^*k)$d") {
		t.Error("'.{3,5}' matched '^*k)$d'")
	}
	if r.Match("") {
		t.Error("'.{3,5}' matched ''")
	}
}

func TestUnicode(t *testing.T) {
	r := NewRegex(".{3,3}")
	if !r.Match("日本語") {
		t.Error("'.{3,3}' did not match '日本語'")
	}
}
