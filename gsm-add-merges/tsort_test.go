package main

import (
	"testing"
)

func TestTopoSort(t *testing.T) {
	/*
		1 2
		1 3
		1 4
		4 5
		8 5
		5 6
		6 7
		10 7
		1 8
		8 9
		3 9
		9 10
		10 11
	*/

	g := Graph{
		Edges: []Edges{
			{Y: []int{}},      // 0
			{Y: []int{}},      // 1
			{Y: []int{1}},     // 2
			{Y: []int{1}},     // 3
			{Y: []int{1}},     // 4
			{Y: []int{4, 8}},  // 5
			{Y: []int{5}},     // 6
			{Y: []int{6, 10}}, // 7
			{Y: []int{1}},     // 8
			{Y: []int{3, 8}},  // 9
			{Y: []int{9}},     // 10
			{Y: []int{10}},    // 11
		},
	}

	tsort := TopoSort(g)
	t.Errorf("tsort=%#v", tsort)
}
