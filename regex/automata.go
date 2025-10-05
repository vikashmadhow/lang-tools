package regex

import (
	"maps"
	"strconv"
	"strings"
)

type (
	stateObj struct{ _ uint8 }

	state *stateObj
	//state uint32

	targetState struct {
		pattern string
		char    *char
		state   state
	}

	NfaTrans map[state][]*targetState

	DfaTrans map[state]map[string]*targetState

	Automata[T interface{ DfaTrans | NfaTrans }] struct {
		Trans    T
		start    state
		final    []state
		finalMap map[state]bool
	}

	Dfa Automata[DfaTrans]

	Nfa Automata[NfaTrans]

	set[T comparable] map[T]bool
)

// Dfa implements Thomson's algorithm for converting regular expression to DFA.
func (nfa *Nfa) Dfa() *Dfa {
	dfa := Dfa{
		Trans:    make(DfaTrans),
		start:    nil,
		final:    []state{},
		finalMap: map[state]bool{},
	}

	dfaStates := map[state]set[state]{}
	reachable := &set[state]{}
	eClosure(nfa.start, nfa.Trans, reachable)
	explored := []set[state]{*reachable}

	dfa.start = &stateObj{}
	dfaStates[dfa.start] = *reachable
	if nfa.containsFinal(reachable) {
		dfa.final = append(dfa.final, dfa.start)
		dfa.finalMap[dfa.start] = true
	}

	for len(explored) > 0 {
		dfaState := explored[0]
		explored = explored[1:]
		source := find(dfaStates, dfaState)

		// union all outgoing character transitions on any state of the DFA state
		chars := map[string][]*char{}
		for s := range dfaState {
			trans := nfa.Trans[s]
			for _, t := range trans {
				if !t.char.IsEmpty() {
					chars[t.pattern] = append(chars[t.pattern], t.char)
				}
			}
		}

		// find reachable set of states for each outgoing character
		for pattern, cs := range chars {
			reachable = &set[state]{}
			groups := set[int]{}
			var combinedChar *char
			for _, c := range cs {
				if combinedChar == nil {
					combinedChar = c
				}
				maps.Insert(groups, maps.All(c.groups))
				//for i := c.groups.Front(); i != nil; i = i.Next() {
				//	groups[i.Value.(int)] = true
				//}
				for s := range dfaState {
					for _, t := range nfa.Trans[s] {
						if t.pattern == pattern {
							eClosure(t.state, nfa.Trans, reachable)
						}
					}
				}
			}

			//union := slices.Sorted(maps.Keys(groups))
			//newGroups := list.New()
			//for _, g := range union {
			//	newGroups.PushBack(g)
			//}
			combinedChar.groups = groups

			target := find(dfaStates, *reachable)
			if target == nil {
				target = &stateObj{}
				dfaStates[target] = *reachable
				explored = append(explored, *reachable)
			}
			if nfa.containsFinal(reachable) && !dfa.finalMap[target] {
				dfa.final = append(dfa.final, target)
				dfa.finalMap[target] = true
			}
			_, ok := dfa.Trans[source]
			if !ok {
				dfa.Trans[source] = map[string]*targetState{}
			}
			dfa.Trans[source][pattern] = &targetState{pattern, combinedChar, target}
		}
	}
	return &dfa
}

