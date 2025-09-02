// author: Vikash Madhow (vikash.madhow@gmail.com)

package regex

import (
	"container/list"
	"maps"
	"reflect"
	"slices"
	"strconv"
)

type (
	stateObj struct{ _ uint8 }

	state *stateObj

	transitions map[state]map[char]state

	set[T comparable] map[T]bool

	automata struct {
		Trans transitions
		start state
		final []state
	}
)

func (auto *automata) dfa() *automata {
	dfa := automata{
		Trans: make(transitions),
		start: nil,
		final: []state{},
	}

	dfaStates := map[state]set[state]{}
	explored := make(chan set[state], 1000)
	reachable := &set[state]{}
	eClosure(auto.start, auto.Trans, reachable)
	explored <- *reachable

	dfa.start = &stateObj{}
	dfaStates[dfa.start] = *reachable
	if auto.containsFinal(reachable) {
		dfa.final = append(dfa.final, dfa.start)
	}

	for len(explored) > 0 {
		dfaState := <-explored
		source := find(dfaStates, dfaState)

		// union all outgoing character transitions on any State of the DFA State
		chars := map[string][]char{}
		for s := range dfaState {
			trans := auto.Trans[s]
			for c := range trans {
				if !c.isEmpty() {
					pattern := c.Pattern()
					chars[pattern] = append(chars[pattern], c)
				}
			}
		}

		// find reachable set of states for each outgoing character
		for _, cs := range chars {
			reachable = &set[state]{}
			groups := set[int]{}
			var combinedChar char = nil
			for _, c := range cs {
				if combinedChar == nil {
					combinedChar = c
				}
				for i := c.groups().Front(); i != nil; i = i.Next() {
					groups[i.Value.(int)] = true
				}
				for s := range dfaState {
					trans := auto.Trans[s]
					if t, ok := trans[c]; ok {
						eClosure(t, auto.Trans, reachable)
					}
				}
			}

			union := slices.Sorted(maps.Keys(groups))
			newGroups := list.New()
			for _, g := range union {
				newGroups.PushBack(g)
			}
			combinedChar.setGroups(newGroups)

			target := find(dfaStates, *reachable)
			if target == nil {
				target = &stateObj{}
				dfaStates[target] = *reachable
				explored <- *reachable
			}
			if auto.containsFinal(reachable) && slices.Index(dfa.final, target) == -1 {
				dfa.final = append(dfa.final, target)
			}
			_, ok := dfa.Trans[source]
			if !ok {
				dfa.Trans[source] = map[char]state{}
			}
			dfa.Trans[source][combinedChar] = target
		}
	}
	return &dfa
}

func (auto *automata) containsFinal(reachable *set[state]) bool {
	for s := range *reachable {
		if s == auto.final[0] {
			return true
		}
	}
	return false
}

func (auto *automata) merge(source *automata) *automata {
	for k, v := range source.Trans {
		auto.Trans[k] = v
	}
	return auto
}

func (auto *automata) addTransitions(from state, to map[char]state) *automata {
	existing, ok := auto.Trans[from]
	if !ok {
		auto.Trans[from] = to
	} else {
		for k, v := range to {
			existing[k] = v
		}
	}
	return auto
}

func (auto *automata) ToGraphViz(title string) string {
	nodeNames := map[state]string{}
	if slices.Index(auto.final, auto.start) == -1 {
		nodeNames[auto.start] = "S"
	}
	for i, f := range auto.final {
		nodeNames[f] = "F" + strconv.Itoa(i+1)
	}
	nodeCount := 1

	spec := "digraph G {\n"
	if len(title) > 0 {
		spec += "\tlabel=\"" + title + "\"\n"
	}
	spec += "\t{\n"
	if slices.Index(auto.final, auto.start) == -1 {
		spec += "\t\t\"" + nodeNames[auto.start] + "\" [shape=circle color=\"lightblue\" style=filled]\n"
	}
	for _, f := range auto.final {
		if f == auto.start {
			spec += "\t\t\"" + nodeNames[f] + "\" [shape=doublecircle color=\"lightblue\" style=filled]\n"
		} else {
			spec += "\t\t\"" + nodeNames[f] + "\" [shape=doublecircle style=filled]\n"
		}
	}
	spec += "\t}\n"
	for s, v := range auto.Trans {
		_, ok := nodeNames[s]
		if !ok {
			nodeNames[s] = strconv.Itoa(nodeCount)
			nodeCount++
		}
		for c, t := range v {
			_, ok := nodeNames[t]
			if !ok {
				nodeNames[t] = strconv.Itoa(nodeCount)
				nodeCount++
			}
			spec += "\t\"" + nodeNames[s] + "\" -> \"" + nodeNames[t] + "\" [label=\"" + c.Pattern() + ":" + label(c.groups()) + "\"]\n"
		}
	}
	spec += "}"
	return spec
}

func eClosure(from state, trans transitions, closure *set[state]) {
	(*closure)[from] = true
	for ch, to := range trans[from] {
		if ch.isEmpty() && !(*closure)[to] {
			eClosure(to, trans, closure)
		}
	}
}

func find(states map[state]set[state], state set[state]) state {
	for k, v := range states {
		if reflect.DeepEqual(v, state) {
			return k
		}
	}
	return nil
}

func label(groups *list.List) string {
	s := ""
	if groups != nil {
		first := true
		for g := groups.Front(); g != nil; g = g.Next() {
			if first {
				first = false
			} else {
				s += ","
			}
			s += strconv.Itoa(g.Value.(int))
		}
	}
	return s
}

func charNfa(c char) *automata {
	a := automata{
		Trans: make(transitions),
		start: &stateObj{},
		final: []state{&stateObj{}},
	}
	a.addTransitions(a.start, map[char]state{c: a.final[0]})
	return &a
}
