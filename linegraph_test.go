package main

import (
	"math/rand/v2"
	"testing"
)

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