// Minimize implements Hopcroft's algorithm for minimizing the number of states in a DFA.
func (dfa *Dfa) Minimize() *Dfa {
	// initially partition states into two: one containing final states and the other, non-final states
	partitions := map[state]int{}
	partitionSize := map[int]int{}
	for s, trans := range dfa.Trans {
		dfa.partition(s, partitions, partitionSize)
		for _, t := range trans {
			dfa.partition(t.state, partitions, partitionSize)
		}
	}
	dfa.partition(dfa.start, partitions, partitionSize)
	for _, f := range dfa.final {
		dfa.partition(f, partitions, partitionSize)
	}
	maxPartitions := len(partitions) - 1 // special-case for empty regex having single partition containing final state

	// split partitions by extracting subset of states equivalent to each other.
	// 2 states are equivalent if they have the same outgoing character DfaTrans
	// to states that are in the same partition. e.g.: (s1, s2) --c--> (t1, t2)
	// continue until no more equivalence classes can be extracted.
	for changed := true; changed; {
		changed = false

	split:
		for s, sPartition := range partitions {
			if partitionSize[sPartition] > 1 {
				for c1, t1 := range dfa.Trans[s] {
					equivPartition := []state{s}
					for other, otherPartition := range partitions {
						if other != s && otherPartition == sPartition {
							equivalent := false
							for c2, t2 := range dfa.Trans[other] {
								if partitions[t1.state] == partitions[t2.state] && c1 == c2 {
									equivalent = true
									break
								}
							}
							if equivalent {
								equivPartition = append(equivPartition, other)
							}
						}
					}
					if len(equivPartition) < partitionSize[sPartition] {
						maxPartitions++
						for _, t := range equivPartition {
							partitions[t] = maxPartitions
						}
						partitionSize[sPartition] -= len(equivPartition)
						partitionSize[maxPartitions] = len(equivPartition)
						changed = true
						break split
					}
				}
			}
		}
	}

	// construct minimized DFA with each partition as a separate state
	newStates := map[int]state{}
	splits := map[int][]state{}
	for s, p := range partitions {
		if _, ok := newStates[p]; !ok {
			newStates[p] = &stateObj{}
		}
		splits[p] = append(splits[p], s)
	}

	newTrans := map[state]map[state][]*char{}
	newAuto := &Dfa{Trans: make(DfaTrans), finalMap: make(map[state]bool)}
	for p, states := range splits {
		from := newStates[p]
		for _, s := range states {
			if dfa.start == s {
				newAuto.start = from
			}
			if dfa.finalMap[s] && !newAuto.finalMap[from] {
				newAuto.final = append(newAuto.final, from)
				newAuto.finalMap[from] = true
			}

			// combine character DfaTrans for identical pairs of states
			for c, t := range dfa.Trans[s] {
				if newTrans[from] == nil {
					newTrans[from] = map[state][]*char{}
				}
				to := newStates[partitions[t.state]]
				existing := false
				for _, c2 := range newTrans[from][to] {
					//if slices.Equal(c.spanSet(), c2.spanSet()) {
					if c == c2.String() {
						existing = true
						break
					}
				}
				if !existing {
					newTrans[from][to] = append(newTrans[from][to], t.char)
				}
			}
		}
	}
	for from, t := range newTrans {
		if newAuto.Trans[from] == nil {
			newAuto.Trans[from] = map[string]*targetState{}
		}
		for to, trans := range t {
			if len(trans) == 1 {
				pattern := trans[0].String()
				newAuto.Trans[from][pattern] = &targetState{pattern, trans[0], to}
			} else {
				var charSets spanSet
				groups := set[int]{}
				for _, c := range trans {
					charSets = append(charSets, c.set...)
					maps.Insert(groups, maps.All(c.groups))

					//for i := c.groups.Front(); i != nil; i = i.Next() {
					//	groups.PushBack(i.Value.(int))
					//}
				}
				ch := chars(trans[0].modifier, groups, false, charSets...)
				pattern := ch.String()
				newAuto.Trans[from][pattern] = &targetState{pattern, ch, to}
			}
		}
	}
	return newAuto
}

// minimize implements the simpler Brzozowski's algorithm for minimizing the number of states in a DFA.
func (dfa *Dfa) minimize() *Dfa {
	return dfa.Reverse().Dfa().Reverse().Dfa()
}

func (dfa *Dfa) Reverse() *Nfa {
	inverse := Nfa{
		Trans:    make(NfaTrans),
		start:    nil,
		final:    []state{dfa.start},
		finalMap: map[state]bool{dfa.start: true},
	}
	for from, trans := range dfa.Trans {
		for c, to := range trans {
			inverse.Trans[to.state] = append(inverse.Trans[to.state],
				&targetState{c, to.char, from})
		}
	}
	if len(dfa.final) > 1 {
		inverse.start = &stateObj{}
		inverse.Trans[inverse.start] = []*targetState{}
		for _, f := range dfa.final {
			inverse.Trans[inverse.start] = append(inverse.Trans[inverse.start],
				&targetState{"", emptyChar, f})
		}
	} else {
		inverse.start = dfa.final[0]
	}
	return &inverse
}

// Add state to partitions and increase partition size, if necessary.
func (dfa *Dfa) partition(s state, partitions map[state]int, partitionSize map[int]int) {
	p := 1
	if dfa.finalMap[s] {
		p = 0
	}
	if _, ok := partitions[s]; !ok {
		partitions[s] = p
		partitionSize[p]++
	}
}

func (nfa *Nfa) containsFinal(reachable *set[state]) bool {
	for s := range *reachable {
		if nfa.finalMap[s] {
			return true
		}
	}
	return false
}

func (nfa *Nfa) merge(source *Nfa) *Nfa {
	for k, v := range source.Trans {
		nfa.Trans[k] = v
	}
	return nfa
}

