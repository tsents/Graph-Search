package main

import (
	"math/rand/v2"
	"sync"
	"testing"
)

func TestSelfFind(t *testing.T){
	var wg sync.WaitGroup
	// for i := 0; i < 1; i++{ just slow and uninformetive
	// 	wg.Add(1)
	// 	go func() {
	// 		S := Gnp(10,1)
	// 		RecursionSearch(S,S,0,0,make([]*list, len(S)),make(map[uint32]uint32))
	// 		wg.Done()
	// 	} ()
	// }
	for i := 0; i < 50; i++{
		wg.Add(1)
		go func () {
			S := Gnp(1e2,1e-1)
			for j := uint32(0); j < 99; j++{
				AddEdge(S,j,j+1)
			}
			ret := RecursionSearch(S,S,0,0,make([]*list, len(S)),make(map[uint32]uint32))
			if ret == 0 {
				t.Errorf("didnt find itself")
			}
			wg.Done()
		} ()
	}

	for i := 0; i < 40; i++{
		wg.Add(1)
		go func () {
			S := Gnp(1e1,1e-1)
			G := Gnp(1e2,1e-2)
			for j := uint32(0); j < 10; j++{
				AddVertex(G,1e2+j,S[j].attribute.color)
			}
			for j := uint32(0); j < 10; j++{
				for k := uint32(0); k < 10; k++{
					AddEdge(G,1e2+j,1e2+k)
				}
			}

			for j := uint32(0); j < 9; j++{
				AddEdge(S,j,j+1)
			}
			ret := 0
			for u := uint32(0); u < uint32(len(G)); u++{
				if G[u].attribute.color == S[0].attribute.color{
					ret += RecursionSearch(G,S,u,0,make([]*list, len(S)),make(map[uint32]uint32))
				}
			}
			if ret < 1 {
				t.Errorf("Too little")
			}
			wg.Done()
		} ()
	}
	wg.Wait()
}

func TestLists(t *testing.T) {
	for j := 0; j < 1e3; j++{
		l := list{nil,nil}
		for i := 0; i < 1e4; i++ {
			el := element{rand.Uint32N(1000000),nil}
			ListAppend(&l,&el)
		}
		if l.end == nil {
			t.Errorf("Wrong end")
		}
		k := 0
		for el := l.start; el != nil; el = el.next{
			k++
		}
		if k != 1e4 {
			t.Errorf("Wrong length")
		}
	}
}

func TestListsSplit(t *testing.T){
	for j := 0; j < 1e2; j++{
		l := list{nil,nil}
		for i := 0; i < 1e2; i++ {
			el := element{rand.Uint32N(1000000),nil}
			ListAppend(&l,&el)
		}
		var dis discriminator = func(u uint32) bool {
			return rand.Float32() <= 0.5
		}
		l1,l2 := SplitList(&l,dis)
		k := 0
		for el := l1.start; el != nil; el = el.next{
			k++
		}
		u := 0
		for el := l2.start; el != nil; el = el.next{
			u++
		}
		if k + u != 1e2 {
			t.Errorf("wrong length split")
		}
		l = *JoinLists(l1,l2)
		r := 0
		for el := l.start; el != nil; el = el.next{
			r++
		}
		if k + u != r {
			t.Errorf("wrong rejoin")
		}

		l1,l2 = SplitList(&l,dis)
		l = *JoinLists(l1,l2)
		l1,l2 = SplitList(&l,dis)
		l = *JoinLists(l1,l2)
		l1,l2 = SplitList(&l,dis)
		l = *JoinLists(l1,l2)
		l1,l2 = SplitList(&l,dis)
		l = *JoinLists(l1,l2)
		l1,l2 = SplitList(&l,dis)
		l = *JoinLists(l1,l2)
		l1,l2 = SplitList(&l,dis)
		l = *JoinLists(l1,l2)
		h := 0
		for el := l.start; el != nil; el = el.next{
			h++
		}
		if h != r {
			t.Errorf("Brocken split rejoin")
		}
		var cpy *element
		for el := l.start; el != nil; el = el.next{
			cpy = el
		}
		if cpy != l.end {
			t.Errorf("Wrong end in split rejoin")
		}
	}

}

func TestListsJoin(t *testing.T) {
	for j := 0; j < 1e3; j++{
		l1 := list{nil,nil}
		for i := 0; i < 1e3; i++ {
			el := element{rand.Uint32N(1000000),nil}
			ListAppend(&l1,&el)
		}
		l2 := list{nil,nil}
		for i := 0; i < 1e3; i++ {
			el := element{rand.Uint32N(1000000),nil}
			ListAppend(&l2,&el)
		}
		k := 0
		for el := l1.start; el != nil; el = el.next{
			k++
		}
		u := 0
		for el := l2.start; el != nil; el = el.next{
			u++
		}
		JoinLists(&l1,&l2)
		var cpy *element
		h := 0
		for el := l1.start; el != nil; el = el.next{
			h++
			cpy = el
		}
		if cpy != l1.end {
			t.Errorf("Wrong end in l1")
		}
		o := 0
		for el := l2.start; el != nil; el = el.next{
			o++
			cpy = el
		}
		if cpy != l2.end {
			t.Errorf("Wrong end in l2")
		}
		if o != u {
			t.Errorf("Wrong length in l2")
		}
		if h != k + u {
			t.Errorf("Wrong length in join")
		}
	}
}

