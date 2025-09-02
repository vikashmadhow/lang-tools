# Language tools

This repository contains various tools for working with formal languages. Currently, it 
contains:

1. A regular expression implementation that can do partial prefix matching whereby it is 
   supplied a string to match one character at a time and responds if the supplied prefix 
   is a partial match, complete match or a failed match.
2. The regular expression engine can also generate random strings that match the structure
   of a regular expression pattern. This can be used to generate test data, or anonymize
   data while keeping their structure.
3. A fast lexer that uses the regular expression engine.
4. An LL(1) parser implementation for languages defined by a set of CFG rules, defined in Golang.

More tools will be added in the future.

## Regular expression engine
The regular expression engine currently supports the following patterns:

| Expression  | Meaning                                                                                                                                  |
|-------------|------------------------------------------------------------------------------------------------------------------------------------------|
| `x*`        | Zero or more of `x`.                                                                                                                     |
| `x+`        | One or more of `x`.                                                                                                                      |
| `x?`        | Zero or one of `x`.                                                                                                                      |
| `x{m,n}`    | `k` of `x` where `m <= k <= n`. If `m` is not provided it is set to 0. If `n` is not provided it is set to infinity.                     |
| `x{m}`      | Same as `x`{m,m}                                                                                                                         |
| `x \| y`    | `x` or `y`.                                                                                                                              |
| `(x)`       | `x` as a numbered capturing group, starting from 1. Group 0 is reserved for the whole expression. Precedence is also overridden by `()`. |

### Character and character classes
| Expression | Meaning                                                                     |
|------------|-----------------------------------------------------------------------------|
| `[a-z@#]`  | Character set: `a` to `z`, `@` and `#`. Matches any of these characters.    |
| `[^a-z@#]` | Inverse of character set: any character other than `a` to `z`, `@` and `#`. |
| `.`        | Matches any character.                                                      |
| `\d`       | Digits `[0-9]`.                                                             |
| `\D`       | Not digits `[^0-9]`.                                                        |
| `\s`       | Whitespace `[ \t\n\f\r]`.                                                   |
| `\S`       | Not whitespace `[^ \t\n\f\r]`.                                              |
| `\w`       | Word characters `[0-9a-zA-Z_]`.                                             |
| `\W`       | Not word characters `[^0-9a-zA-Z_]`.                                        |

### Modifiers
Modifiers control the behavior of regular expression matching and text generation. They are 
specified at any point in the regular expression but apply globally, which is why they are 
usually specified at the start of the expression.

| Expression | Meaning                                                                                                                                                                                                              |
|------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `(?i)`     | Case-insensitive mode: matching ignore case and random generation will include both cases.                                                                                                                           |
| `(?u)`     | Unicode mode: random generation includes Unicode characters. Matching always include Unicode characters. Without this By default, random generation is limited to printable ASCII characters (ASCII code 32 to 126). |

### Boundary matching (not implemented yet)
Boundary patterns match the start or end of a string or words.

| Expression | Meaning                                                                                                                                                        |
|------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `\b`       | Matches position between a "word character" and a "non-word character," or at the beginning/end of the string if the first/last character is a word character. |
| `\B`       | Matches any position that is not a word boundary. This is useful for finding patterns that are part of a larger word.                                          |
| `^`        | Start of string or line.                                                                                                                                       |
| `$`        | End of string or line.                                                                                                                                         |

## Lexer
The lexer is implemented using the regular expression engine
