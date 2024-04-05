package env

import (
	"sort"
)

const TopologySystemCount = 1

type QDaggerBase struct {
	parent *QDaggerBase
}

type QDaggerBasePin struct {
	QDaggerBase
	parentNode *QDaggerNode
	inputPins  [TopologySystemCount]*QDaggerPinCollection
	outputPins [TopologySystemCount]*QDaggerPinCollection
}

type PinDirection int

const (
	PinDirectionInput PinDirection = iota
	PinDirectionOutput
)

type QDaggerInputPin struct {
	QDaggerBasePin
	connectedTo *QDaggerBasePin
}

type QDaggerOutputPin struct {
	QDaggerBasePin
	connectedTo []interface{}
}

type QDaggerPinCollection struct {
	QDaggerBase
	parentNode     *QDaggerNode
	direction      PinDirection
	orderedPins    []interface{}
	pinMap         map[string]*QDaggerBasePin
	topologySystem int
}

type QDaggerNode struct {
	QDaggerBase
	parentGraph         *QDaggerGraph
	inputPins           [TopologySystemCount]*QDaggerPinCollection
	outputPins          [TopologySystemCount]*QDaggerPinCollection
	ordinal             [TopologySystemCount]int
	subgraphAffiliation [TopologySystemCount]int
	descendents         [TopologySystemCount][]*QDaggerNode
	currentTSystemEval  int
}

type QDaggerGraph struct {
	QDaggerBase
	nodes         []*QDaggerNode
	subGraphCount [TopologySystemCount]int
	maxOrdinal    [TopologySystemCount]int
}

func (n *QDaggerNode) IsTopLevel(topologySystem int) bool {
	for _, pin := range n.inputPins[topologySystem].orderedPins {
		ipin := pin.(*QDaggerInputPin)
		if ipin.connectedTo != nil && ipin.connectedTo.parentNode != nil {
			return false
		}
	}
	return true
}

func (g *QDaggerGraph) TopLevelNodes(topologySystem int) []*QDaggerNode {
	var retv []*QDaggerNode
	for _, node := range g.nodes {
		if node.IsTopLevel(topologySystem) {
			retv = append(retv, node)
		}
	}
	return retv
}

func nodeComparer(n1, n2 *QDaggerNode) bool {
	return n1.ordinal[n1.currentTSystemEval] < n2.ordinal[n2.currentTSystemEval]
}

func (g *QDaggerGraph) recurseCalculateTopology(level int, node *QDaggerNode, touchedSet map[*QDaggerNode]bool, topologySystem int) map[*QDaggerNode]bool {
	retv := make(map[*QDaggerNode]bool)
	if node == nil {
		return retv
	}

	node.ordinal[topologySystem] = max(level, node.ordinal[topologySystem])
	g.maxOrdinal[topologySystem] = max(g.maxOrdinal[topologySystem], node.ordinal[topologySystem])

	for _, p1 := range node.outputPins[topologySystem].orderedPins {
		output := p1.(*QDaggerOutputPin)
		for _, p2 := range output.connectedTo {
			inpin := p2.(*QDaggerInputPin)
			rset := g.recurseCalculateTopology(level+1, inpin.parentNode, touchedSet, topologySystem)
			for k := range rset {
				retv[k] = true
			}
		}
	}

	touchedSet[node] = true

	for snode := range retv {
		if !contains(node.descendents[topologySystem], snode) {
			node.descendents[topologySystem] = append(node.descendents[topologySystem], snode)
		}
	}

	for _, tn := range node.descendents[topologySystem] {
		tn.currentTSystemEval = topologySystem
	}
	sort.Slice(node.descendents[topologySystem], func(i, j int) bool {
		return nodeComparer(node.descendents[topologySystem][i], node.descendents[topologySystem][j])
	})

	retv[node] = true

	return retv
}

func (g *QDaggerGraph) CalculateTopology() {
	for t := 0; t < TopologySystemCount; t++ {
		g.maxOrdinal[t] = 0

		for _, node := range g.nodes {
			node.ordinal[t] = -1
			node.subgraphAffiliation[t] = -1
			node.descendents[t] = nil
		}

		tnodes := g.TopLevelNodes(t)

		var touchedSetList []map[*QDaggerNode]bool

		for i, node := range tnodes {
			node.ordinal[t] = 0

			touchedSet := make(map[*QDaggerNode]bool)

			for _, p1 := range node.outputPins[t].orderedPins {
				outpin := p1.(*QDaggerOutputPin)
				for _, p2 := range outpin.connectedTo {
					inpin := p2.(*QDaggerInputPin)
					newset := g.recurseCalculateTopology(1, inpin.parentNode, touchedSet, t)

					for setnode := range newset {
						if !contains(node.descendents[t], setnode) {
							node.descendents[t] = append(node.descendents[t], setnode)
						}
					}

					desc := node.descendents[t]
					for _, tn := range desc {
						tn.currentTSystemEval = t
					}
					sort.Slice(node.descendents[t], func(i, j int) bool {
						return nodeComparer(node.descendents[t][i], node.descendents[t][j])
					})
				}
			}

			touchedSet[node] = true

			if i == 0 {
				touchedSetList = append(touchedSetList, touchedSet)
			} else {
				merged := false
				for u := 0; u < len(touchedSetList); u++ {
					intersection := make(map[*QDaggerNode]bool)
					for k := range touchedSet {
						if touchedSetList[u][k] {
							intersection[k] = true
						}
					}

					if len(intersection) > 0 {
						for k := range touchedSet {
							touchedSetList[u][k] = true
						}
						merged = true
						break
					}
				}

				if !merged {
					touchedSetList = append(touchedSetList, touchedSet)
				}
			}
		}

		for i, set := range touchedSetList {
			for node := range set {
				node.subgraphAffiliation[t] = i
			}
		}

		g.subGraphCount[t] = len(touchedSetList)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func contains(slice []*QDaggerNode, node *QDaggerNode) bool {
	for _, n := range slice {
		if n == node {
			return true
		}
	}
	return false
}