func (nfa *Nfa) addTransitions(from state, to ...*targetState) *Nfa {
	_, ok := nfa.Trans[from]
	if !ok {
		nfa.Trans[from] = to
	} else {
		for _, v := range to {
			nfa.Trans[from] = append(nfa.Trans[from], v)
		}
	}
	return nfa
}

func (nfa *Nfa) GraphViz(title string) string {
	nodeNames := map[state]string{}
	if !nfa.finalMap[nfa.start] {
		nodeNames[nfa.start] = "S"
	}
	for i, f := range nfa.final {
		nodeNames[f] = "F" + strconv.Itoa(i+1)
	}
	nodeCount := 1

	spec := "digraph G {\n"
	if len(title) > 0 {
		spec += "\tlabel=\"" + title + "\"\n"
	}
	spec += "\t{\n"
	if !nfa.finalMap[nfa.start] {
		spec += "\t\t\"" + nodeNames[nfa.start] + "\" [shape=circle color=\"lightblue\" style=filled]\n"
	}
	for _, f := range nfa.final {
		if f == nfa.start {
			spec += "\t\t\"" + nodeNames[f] + "\" [shape=doublecircle color=\"lightblue\" style=filled]\n"
		} else {
			spec += "\t\t\"" + nodeNames[f] + "\" [shape=doublecircle style=filled]\n"
		}
	}
	spec += "\t}\n"

	for s, v := range nfa.Trans {
		_, ok := nodeNames[s]
		if !ok {
			nodeNames[s] = strconv.Itoa(nodeCount)
			nodeCount++
		}
		for _, t := range v {
			_, ok := nodeNames[t.state]
			if !ok {
				nodeNames[t.state] = strconv.Itoa(nodeCount)
				nodeCount++
			}
			spec += "\t\"" + nodeNames[s] + "\" -> \"" + nodeNames[t.state] + "\" [label=\"" + t.pattern + ":" + label(t.char.groups) + "\"]\n"
		}
	}

	spec += "}"
	return spec
}

func (dfa *Dfa) GraphViz(title string) string {
	nodeNames := map[state]string{}
	if !dfa.finalMap[dfa.start] {
		nodeNames[dfa.start] = "S"
	}
	for i, f := range dfa.final {
		nodeNames[f] = "F" + strconv.Itoa(i+1)
	}
	nodeCount := 1

	spec := "digraph G {\n"
	if len(title) > 0 {
		spec += "\tlabel=\"" + title + "\"\n"
	}
	spec += "\t{\n"
	if !dfa.finalMap[dfa.start] {
		spec += "\t\t\"" + nodeNames[dfa.start] + "\" [shape=circle color=\"lightblue\" style=filled]\n"
	}
	for _, f := range dfa.final {
		if f == dfa.start {
			spec += "\t\t\"" + nodeNames[f] + "\" [shape=doublecircle color=\"lightblue\" style=filled]\n"
		} else {
			spec += "\t\t\"" + nodeNames[f] + "\" [shape=doublecircle style=filled]\n"
		}
	}
	spec += "\t}\n"

	for s, v := range dfa.Trans {
		_, ok := nodeNames[s]
		if !ok {
			nodeNames[s] = strconv.Itoa(nodeCount)
			nodeCount++
		}
		for _, t := range v {
			_, ok := nodeNames[t.state]
			if !ok {
				nodeNames[t.state] = strconv.Itoa(nodeCount)
				nodeCount++
			}
			spec += "\t\"" + nodeNames[s] + "\" -> \"" + nodeNames[t.state] + "\" [label=\"" + t.pattern + ":" + label(t.char.groups) + "\"]\n"
		}
	}

	spec += "}"
	return spec
}

func eClosure(from state, trans NfaTrans, closure *set[state]) {
	(*closure)[from] = true
	for _, to := range trans[from] {
		if to.char.IsEmpty() && !(*closure)[to.state] {
			eClosure(to.state, trans, closure)
		}
	}
}

func find(states map[state]set[state], state set[state]) state {
	for k, v := range states {
		if maps.Equal(v, state) {
			return k
		}
	}
	return nil
}

func label(groups set[int]) string {
	var s strings.Builder
	if groups != nil {
		first := true
		//for g := groups.Front(); g != nil; g = g.Next() {
		for g := range groups {
			if first {
				first = false
			} else {
				s.WriteRune(',')
			}
			s.WriteString(strconv.Itoa(g))
		}
	}
	return s.String()
}

func charNfa(c *char) *Nfa {
	final := &stateObj{}
	a := Nfa{
		Trans:    make(NfaTrans),
		start:    &stateObj{},
		final:    []state{final},
		finalMap: map[state]bool{final: true},
	}
	a.addTransitions(a.start, &targetState{c.String(), c, final})
	return &a
}
