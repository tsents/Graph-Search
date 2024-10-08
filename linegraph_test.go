package main

import (
	"sync"
	"testing"
)

func SetFlags() {
	prior := 1
	prior_policy = &prior
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
			ret := FindAll(S, S, nil)
			if ret != 1 {
				t.Errorf("Find all multiple finds?")
				t.Log(ret)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
