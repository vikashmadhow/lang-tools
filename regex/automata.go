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
	//state uint32

	transitions map[state]map[char]state

	set[T comparable] map[T]bool

	automata struct {
		Trans    transitions
		start    state
		final    []state
		finalMap map[state]bool
	}
)

// Thomson's algorithm for converting regular expression to DFA.
func (auto *automata) dfa() *automata {
	dfa := automata{
		Trans:    make(transitions),
		start:    nil,
		final:    []state{},
		finalMap: map[state]bool{},
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
		dfa.finalMap[dfa.start] = true
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
					pattern := c.String()
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
			if auto.containsFinal(reachable) && !dfa.finalMap[target] {
				dfa.final = append(dfa.final, target)
				dfa.finalMap[target] = true
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

// Hopcroft's algorithm for minimizing DFA
func (auto *automata) minimize() *automata {
	// initially partition states into two: one containing final states and the other, non-final states
	partitions := map[state]int{}
	partitionSize := map[int]int{}
	for s, trans := range auto.Trans {
		auto.partition(s, partitions, partitionSize)
		for _, t := range trans {
			auto.partition(t, partitions, partitionSize)
		}
	}
	auto.partition(auto.start, partitions, partitionSize)
	for _, f := range auto.final {
		auto.partition(f, partitions, partitionSize)
	}
	maxPartitions := len(partitions) - 1 // special-case for empty regex having single partition containing final state

	// split partitions by extracting subset of states equivalent to each other.
	// 2 states are equivalent if they have the same outgoing character transitions
	// to states that are in the same partition. e.g.: (s1, s2) --c--> (t1, t2)
	// continue until no more equivalence classes can be extracted.
	changed := true
	for changed {
		changed = false

	split:
		for s, sPartition := range partitions {
			if partitionSize[sPartition] > 1 {
				for c1, t1 := range auto.Trans[s] {
					equivPartition := []state{s}
					for other, otherPartition := range partitions {
						if other != s && otherPartition == sPartition {
							for c2, t2 := range auto.Trans[other] {
								if partitions[t1] == partitions[t2] && slices.Equal(c1.spanSet(), c2.spanSet()) {
									equivPartition = append(equivPartition, other)
									break
								}
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

	newTrans := map[state]map[state][]char{}
	newAuto := &automata{Trans: make(transitions), finalMap: make(map[state]bool)}
	for p, states := range splits {
		from := newStates[p]
		for _, s := range states {
			if auto.start == s {
				newAuto.start = from
			}
			if auto.finalMap[s] && !newAuto.finalMap[from] {
				newAuto.final = append(newAuto.final, from)
				newAuto.finalMap[from] = true
			}

			// combine character transitions for identical pair of states
			for c, t := range auto.Trans[s] {
				if newTrans[from] == nil {
					newTrans[from] = map[state][]char{}
				}
				to := newStates[partitions[t]]
				existing := false
				for _, c2 := range newTrans[from][to] {
					if slices.Equal(c.spanSet(), c2.spanSet()) {
						existing = true
						break
					}
				}
				if !existing {
					newTrans[from][to] = append(newTrans[from][to], c)
				}
			}
		}
	}
	for from, t := range newTrans {
		if newAuto.Trans[from] == nil {
			newAuto.Trans[from] = map[char]state{}
		}
		for to, trans := range t {
			if len(trans) == 1 {
				newAuto.Trans[from][trans[0]] = to
			} else {
				charSets := list.New()
				for _, c := range trans {
					charSets.PushBack(c)
				}
				ch := &charSet{trans[0].modifier(), false, *charSets, cp(trans[0].groups()), nil}
				newAuto.Trans[from][ch] = to
			}
		}
	}
	return newAuto
}

// Add state to partitions and increase partition size, if necessary.
func (auto *automata) partition(s state, partitions map[state]int, partitionSize map[int]int) {
	p := 1
	if auto.finalMap[s] {
		p = 0
	}
	if _, ok := partitions[s]; !ok {
		partitions[s] = p
		partitionSize[p]++
	}
}

func (auto *automata) chars() map[char]bool {
	// union all outgoing character transitions on any State of the DFA State
	chars := map[char]bool{}
	for _, trans := range auto.Trans {
		for c := range trans {
			chars[c] = true
		}
	}
	return chars
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

func (auto *automata) GraphViz(title string) string {
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
			spec += "\t\"" + nodeNames[s] + "\" -> \"" + nodeNames[t] + "\" [label=\"" + c.String() + ":" + label(c.groups()) + "\"]\n"
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
