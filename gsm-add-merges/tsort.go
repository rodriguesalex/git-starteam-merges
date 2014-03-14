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

type Vertex struct {
	Edges []int
}

type Graph struct {
	Vertices []Vertex
}

func (g *Graph) AddEdge(src, dst int) {
	g.Vertices[src].Edges = append(g.Vertices[src].Edges, dst)
}

type TSort struct {
	sorted     []int
	discovered []bool
	processed  []bool
}

func TopoSort(g Graph) ([]int, error) {
	t := TSort{
		sorted:     make([]int, 0, len(g.Vertices)),
		discovered: make([]bool, len(g.Vertices)),
		processed:  make([]bool, len(g.Vertices)),
	}
	err := t.Sort(g)
	if err != nil {
		return nil, err
	}
	return t.sorted, nil
}

func (t *TSort) Sort(g Graph) error {
	for i, _ := range g.Vertices {
		if !t.discovered[i] {
			if err := t.dfs(g, i); err != nil {
				return err
			}
		}
	}
	return nil
}

var errNotDAG = errors.New("graph is not a DAG")

func (t *TSort) dfs(g Graph, v int) error {
	t.discovered[v] = true

	for _, y := range g.Vertices[v].Edges {
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
