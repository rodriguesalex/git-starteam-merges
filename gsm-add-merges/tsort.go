package main

import (
	"errors"
)

type edgeType int

// Tree edge classifications
const (
	UNKNOWN edgeType = iota
	TREE
	BACK
	FORWARD
	CROSS
)

/* Code directly translated from Skiena
func topsort(g *graph) {
	initStack(&sorted)
	for i := 1; i <= g.nvertices; i++ {
		if !discovered[i] {
			dfs(g, i)
		}
	}
}

func dfs(g *graph, v int) {
	var (
		p *edgenode // temporary pointer
		y int       // successor vertex
	)

	if finished {
		return
	}

	discovered[v] = true
	time++
	entryTime[v] = time

	processVertexEarly(v)

	p = g.edges[v]
	for p != nil {
		y = p.y
		if !discovered[y] {
			parent[y] = v
			processEdge(v, y)
			dfs(g, y)
		} else if !processed[y] {
			processEdge(v, y)
		}

		if finished {
			return
		}

		p = p.next
	}

	processVertexLate(v)
	time++
	exitTime[v] = time

	processed[v] = true
}

func processVertexLate(v int) {
	push(&sorted, v)
}

func classifyEdge(x, y int) edgeType {
	switch {
	case parent[y] == x:
		return TREE
	case discovered[y] && !processed[y]:
		return BACK
	case processed[y] && (entryTime[y] > entryTime[x]):
		return FORWARD
	case processed[y] && (entryTime[y] < entryTime[x]):
		return CROSS
	}
	return UNKNOWN
}

var errNotDAG = errors.New("graph is not a DAG")

func processEdge(x, y int) error {
	if classifyEdge(x, y) == BACK {
		return errNotDAG
	}
}
*/

type Edges struct {
	Y []int
}

type Graph struct {
	Edges []Edges
}

type TSort struct {
	sorted     []int
	discovered []bool
	processed  []bool
}

func TopoSort(g Graph) []int {
	t := TSort{
		discovered: make([]bool, len(g.Edges)),
		processed:  make([]bool, len(g.Edges)),
	}
	t.Sort(g)
	return t.sorted
}

func (t *TSort) Sort(g Graph) {
	for i, _ := range g.Edges {
		if !t.discovered[i] {
			t.dfs(g, i)
		}
	}
}

var errNotDAG = errors.New("graph is not a DAG")

func (t *TSort) dfs(g Graph, v int) error {
	t.discovered[v] = true

	for _, y := range g.Edges[v].Y {
		switch {
		case !t.discovered[y]:
			if err := t.dfs(g, y); err != nil {
				return err
			}
		case !t.processed[y]:
			return errNotDAG
		}
	}

	t.sorted = append(t.sorted, v)
	t.processed[v] = true

	return nil
}
