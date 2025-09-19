# Changelog

## [0.5.3] - 2025-09-18
- Fixed the binary search matching for character sets where the iteration could be
  trapped in an infinite loop.
- `finalMap` added to `Dfa` to make final state check faster.
- `Groups` in `Matcher` now uses a `strings.Builder` instead of simple string concatenation
  resulting in more that 100x speedup for complex regex.
- `Nfa` dropped from `Regex` as it can be obtained through the `Pattern` linked to 
  the `Regex`.
- Slight optimization of linear matching for character sets, where the matching is stopped
  as soon as a span that is greater than the current character is found.

## [0.5.2] - 2025-09-18
- Renamed `Regex` to `Pattern` and `CompiledRegex` to `Regex`, as the latter is
  the public interface for regular expressions. `Pattern` are the uncompiled
  regular expression which is only used when defining new regex.
- Character set matching now uses a binary search over the spans of the individual
  `char` comprising the set, if there are more than 5 of them. This result in a
  60% speedup for matching a large set of characters.
- Refactoring and cleanup of regex to simplify interface.

## [0.5.1] - 2025-09-17
- Fixed 2 issues with lexer:
  - Incorrect calculation and removal of overlapping text from unmatched tokens. 
  - Lexer was being reset after emitting an unmatched token, which caused it to
    miss the starting part of the next valid token.

## [0.5.0] - 2025-09-17
- New lexer implementation that can be used for tokenization as well as lexing.
  - Tokenization and splitting support.
  - Fixes longest-string identification for patterns which an optional suffix
    (e.g. xyz(ab)*)
  - Read buffer is dynamically increased when there are not enough characters
    for the lexer to make a decision on emitting a token.
- `Matcher` keeps full and partial matches.
- Special token types now use code points in the Unicode PUA to reduce possibility
  of clash with regular characters.
- Tree matching cleanup.

## [0.4.7] - 2025-09-10
- Test minimal LL(1) language extended with function definitions, function calls, 
  variable declarations, and comparison operators.

## [0.4.6] - 2025-09-09
- Subtrees can be searched by specifying a set of patterns consisting of language
  elements (tokens-types and non-terminals) and their relationships: PARENT, ANCESTOR, 
  or SIBLING. E.g., `e << f - d < a << x` search for a subtree that has a root of `e`
  which is an ancestor of `f`, which has a sibling `d` which is, in turn, a parent 
  of `a` and, finally, which is an ancestor of `x`.
- Matched subtrees can be transformed by applying a mapping function to them.
- Tree.Map function provides a functionally persistent path from the root of the tree
  to the node being mapped. This path is useful for more complex tree transformations.
- Fully functional LL(1) parser.
- An LL(1) micro-language for testing.
- Detection of left-recursion when constructing LL(1) parsing table.

## [0.4.5] - 2025-09-02
- Simplify interface and method signatures.
- Parser and lexer refactored into their own packages.
- `Element` interface now embeds `fmt.Stringer`, which requires that all language
  elements implement it.
- LL(1) Parser now produces a proper parse tree.
- `Tree` struct for representing a parse tree, and, later, syntax trees.
- Tree mapping functions allow for tree transformation in a functional persistent manner
  (preserving immutability of the tree).
- Useful included tree mappers:
  - `DropOrphanNonTerminal`: drops any leaf of a non-terminal without any children.
    These leaves typically map a non-terminal to the empty string, thus contributing
    nothing to the parse tree.
  - `PromoteSingleChild`: promotes a single child of a parent to be the parent. This
    removes long single branches from the tree.
  - `Compact`: remove `nil` children from a branch.

## [0.4.4] - 2025-09-02
- Code cleanup and restructuring.

## [0.4.3] - 2025-09-01
- Definition of a grammar with simple rule syntax encompassing the definition of
  both lexer tokens and grammar productions.
- Working implementation of an LL(1) parser.

## [0.4.2] - 2025-08-26
- Restructuring of token, lexer, and grammar packages in preparation for:
  - support for more regular expression syntax, including boundary matchers.
  - implementation of a LL(1) parser.

## [0.4.1] - 2025-08-23
- Conversion modifiers in `(:list)` as an optional suffix (`(:list:lumts)`)
  - `l`: converts the generated word to lowercase.
  - `u`: converts the generated word to uppercase.
  - `t`: converts the generated word to titlecase (first character to uppercase and rest to lowercase).
  - `m`: trims the generated word of leading and trailing whitespace.
  - `s`: replaces all whitespace characters with a single space.

