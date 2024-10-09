package main

import (
	"testing"
)

func SetFlags() {
	dir := false
	directed = &dir
	prior := 2
	prior_policy = &prior
}

func TestFindAll(t *testing.T) {
	SetFlags()
	for i := 0; i < 10; i++ {
		S := Gnp(1e3, 1e-2)

		for j := uint64(0); j < 1e3-1; j++ {
			S.AddEdge(j, j+1)
		}
		ret := FindAll(S, S, nil)
		if ret != 1 {
			t.Errorf("Find all multiple finds?")
			t.Log(ret)
		}
	}
}

func TestDirected(t *testing.T) {
	SetFlags()
	dir := true
	directed = &dir
	for i := 0; i < 10; i++ {
		G := Gnp(1e3, 1e-2)
		var S graph = make(graph)
		for j := uint64(0); j < 1e3-1; j++ {
			G.AddEdge(j, j+1)
		}
		{ //code block for the base case in the j loop
			S.AddVertex(0, 0)
			vertex := G[0]
			vertex.attribute.color = 0
			G[0] = vertex
		}
		for j := uint64(0); j < 1e1-1; j++ {
			vertex := G[j+1]
			vertex.attribute.color = 0
			G[j+1] = vertex
			S.AddVertex(j+1, 0)
			S.AddEdge(j, j+1)
		}
		ret := FindAll(G, S, nil)
		if ret == 0 {
			t.Errorf("should be multiple finds")
			t.Log(ret)
		}
		for j := uint64(1e1); j < 2e1-1; j++ {
			vertex := G[j+1]
			vertex.attribute.color = 0
			G[j+1] = vertex
			S.AddVertex(j+1, 0)
			S.AddEdge(j+1, j)
		}
		ret = FindAll(G, S, nil)
		if ret != 0 {
			t.Errorf("exess finds?")
			t.Log(ret)
		}
	}
}
