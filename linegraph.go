package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"runtime/pprof"
	"time"
	"sync"
)

type void struct{}

type graph map[uint32]vertex

type fisset []*list

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

	for t := 0; t < 10; t++ {
		Gnp(1e2, 1e-1)
		Subgraph := Gnp(1e2, 1e-1) //for some reason with more edges it sometimes loses a match
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

		fmt.Println("starting search")

		file, err := os.Create(folderName + "/output" + fmt.Sprintf("%d", t) + ".txt")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		start_t := time.Now()
		find_all_subgraph(Subgraph, Subgraph, file)
		diff_t := time.Since(start_t)
		avg_time += diff_t.Seconds()
		fmt.Println("time", diff_t)
	}
	avg_time /= 50
	fmt.Println(avg_time)
}

func find_all_subgraph(Graph graph, Subgraph graph, file *os.File) {
	var wg sync.WaitGroup
	for v, k := range Graph {
		if k.attribute.color == Subgraph[0].attribute.color {
			wg.Add(1)
			func(v uint32) {
			concurrent_search_wrapper(Graph, Subgraph, v, file)
				wg.Done()
			}(v)
		}
	}
	wg.Wait()
	fmt.Println("done")
}

func concurrent_search_wrapper(Graph graph, Subgraph graph, v uint32, file *os.File) {
	recursion_search(Graph, Subgraph, v, 0, make(fisset, len(Subgraph)), make(map[uint32]uint32, 0), file)
}

func recursion_search(Graph graph, Subgraph graph, v_g uint32, v_s uint32,
	restrictions fisset, path map[uint32]uint32, file *os.File) {
	if _, ok := path[v_g]; ok {
		return
	}
	if len(Subgraph) == (len(path) + 1) {
		path[v_g] = v_s
		file.WriteString(fmt.Sprintf("%v\n", path))
		delete(path, v_g)
		return
	}
	path[v_g] = v_s

	deep_cpy := ListsArrDeepCopy(restrictions)

	func() {
		debug_rest := ListsArrDeepCopy(restrictions)
		inverse_restrictions1, _ := update_restrictions(Graph, Subgraph, v_g, v_s, debug_rest)
		cpy_test := ListsArrDeepCopy(inverse_restrictions1)
		reverse_restrictions(debug_rest, inverse_restrictions1)
		if !AreListsArrEqual(debug_rest, restrictions) {
			fmt.Println("bingo!")
		}
		inverse_restrictions2, _ := update_restrictions(Graph, Subgraph, v_g, v_s, debug_rest)
		if !AreListsArrEqual(inverse_restrictions2, cpy_test) {
			fmt.Println("bingo2!")
		}
		reverse_restrictions(debug_rest, inverse_restrictions2)
		if !AreListsArrEqual(debug_rest, restrictions) {
			fmt.Println("bingo3!")
		}
	}()

	inverse_restrictions, empty := update_restrictions(Graph, Subgraph, v_g, v_s, restrictions)
	defer reverse_restrictions(restrictions, inverse_restrictions)
	defer func() {
		if !AreListsArrEqual(deep_cpy, restrictions) {
			fmt.Println("err2")
		}
	}()

	if !empty {
		for u_instance := restrictions[v_s+1].start; u_instance != nil; u_instance = u_instance.next {
			recursion_search(Graph, Subgraph, u_instance.value, v_s+1, restrictions, path, file) // for some resone modifies the retrictions?! but doesnt reverse currectly
			if !AreListsArrEqual(deep_cpy, restrictions) {
				fmt.Println("err")
			}
		}
	}
	delete(path, v_g)
}

func reverse_restrictions(restrictions fisset, inverse_restrictions fisset) {
	for u, rest_u := range inverse_restrictions { //tested and works
		if rest_u != nil {
			if rest_u.start != nil && rest_u.start.value == ^uint32(0) {
				restrictions[u] = nil
			} else {
				join_lists(restrictions[u], rest_u)
			}
		}
	}
}

func update_restrictions(G graph, S graph, v_g uint32, v_s uint32, restrictions fisset) (fisset, bool) {
	empty := false
	inverse_restrictions := make(fisset, len(S))
	for u := range S[v_s].neighborhood {
		if restrictions[u] == nil {
			restrictions[u] = neighboorhood_with_color(G, v_g, S[u].attribute.color)
			el := element{^uint32(0), nil}
			inverse_restrictions[u] = &list{&el, &el}
		} else {
			restrictions[u], inverse_restrictions[u] = uninitialized_update(G, restrictions, u, v_g)
			if restrictions[u].start == nil {
				empty = true
			}
		}
	}
	return inverse_restrictions, empty
}