## [0.4.0] - 2025-08-23
- `(:list)` syntax in regular expression for generating random words from the given list.
- Embedded sanitized list of english and french words (word_en, word_fr).

## [0.3.3] - 2025-08-21
- Improved matching code in span, spanSet, and charSet.

## [0.3.2] - 2025-08-20
- Fixed all unit tests.
- Cleaned up code.

## [0.3.1] - 2025-08-20
- Support for modifiers through the (?) syntax.
  - (?i) turns on case-insensitive mode, whereby matches ignore case and random
    generation will include both cases.
  - (?u) turns on Unicode mode, whereby random generation will include Unicode
    characters. This modifier has no effect on matching which is always on Unicode
    characters. By default, random generation is limited to printable ASCII characters
    (ASCII code 32 to 126).

## [0.3.0] - 2025-08-19
- regex.Generate() method for generating a random string matching the regex.
- Support for random string generation as span and spanSet classes representing
  a range of characters, or a set of ranges, respectively, together with various
  operations to merge, invert, subtract, and intersect them.

## [0.2.5] - 2025-08-14
- Basic parser implementation.

## [0.2.4] - 2025-01-07
- `LanguageElement` can now return a `TreeRetention` value of `Retain`, `Drop` or
  `Promote`, whereby the element will be retained in the `SyntaxTree`, removed from
  it, or promoted to its parent (replacing it), respectively, by the grammar parser.
  This allows the parser to generate smaller trees that are easier to analyse and
  restructure by downstream components.

## [0.2.3] - 2025-01-06
- Using a channel for holding push-backed tokens in `TokenSeq` instead of a slice
  to improve performance.
- Graphviz representation of `SyntaxTree`.
- Mapping functions documented.
- First working grammar parser of a simple test program.

## [0.2.2] - 2025-01-01
- Improvement to type definition of `Modulator`, `filter`, `map` and `flatMap`.
- `filter`, `map` and `flatMap` can now work with type aliases of the predicate and
  mapping function.
- Flatmap fixed to close holding channel and continue to return data after first 
  mapping.
- Lexer now adds the special `EOF` token at the end of the token stream to signal end
  of input to downstream modulators and parser. This allows for the creation of 
  downstream logic that needs to operate when all the tokens have been lexed; E.g.,
  a modulator that reverses the stream of tokens.
- `Ignore` Modulator for ignoring specific tokens in the token stream (such as whitespace).
- `Reverse` example Modulator for reversing the token stream.

## [0.2.1] - 2024-12-30
- Filtering, mapping, and flat-mapping can now work on both the pull and push versions of iter.Seq
  and iter.Seq2.
- Lexer `Lex` methods for returning a push version of iter.Seq2, which is simple to iterate over.
  However, these versions do not support token pushback.

## [0.2.0] - 2024-12-30
- A set of utility functions for filtering, mapping and flat-mapping over iter.Seq and iter.Seq2.
- Lexer now reads its input from an `io.Reader` which is more memory efficient.
- Flatmap functions can be attached to a lexer to modulate its output, arbitrarily changing the
  token stream. This can be used to ignore certain tokens, modify tokens, or insert new tokens
  at arbitrary points in the token stream.
- Simple context-free-grammar definition and predictive parser (untested).

## [0.1.3] - 2024-12-02
- Lexer will not provide the token(s) which failed and the expected next character(s) for each.
- Error information is now provided if the last token in the input stream has an error,

## [0.1.2] - 2024-11-30
- Improved lexer matching, error and position tracking.

## [0.1.1] - 2024-11-27
- Line and column where each token matched reported in `Token` by `Lexer`. 

## [0.1.0] - 2024-11-25
- Numbered capturing groups with group 0 matching whole string and each opening
  parenthesis `(` starting a new capturing group. Capturing groups can be nested.

## [0.0.0-20241124] - 2024-11-24
- Character classes shortcuts (`\d`, `\w`, etc.)
- Matching any character with `.`.
- More regular expression tests.

## [0.0.0-20241123] - 2024-11-23
- Lexer and regex tests.
- `match` method in `CompiledRegex` match regular expression to string exactly.
- Range repetition in regular expression (`re{m,n}`).

## [0.0.0-20241122] - 2024-11-22
- Working DFA implementation for regular expression matching.
- Base incremental lexer working and tested.
