package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"runtime/pprof"
	"sync"
	"time"
)

type void struct{}

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

var special *list

// idea - store Graph as map of vertex -> neighboorhood -> void (neighborhood is a set)
func main() {
	fmt.Println("hello world")

	currentTime := time.Now()
	folderName := currentTime.Format("2006-01-02 15:04:05")
	folderName = "dat/" + folderName
	err := os.Mkdir(folderName, 0755)
	if err != nil {
		fmt.Println(err)
		return
	}
	avg_time := 0.0

	f, err := os.Create("myprog.prof")
	if err != nil {

		fmt.Println(err)
		return

	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	for t := 0; t < 50; t++ {
		Gnp(1e2, 1e-1)
		Subgraph := Gnp(1e2, 0) //for some reason with more edges it sometimes loses a match
		for i := uint32(0); i < 99; i++ {
			add_edge(Subgraph, i, i+1)
		}

		//plant the graph for testing
		// for idx, v := range Subgraph {
		// 	v_new := vertex{v.attribute,Graph[idx].neighborhood}
		// 	Graph[idx] = v_new
		// 	for u := range v.neighborhood{
		// 		add_edge(Graph,idx,u)
		// 	}
		// }

		l := list{nil, nil}
		special = &l
		fmt.Println("starting search")

		file, err := os.Create(folderName + "/output" + fmt.Sprintf("%d", t) + ".txt")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		start_t := time.Now()
		find_all_subgraph(Subgraph, Subgraph, file)
		diff_t := time.Now().Sub(start_t)
		avg_time += diff_t.Seconds()
		fmt.Println("time", diff_t)
	}
	avg_time /= 50
	fmt.Println(avg_time)
}

func find_all_subgraph(Graph map[uint32]vertex, Subgraph map[uint32]vertex, file *os.File) {
	var wg sync.WaitGroup
	for v, k := range Graph {
		if k.attribute.color == Subgraph[0].attribute.color {
			wg.Add(1)
			func(v uint32) {
				concurrent_search_wrapper(Graph, Subgraph, v, file)
				defer wg.Done()
			}(v)
		}
	}
	wg.Wait()
	fmt.Println("done")
}

func concurrent_search_wrapper(Graph map[uint32]vertex, Subgraph map[uint32]vertex, v uint32, file *os.File) {
	recursion_search(Graph, Subgraph, v, 0, make([]*list, len(Subgraph)), make(map[uint32]uint32, 0), file)
}

func recursion_search(Graph map[uint32]vertex, Subgraph map[uint32]vertex, v_g uint32, v_s uint32,
	restrictions []*list, path map[uint32]uint32, file *os.File) {
	if _, ok := path[v_g]; ok {
		return
	}
	if len(Subgraph) == (len(path) + 1) {
		path[v_g] = v_s
		file.WriteString(fmt.Sprintf("%v\n", path))
		// Verify_result(Graph,Subgraph,path)
		delete(path, v_g)
		return
	}
	path[v_g] = v_s
	inverse_restrictions, empty := update_restrictions(Graph, Subgraph, v_g, v_s, restrictions)
	if !empty {
		fmt.Println("block")
		for u_instance := restrictions[v_s+1].start; u_instance != nil; u_instance = u_instance.next {
			recursion_search(Graph, Subgraph, u_instance.value, v_s+1, restrictions, path, file)
			fmt.Println("block2")
			for i :=0; i < 10; i++ {
				print_list(restrictions[i])
			}
		}
	}
	for u, rest_u := range inverse_restrictions { //tested and works
		// fmt.Println("block")
		// print_list(restrictions[u])
		// print_list(inverse_restrictions[u])

		if rest_u != nil {
			if rest_u == special {
				restrictions[u] = nil
			} else {
				join_lists(restrictions[u], rest_u)
			}
		}
		// print_list(restrictions[u])
	}
	delete(path, v_g)
}

func update_restrictions(G map[uint32]vertex, S map[uint32]vertex, v_g uint32, v_s uint32, restrictions []*list) ([]*list, bool) {
	empty := false
	inverse_restrictions := make([]*list, len(S))
	for u := range S[v_s].neighborhood {
		if restrictions[u] == nil {
			restrictions[u] = neighboorhood_with_color(G, v_g, S[u].attribute.color)
			inverse_restrictions[u] = special
		} else {
			temp_inverse := list{nil, nil}
			temp_retriction := list{nil, nil}
			for u_instance := restrictions[u].start; u_instance != nil; u_instance = u_instance.next {
				if _, ok := G[v_g].neighborhood[u_instance.value]; ok {
					list_append(&temp_retriction, u_instance.value)
				} else {
					list_append(&temp_inverse, u_instance.value)
				}
			}
			restrictions[u] = &temp_retriction
			inverse_restrictions[u] = &temp_inverse
			if temp_retriction.start == nil {
				empty = true
			}
		}
	}
	return inverse_restrictions, empty
}

// modefies both! should only use l1 after
func join_lists(l1 *list, l2 *list) {
	if l1.end == nil {
		l1.start = l2.start
		l1.end = l2.end
		return
	}
	l1.end.next = l2.start
	l1.end = l2.end
}

func print_list(l *list) {
	if l == nil {
		fmt.Println("null list pointer")
		return
	}
	fmt.Print("[")
	for el := l.start; el != nil; el = el.next {
		fmt.Print(el.value, ", ")
	}
	fmt.Print("] \n")
}

func list_append(l *list, val uint32) {
	el := &element{val, nil}
	if l.start == nil {
		l.start = el
		l.end = el
		return
	}
	l.end.next = el
	l.end = el
}

func neighboorhood_with_color(Graph map[uint32]vertex, u uint32, c uint8) *list {
	output := list{nil, nil}
	for v := range Graph[u].neighborhood {
		if Graph[v].attribute.color == c {
			list_append(&output, v)
		}
	}
	return &output
}

func print_graph(Graph map[uint32]vertex) {
	for k, v := range Graph {
		fmt.Print("vertex ", k, " color ", v.attribute.color, " neighborhood : ")
		for t := range v.neighborhood {
			fmt.Print(t, " ")
		}
		fmt.Println("")
	}
}

func add_vertex(Graph map[uint32]vertex, u uint32, c uint8) {
	if _, ok := Graph[u]; !ok {
		Graph[u] = vertex{neighborhood: make(map[uint32]void), attribute: att{color: c}}
	}
}

func add_edge(Graph map[uint32]vertex, u uint32, v uint32) {
	Graph[u].neighborhood[v] = void{}
	Graph[v].neighborhood[u] = void{}
}

func Gnp(n uint32, p float32) map[uint32]vertex {
	Graph := make(map[uint32]vertex)
	for i := uint32(0); i < n; i++ {
		color := rand.N(5)
		add_vertex(Graph, i, uint8(color))
	}

	for i := uint32(0); i < n; i++ {
		for j := i + 1; j < n; j++ {
			if rand.Float32() <= p {
				add_edge(Graph, i, j)
			}
		}
	}
	return Graph
}

// debug functions
func Verify_result(G map[uint32]vertex, S map[uint32]vertex, path map[uint32]uint32) {
	for k, v := range path { //k is v_g, v is v_s
		if G[k].attribute.color != S[v].attribute.color {
			fmt.Println("wrong color", G[k], S[v])
		}
	}
	rev_path := reverseMap(path)
	for k, v := range rev_path { //k is v_s, v is v_g
		for u := range S[k].neighborhood {
			if _, ok := G[v].neighborhood[rev_path[u]]; !ok {
				fmt.Println("missing edge", v, u)
			}
		}
	}
}

func reverseMap(m map[uint32]uint32) map[uint32]uint32 {
	n := make(map[uint32]uint32, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}
