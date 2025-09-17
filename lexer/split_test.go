// author: Vikash Madhow (vikash.madhow@gmail.com)

package lexer

import (
	"fmt"
	"slices"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		patterns []string
		expected []string
	}{
		{"a+b", []string{"\\+"}, []string{"a", "+", "b"}},
		{"123 +  456 -5+4", []string{`\s*\+\s*`, `\s*-\s*`}, []string{"123", " +  ", "456", " -", "5", "+", "4"}},
		//{"x = y", []string{"\\+"}, []string{"x", "=", "y"}},
		//{" a + b ", []string{"\\+"}, []string{"a", "+", "b"}},
		//{"\"hello\"", []string{"\\+"}, []string{"\"hello\""}},
		//{"1.23", []string{"\\+"}, []string{"1.23"}},
		//{"a++b", []string{"\\+"}, []string{"a", "++", "b"}},
		//{"a = b + c", []string{"\\+"}, []string{"a", "=", "b", "+", "c"}},
	}

	for _, test := range tests {
		var result []string
		for s := range Tokenize(test.input, test.patterns...) {
			fmt.Println("-" + s + "-")
			result = append(result, s)
		}
		if !slices.Equal(result, test.expected) {
			t.Errorf("For input %q: expected %v, got %v", test.input, test.expected, result)
		}
	}
}