func uninitialized_update(G graph, restrictions fisset, u uint32, v_g uint32) (*list, *list) {
	temp_retriction := list{nil, nil}
	temp_inverse := list{nil, nil}
	var next *element
	for u_instance := restrictions[u].start; u_instance != nil; u_instance = next {
		next = u_instance.next
		if _, ok := G[v_g].neighborhood[u_instance.value]; ok {
			list_append(&temp_retriction, u_instance)
		} else {
			list_append(&temp_inverse, u_instance)
		}
	}
	if temp_retriction.end != nil {
		temp_retriction.end.next = nil
	}
	if temp_inverse.end != nil {
		temp_inverse.end.next = nil
	}

	return &temp_retriction, &temp_inverse
}

// modefies l1, l2 remain itself
func join_lists(l1 *list, l2 *list) {
	defer func() {
		l2 = &list{nil, nil}
		l2 = &list{nil, nil}
	}()
	if l2.end == nil && l2.start != nil {
		fmt.Println("what")
	}
	if l1.end == nil {
		l1.start = l2.start
		l1.end = l2.end
		return
	}
	l1.end.next = l2.start
	l1.end = l2.end
}

func PrintList(l *list) {
	if l == nil {
		fmt.Print("_")
		return
	}
	fmt.Print("[")
	for el := l.start; el != nil; el = el.next {
		fmt.Print(el.value, ", ")
	}
	fmt.Print("]")
}

func PrintListArr(arr []*list) {
	fmt.Print("[")
	for i := 0; i < len(arr); i++ {
		PrintList(arr[i])
	}
	fmt.Print("]\n")
}

func list_append(l *list, el *element) {
	el.next = nil
	if l.start == nil {
		l.start = el
		l.end = el
		return
	}
	l.end.next = el
	l.end = el
}

func neighboorhood_with_color(Graph graph, u uint32, c uint8) *list {
	output := list{nil, nil}
	for v := range Graph[u].neighborhood {
		if Graph[v].attribute.color == c {
			el := element{v,nil}
			list_append(&output, &el)
		}
	}
	return &output
}

// func print_graph(Graph graph) {
// 	for k, v := range Graph {
// 		fmt.Print("vertex ", k, " color ", v.attribute.color, " neighborhood : ")
// 		for t := range v.neighborhood {
// 			fmt.Print(t, " ")
// 		}
// 		fmt.Println("")
// 	}
// }

func add_vertex(Graph graph, u uint32, c uint8) {
	if _, ok := Graph[u]; !ok {
		Graph[u] = vertex{neighborhood: make(map[uint32]void), attribute: att{color: c}}
	}
}

func add_edge(Graph graph, u uint32, v uint32) {
	Graph[u].neighborhood[v] = void{}
	Graph[v].neighborhood[u] = void{}
}

func Gnp(n uint32, p float32) graph {
	Graph := make(graph)
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
func VerifyResult(G graph, S graph, path map[uint32]uint32) {
	for k, v := range path { //k is v_g, v is v_s
		if G[k].attribute.color != S[v].attribute.color {
			fmt.Println("wrong color", G[k], S[v])
		}
	}
	rev_path := ReverseMap(path)
	for k, v := range rev_path { //k is v_s, v is v_g
		for u := range S[k].neighborhood {
			if _, ok := G[v].neighborhood[rev_path[u]]; !ok {
				fmt.Println("missing edge", v, u)
			}
		}
	}
}

func AreListsArrEqual(arr1 []*list, arr2 []*list) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for i := 0; i < len(arr1); i++ {
		if !AreListsEqual(arr1[i], arr2[i]) {
			return false
		}
	}
	return true
}

func AreListsEqual(list1 *list, list2 *list) bool {
	if list1 == nil {
		return list2 == nil
	}
	if list1.start == nil {
		return list2.start == nil
	}
	map1 := make(map[uint32]bool)
	map2 := make(map[uint32]bool)
	for e := list1.start; e != nil; e = e.next {
		map1[e.value] = true
	}

	for e := list2.start; e != nil; e = e.next {
		map2[e.value] = true
	}

	if len(map1) != len(map2) {
		return false
	}

	for k := range map1 {
		if !map2[k] {
			return false
		}
	}

	return true
}

func ListsDeepCopy(l *list) *list {
	if l == nil {
		return nil
	}
	if l.start == nil {
		return &list{nil, nil}
	}
	out := list{nil, nil}
	for el := l.start; el != nil; el = el.next {
		cpy := element{el.value,nil}
		list_append(&out, &cpy)
	}
	return &out
}

func ListsArrDeepCopy(arr []*list) []*list {
	if arr == nil {
		return nil
	}
	new := make([]*list, len(arr))
	for i := 0; i < len(arr); i++ {
		new[i] = ListsDeepCopy(arr[i])
	}
	return new
}

func ReverseMap(m map[uint32]uint32) map[uint32]uint32 {
	n := make(map[uint32]uint32, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}
