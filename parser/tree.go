package parser

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/vikashmadhow/lang-tools/grammar"
)

type (
	Tree[E fmt.Stringer] struct {
		Node     E
		Children []*Tree[E]
	}

	TreeMapper[E fmt.Stringer] func(*Tree[E]) *Tree[E]
)

func NewParseTree(node grammar.Element) *Tree[grammar.Element] {
	return &Tree[grammar.Element]{
		node,
		[]*Tree[grammar.Element]{},
	}
}

func (tree *Tree[E]) Copy() *Tree[E] {
	return &Tree[E]{tree.Node, slices.Clone(tree.Children)}
}

func GraphViz[T fmt.Stringer](tree *Tree[T], title string) string {
	spec := "digraph G {\n"
	if len(title) > 0 {
		spec += "\tlabel=\"" + title + "\"\n"
	}
	spec += graphVizNode(tree, "0")
	spec += "}"
	return spec
}

func graphVizNode[T fmt.Stringer](tree *Tree[T], position string) string {
	spec := ""
	for i, c := range tree.Children {
		if c != nil {
			pos := position + strconv.Itoa(i)
			spec += "\t\"" + tree.Node.String() + " [" + position + "]" +
				"\" -> \"" + c.Node.String() + " [" + pos + "]" + "\"\n"
			spec += graphVizNode(c, pos)
		}
	}
	return spec
}

// Map maps the tree to another tree where modified nodes (and all their
// ancestors) are copied and non-modified nodes are shared between this tree
// and the mapped tree. This method allows for the trees to be immutable
// with all mutating operations creating a new persistent tree.
func (tree *Tree[E]) Map(mapper TreeMapper[E]) *Tree[E] {
	var cp *Tree[E]
	for i, child := range tree.Children {
		if child != nil {
			mappedChild := child.Map(mapper)
			if mappedChild != child {
				if cp == nil {
					cp = tree.Copy()
					//path = path.replaceHead(cp)   // replace head of path with its copy
				}
				cp.Children[i] = mappedChild
			}
		}
	}
	if cp == nil {
		return mapper(tree)
	} else {
		return mapper(cp)
	}
}

func Compact[E fmt.Stringer](tree *Tree[E]) *Tree[E] {
	nonNil := nonNilChildren(tree)
	if slices.Equal(nonNil, tree.Children) {
		return tree
	} else {
		return &Tree[E]{tree.Node, nonNil}
	}
}

func PromoteSingleChild[E fmt.Stringer](tree *Tree[E]) *Tree[E] {
	nonNil := nonNilChildren(tree)
	if len(nonNil) == 1 {
		return nonNil[0]
	}
	return tree
}

func DropOrphanNonTerminal[E grammar.Element](tree *Tree[E]) *Tree[E] {
	if !tree.Node.Terminal() && len(nonNilChildren(tree)) == 0 {
		return nil
	} else {
		return tree
	}
}

func Compose[E fmt.Stringer](mapper ...TreeMapper[E]) TreeMapper[E] {
	return func(tree *Tree[E]) *Tree[E] {
		for i := len(mapper) - 1; i >= 0; i-- {
			tree = mapper[i](tree)
			if tree == nil {
				break
			}
		}
		return tree
	}
}

func nonNilChildren[E fmt.Stringer](tree *Tree[E]) []*Tree[E] {
	changed := false
	var newChildren []*Tree[E]
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

//func (mapper TreeMapper[E]) Then(next TreeMapper[E]) TreeMapper[E] {
//	return func(tree *Tree[E]) *Tree[E] {
//		tree = mapper(tree)
//		if tree == nil {
//			tree = next(tree)
//		}
//		return tree
//	}
//}
