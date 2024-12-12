package main

import (
	"sync"
	"testing"
)

func SetFlags() {
	var rec int = -1
	recolor_policy = &rec
}

func TestOrdering(t *testing.T) {
	SetFlags()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			S := Gnp(1e2, 1e-2)

			for j := uint64(0); j < 1e2-1; j++ {
				S.AddEdge(j, j+1)
			}

			ret1 := FindAll(S, S, nil, 2)
			ret2 := FindAll(S, S, nil, 2)
			if ret1 != ret2 {
				t.Errorf("Ordering changed output")
			}
			if ret1 > 3 {
				t.Errorf("Find more finds? ordered")
				t.Log(ret1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestFindAll(t *testing.T) {
	SetFlags()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			S := Gnp(1e3, 1e-2)

			for j := uint64(0); j < 1e3-1; j++ {
				S.AddEdge(j, j+1)
			}
			ret := FindAll(S, S, nil, 2)
			if ret != 1 {
				t.Errorf("Find all multiple finds?")
				t.Log(ret)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestSelfFind(t *testing.T) {
	SetFlags()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			S := Gnp(1e2, 1e-1)
			for j := uint64(0); j < 99; j++ {
				S.AddEdge(j, j+1)
			}
			ordering := make([]uint64, len(S))
			for i := 0; i < len(ordering); i++ {
				ordering[i] = uint64(i)
			}
			context := context{
				Graph:        S,
				Subgraph:     S,
				restrictions: make(map[uint64]map[uint64]void),
				path:         make(map[uint64]uint64),
				chosen:       make(map[uint64]void),
				prior:        nil,
			}
			ret := RecursionSearch(&context, 0, 0)
			if ret == 0 {
				t.Errorf("didnt find itself")
			}
			wg.Done()
		}()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			S := Gnp(1e1, 1e-1)
			G := Gnp(1e2, 1e-2)
			for j := uint64(0); j < 10; j++ {
				G.AddVertex(1e2+j, S[j].attribute.color)
			}
			for j := uint64(0); j < 10; j++ {
				for k := uint64(0); k < 10; k++ {
					G.AddEdge(1e2+j, 1e2+k)
				}
			}

			for j := uint64(0); j < 9; j++ {
				S.AddEdge(j, j+1)
			}
			ret := 0
			ordering := make([]uint64, len(S))
			for i := 0; i < len(ordering); i++ {
				ordering[i] = uint64(i)
			}
			for u := uint64(0); u < uint64(len(G)); u++ {
				if G[u].attribute.color == S[0].attribute.color {
					context := context{
						Graph:        G,
						Subgraph:     S,
						restrictions: make(map[uint64]map[uint64]void),
						path:         make(map[uint64]uint64),
						chosen:       make(map[uint64]void),
						prior:        nil,
						prior_policy: 2,
					}
					ret += RecursionSearch(&context, u, 0)
				}
			}
			if ret < 1 {
				t.Errorf("Too little")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
