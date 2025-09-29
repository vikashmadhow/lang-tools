package grammar

import (
	"slices"
	"strconv"
	"strings"

	"github.com/vikashmadhow/lang-tools/seq"
)

type (
	Tree struct {
		Node     Element
		Children []*Tree
	}

	TreeMap map[string][]*Tree

	TreeParent map[*Tree]*Tree

	TreeIndex struct {
		Tree    *Tree
		TreeMap TreeMap
		Parent  TreeParent
	}

	Path = *seq.List[*Tree]

	/*
	   <  : child
	   << : descendent
	   -  : sibling

	   e <  e' << ID - PLUS - ID
	   e << f - d << x < ID
	*/

	Map func(tree *Tree, pathFromRoot Path) *Tree

	MatchMap func(*Tree, *MatchResult) *Tree

	MatchResult struct {
		Tree     *Tree
		Paths    []Path
		Bindings map[string][]*Tree
	}

	MatchPattern [][]string
)

const (
	PARENT   = "<"
	ANCESTOR = "<<"
	SIBLING  = "-"
)

func NewParseTree(node Element) *Tree {
	return &Tree{
		node,
		[]*Tree{},
	}
}

func (tree *Tree) Copy() *Tree {
	return &Tree{tree.Node, slices.Clone(tree.Children)}
}

func (tree *Tree) GraphViz(title string) string {
	spec := "digraph G {\n"
	if len(title) > 0 {
		spec += "\tlabel=\"" + title + "\"\n"
	}
	spec += tree.graphVizNode("0")
	spec += "}"
	return spec
}

func (tree *Tree) graphVizNode(position string) string {
	spec := ""
	for i, c := range tree.Children {
		if c != nil {
			pos := position + strconv.Itoa(i)
			spec += "\t\"" + tree.Node.String() + " [" + position + "]" +
				"\" -> \"" + c.Node.String() + " [" + pos + "]" + "\"\n"
			spec += c.graphVizNode(pos)
		}
	}
	return spec
}

// Map maps the tree to another tree where modified nodes (and all their
// ancestors) are copied and non-modified nodes are shared between this tree
// and the mapped tree. This method allows for the trees to be immutable
// with all mutating operations creating a new persistent tree.
func (tree *Tree) Map(mapper Map) *Tree {
	return tree.MapWithPath(mapper, seq.NewList(tree, nil))
}

func (tree *Tree) MapWithPath(mapper Map, path Path) *Tree {
	var cp *Tree
	for i, child := range tree.Children {
		if child != nil {
			mappedChild := child.MapWithPath(mapper, path.Push(child))
			if mappedChild != child {
				if cp == nil {
					cp = tree.Copy()
					path = path.Replace(cp) // replace head of path with its copy
				}
				cp.Children[i] = mappedChild
			}
		}
	}
	if cp == nil {
		return mapper(tree, path)
	} else {
		return mapper(cp, path)
	}
}

func ParsePattern(pattern string) MatchPattern {
	var pat MatchPattern
	for _, line := range strings.Split(pattern, "\n") {
		pat = append(pat, strings.Split(line, " "))
	}
	return pat
}

func (tree *Tree) MapMatch(mapper MatchMap, pattern MatchPattern) *Tree {
	// build tree index
	index := tree.BuildIndex()

	// find matches
	result := index.Match(pattern)
	resultMap := make(map[*Tree]*MatchResult)
	for _, r := range result {
		resultMap[r.Tree] = r
	}

	// map matches
	if result != nil {
		return tree.Map(func(t *Tree, path Path) *Tree {
			if resultMap[t] != nil {
				return mapper(t, resultMap[t])
			} else {
				return t
			}
		})
	} else {
		return tree
	}
}

func Compact(tree *Tree, _ Path) *Tree {
	nonNil := tree.nonNilChildren()
	if slices.Equal(nonNil, tree.Children) {
		return tree
	} else {
		return &Tree{tree.Node, nonNil}
	}
}

func PromoteSingleChild(tree *Tree, _ Path) *Tree {
	nonNil := tree.nonNilChildren()
	if len(nonNil) == 1 {
		return nonNil[0]
	}
	return tree
}

func DropOrphanNonTerminal(tree *Tree, _ Path) *Tree {
	if !tree.Node.Terminal() && len(tree.nonNilChildren()) == 0 {
		return nil
	} else {
		return tree
	}
}

func Compose(mapper ...Map) Map {
	return func(tree *Tree, path Path) *Tree {
		for _, mapper := range mapper {
			tree = mapper(tree, path)
			if tree == nil {
				break
			}
		}
		return tree
	}
}

