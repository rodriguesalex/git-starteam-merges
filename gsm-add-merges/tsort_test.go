package main

import (
	"reflect"
	"testing"
)

func TestTopoSortCycle(t *testing.T) {
	g := Graph{
		Vertices: []Vertex{
			{Edges: []int{1}}, // 0
			{Edges: []int{2}}, // 1
			{Edges: []int{0}}, // 2
		},
	}
	_, err := TopoSort(g)
	if err != errNotDAG {
		t.Errorf("want errNotDAG on cycle")
	}
}

func TestTopoSort(t *testing.T) {
	for i, tt := range []struct {
		g    Graph
		want []int
	}{
		{
			g: Graph{
				Vertices: []Vertex{
					{Edges: []int{}},  // 0
					{Edges: []int{0}}, // 1
					{Edges: []int{1}}, // 2
					{Edges: []int{2}}, // 3
				},
			},
			want: []int{0, 1, 2, 3},
		},
		{
			g: Graph{
				Vertices: []Vertex{
					{Edges: []int{1}}, // 0
					{Edges: []int{2}}, // 1
					{Edges: []int{3}}, // 2
					{Edges: []int{}},  // 3
				},
			},
			want: []int{3, 2, 1, 0},
		},
		{
			g: Graph{
				Vertices: []Vertex{
					{Edges: []int{1, 2, 3}}, // 0
					{Edges: []int{}},        // 1
					{Edges: []int{}},        // 2
					{Edges: []int{}},        // 3
				},
			},
			want: []int{1, 2, 3, 0},
		},
		{
			g: Graph{
				Vertices: []Vertex{
					{Edges: []int{3}}, // 0
					{Edges: []int{3}}, // 1
					{Edges: []int{3}}, // 2
					{Edges: []int{}},  // 3
					{Edges: []int{7}}, // 4
					{Edges: []int{7}}, // 5
					{Edges: []int{7}}, // 6
					{Edges: []int{}},  // 7
				},
			},
			want: []int{3, 0, 1, 2, 7, 4, 5, 6},
		},
		{
			g: Graph{
				Vertices: []Vertex{
					{Edges: []int{}},      // 0
					{Edges: []int{}},      // 1
					{Edges: []int{1}},     // 2
					{Edges: []int{1}},     // 3
					{Edges: []int{1}},     // 4
					{Edges: []int{4, 8}},  // 5
					{Edges: []int{5}},     // 6
					{Edges: []int{6, 10}}, // 7
					{Edges: []int{1}},     // 8
					{Edges: []int{3, 8}},  // 9
					{Edges: []int{9}},     // 10
					{Edges: []int{10}},    // 11
				},
			},
			want: []int{0, 1, 2, 3, 4, 8, 5, 6, 9, 10, 7, 11},
		},
	} {
		got, err := TopoSort(tt.g)
		if err != nil {
			t.Errorf("[%d] error: %v", err)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("[%d] tsort got=%#v, want=%#v", i, got, tt.want)
		}
	}

}
