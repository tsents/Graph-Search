package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
)

type void struct{}

type discriminator func(uint32) bool

type graph map[uint32]vertex

type element struct {
	value uint32
	next  *element
}

type list struct {
	start *element
	end   *element
}

type vertex struct {
	attribute    att
	neighborhood map[uint32]void
}

type att struct {
	color uint8
}

// idea - store Graph as map of vertex -> neighboorhood -> void (neighborhood is a set)
func main() {
	fmt.Println("hello world")
	for t := 0; t < 50; t++{
		G := Gnp(1e4, 1e-3)
		S := Gnp(1e2, 1e-1)
		for j := uint32(0); j < uint32(len(S)) - 1; j++{
			AddEdge(S,j,j+1)
		}
		FindAllSubgraph(G,S)
		fmt.Println("done",t)
	}
}

func FindAllSubgraph(Graph graph, Subgraph graph) {
	var wg sync.WaitGroup
	for u := uint32(0); u < uint32(len(Graph)); u++ {
		if Graph[u].attribute.color == Subgraph[0].attribute.color {
			wg.Add(1)
			go func(u uint32) {
				RecursionSearch(Graph, Subgraph, u, 0, make([]*list, len(Subgraph)), make(map[uint32]uint32))
				wg.Done()
			} (u)
		}
	}
	wg.Wait()
}

func RecursionSearch(Graph graph, Subgraph graph, v_g uint32, v_s uint32,
	restrictions []*list, path map[uint32]uint32) int {
	if _, ok := path[v_g]; ok {
		return 0
	}
	if len(Subgraph) == (len(path) + 1) {
		path[v_g] = v_s
		// file.WriteString(fmt.Sprintf("%v\n", path))
		// Verify_result(Graph,Subgraph,path)
		// fmt.Println("ye", path)
		delete(path, v_g)
		return 1
	}
	ret := 0
	path[v_g] = v_s
	inverse_restrictions, empty := UpdateRestrictions(Graph, Subgraph, v_g, v_s, restrictions)
	if !empty {
		targets := []uint32{}
		for u_instance := restrictions[v_s+1].start; u_instance != nil; u_instance = u_instance.next {
			targets = append(targets, u_instance.value)
		}
		for i := 0; i < len(targets); i++ {
			ret += RecursionSearch(Graph, Subgraph, targets[i], v_s+1, restrictions, path)
		}
	}
	for u := 0; u < len(inverse_restrictions); u++ {
		if inverse_restrictions[u] != nil {
			if inverse_restrictions[u].start != nil && inverse_restrictions[u].start.value == ^uint32(0) {
				restrictions[u] = nil
			} else {
				restrictions[u] = JoinLists(restrictions[u], inverse_restrictions[u])
			}
		}
	}
	delete(path, v_g)
	return ret
}

func UpdateRestrictions(G graph, S graph, v_g uint32, v_s uint32, restrictions []*list) ([]*list, bool) {
	empty := false
	inverse_restrictions := make([]*list, len(S))
	for u := range S[v_s].neighborhood {
		if restrictions[u] == nil {
			restrictions[u] = ColoredNeighborhood(G, v_g, S[u].attribute.color)
			el := element{^uint32(0), nil}
			inverse_restrictions[u] = &list{&el, &el}
		} else {
			var dis discriminator = func(u_instance uint32) bool {
				_, ok := G[v_g].neighborhood[u_instance]
				return ok
			}
			restrictions[u], inverse_restrictions[u] = SplitList(restrictions[u], dis)
			if restrictions[u].start == nil {
				empty = true
			}
		}
	}
	return inverse_restrictions, empty
}

func ColoredNeighborhood(Graph map[uint32]vertex, u uint32, c uint8) *list {
	output := list{nil, nil}
	for v := range Graph[u].neighborhood {
		if Graph[v].attribute.color == c {
			el := element{v, nil}
			ListAppend(&output, &el)
		}
	}
	return &output
}

func SplitList(l *list, which discriminator) (*list, *list) {
	l1 := &list{nil, nil}
	l2 := &list{nil, nil}
	var next *element
	for el := l.start; el != nil; el = next {
		next = el.next
		if which(el.value) {
			ListAppend(l1, el)
		} else {
			ListAppend(l2, el)
		}
	}
	return l1, l2
}

func JoinLists(l1 *list, l2 *list) *list {
	if l2 == nil {
		return l1
	}
	if l1 == nil {
		return l2
	}
	if l1.start == nil {
		return l2
	}
	if l2.start == nil {
		return l1
	}
	// fmt.Println(l1,l2)
	l1.end.next = l2.start
	l1.end = l2.end
	return l1
}

func ListAppend(l *list, el *element) {
	el.next = nil
	if l.start == nil {
		l.start = el
		l.end = el
		return
	}
	l.end.next = el
	l.end = el
}

func AddVertex(Graph graph, u uint32, c uint8) {
	if _, ok := Graph[u]; !ok {
		Graph[u] = vertex{neighborhood: make(map[uint32]void), attribute: att{color: c}}
	}
}

func AddEdge(Graph graph, u uint32, v uint32) {
	Graph[u].neighborhood[v] = void{}
	Graph[v].neighborhood[u] = void{}
}

func Gnp(n uint32, p float32) graph {
	Graph := make(graph)
	for i := uint32(0); i < n; i++ {
		color := rand.N(5)
		AddVertex(Graph, i, uint8(color))
	}

	for i := uint32(0); i < n; i++ {
		for j := i + 1; j < n; j++ {
			if rand.Float32() <= p {
				AddEdge(Graph, i, j)
			}
		}
	}
	return Graph
}