var CleanUp = Compose(DropOrphanNonTerminal, PromoteSingleChild, Compact)

func (tree *Tree) nonNilChildren() []*Tree {
	changed := false
	var newChildren []*Tree
	for _, child := range tree.Children {
		if child == nil {
			changed = true
		} else {
			newChildren = append(newChildren, child)
		}
	}
	if changed {
		return newChildren
	} else {
		return tree.Children
	}
}

func (index *TreeIndex) Match(patterns [][]string) []*MatchResult {
	var tops map[*Tree][]Path
	for _, pattern := range patterns {
		previousRel := ""
		bottoms := make(map[*Tree]Path)

		// Ensures that we don't match the same sibling twice in, e.g., ID - PLUS - ID.
		// We don't want to match the same ID for both 'ID' criteria in the pattern.
		//
		// This also prevents the same tree from being matched twice. For example,
		// searching bottom-up for the pattern "ID - PLUS - ID", we'll match two
		// nodes with the ID pattern, but both could be part of the same sub-tree:
		//      EXPR
		//     / | \
		//   ID PLUS ID
		//
		// Initially, we'll have 2 matches for the ID pattern, one for the left and
		// another for the right. When we then apply the sibling operator on the right
		// ID we get a match with the PLUS. We then don't want the right ID to be
		// matched with the same PLUS pattern. This map prevents that from by having
		// already recorded the PLUS as having matched the first time and preventing
		// further matches with it. The second ID is then dropped from potential matches.
		matched := make(map[*Tree]bool)

		for j := len(pattern) - 1; j >= 0; j-- {
			rel := pattern[j]

			if j == len(pattern)-1 {
				for _, t := range index.TreeMap[rel] {
					bottoms[t] = seq.NewList(t, nil)
					matched[t] = true
				}
			} else {
				if previousRel == "" {
					if rel != PARENT && rel != ANCESTOR && rel != SIBLING {
						panic("bad relation pattern " + rel)
					}
					previousRel = rel
				} else {
					// match from bottoms
					newBottoms := make(map[*Tree]Path)
					switch previousRel {
					case SIBLING:
						for t, path := range bottoms {
							parent := index.Parent[t]
							if parent != nil {
								for k := len(parent.Children) - 1; k >= 0; k-- {
									child := parent.Children[k]
									if !matched[child] && child.Node.TreePattern() == rel {
										newBottoms[child] = path.Push(child)
										matched[child] = true
										break
									}
								}
							}
						}
					case PARENT:
						for t, path := range bottoms {
							parent := index.Parent[t]
							if parent != nil {
								if !matched[parent] && parent.Node.TreePattern() == rel {
									newBottoms[parent] = path.Push(parent)
									matched[parent] = true
								}
							}
						}
					case ANCESTOR:
						for t, path := range bottoms {
							p := index.Parent[t]
							for ; p != nil && p.Node.TreePattern() != rel; p = index.Parent[p] {
							}
							if p != nil {
								newBottoms[p] = path.Push(p)
								matched[p] = true
							}
						}
					default:
						panic("bad relation pattern " + previousRel)
					}
					bottoms = newBottoms
					previousRel = ""
				}
			}
			if len(bottoms) == 0 {
				// no match
				return nil
			}
		}
		if tops == nil {
			tops = make(map[*Tree][]Path)
			for t, path := range bottoms {
				tops[t] = []Path{path}
			}
		} else {
			newTops := make(map[*Tree][]Path)
			for t, path := range bottoms {
				if tops[t] != nil {
					newTops[t] = append(tops[t], path)
				}
			}
			tops = newTops
		}
		if len(tops) == 0 {
			return nil
		}
	}
	var result []*MatchResult
	for t, paths := range tops {
		result = append(result, newMatchResult(t, paths))
	}
	return result
}

func newMatchResult(tree *Tree, paths []Path) *MatchResult {
	bindings := make(map[string][]*Tree)
	for _, path := range paths {
		for t := range path.Iter() {
			pattern := t.Node.TreePattern()
			bindings[pattern] = append(bindings[pattern], t)
		}
	}
	return &MatchResult{
		tree,
		paths,
		bindings,
	}
}

func (tree *Tree) BuildIndex() *TreeIndex {
	treeMap := TreeMap{}
	parent := TreeParent{}

	tree.Map(func(t *Tree, path Path) *Tree {
		treeMap[t.Node.TreePattern()] = append(treeMap[t.Node.TreePattern()], t)
		if path.Tail != nil {
			parent[t] = path.Tail.Head
		}
		return t
	})

	return &TreeIndex{
		tree,
		treeMap,
		parent,
	}
}
