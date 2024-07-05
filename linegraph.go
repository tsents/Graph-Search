package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"sync"
	"sync/atomic"
	"time"
	// "runtime/pprof"
	// "os/signal"
)

type void struct{}

type discriminator func(uint32) bool

type graph map[uint32]vertex

type element struct {
	value uint32
	next  *element
}

type list struct {
	start  *element
	end    *element
	length uint32
}

type vertex struct {
	attribute    att
	neighborhood map[uint32]void
}

type att struct {
	color uint16
}

// idea - store Graph as map of vertex -> neighboorhood -> void (neighborhood is a set)
func main() {
	fmt.Println("hello world")

	// f, err := os.Create("cpu.pprof")
	// if err != nil {
	// 	panic(err)
	// }
	// if err := pprof.StartCPUProfile(f); err != nil {
	// 	panic(err)
	// }
	// defer pprof.StopCPUProfile()
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)

	// go func() {
	// 	for sig := range c {
	// 		// sig is a ^C (interrupt), handle it
	// 		if sig == os.Interrupt {
	// 			pprof.StopCPUProfile()
	// 			os.Exit(0)
	// 		}
	// 	}
	// }()

	// runtime.GOMAXPROCS(64)

	i := os.Args[1]
	j := os.Args[2]

	G := ReadGraph(fmt.Sprintf("inputs/graph%v.json", i))
	S := ReadGraph(fmt.Sprintf("inputs/graph%v.json", j))
	fmt.Println(len(G), len(S))
	ordering := ReadOrdering(fmt.Sprintf("inputs/ordering_%v_%v.json", i, j))
	start := time.Now()
	matches := FindAllSubgraphPathgraph(G, S, ordering, fmt.Sprintf("output%v_%v", i, j))
	algo_time := time.Since(start)
	fmt.Println("done", matches, algo_time.Seconds())

	// for t := 0; t < 10; t++ {
	// 	G := Gnp(1e3, 0.5)
	// 	S := Gnp(1e2, 0.5)
	// 	for j := uint32(0); j < uint32(len(S))-1; j++ {
	// 		S.AddEdge(j, j+1)
	// 	}
	// 	ordering := make([]uint32, len(S))
	// 	for i := 0; i < len(ordering); i++ {
	// 		ordering[i] = uint32(i)
	// 	}
	// 	start := time.Now()
	// 	FindAllSubgraphPathgraph(G, S, ordering)
	// 	algo_time := time.Since(start)
	// 	fmt.Println("done 1e3 1e2", algo_time.Seconds())
	// }
}

func FindAllSubgraphPathgraph(Graph graph, Subgraph graph, ordering []uint32, fname string) uint64 {
	var wg sync.WaitGroup
	var ops atomic.Uint64
	// t := time.Now()
	// f, err := os.Create("dat/" + t.Format("2006-01-02 15:04:05.999999") + ".txt")
	f, err := os.Create("dat/" + fname + ".txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	start_time := time.Now()
	for u := uint32(0); u < uint32(len(Graph)); u++ {
		if Graph[u].attribute.color == Subgraph[ordering[0]].attribute.color {
			wg.Add(1)
			go func(u uint32) {
				ret := RecursionSearch(Graph, Subgraph, u, ordering[0], make(map[uint32]map[uint32]void, len(Subgraph)),
					make(map[uint32]uint32),make(map[uint32]void), f, ordering)

				fmt.Println("done run", u, time.Since(start_time))
				ops.Add(uint64(ret))
				wg.Done()
			}(u)
		}
		if u%512 == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
	return ops.Load()
}

func RecursionSearch(Graph graph, Subgraph graph, v_g uint32, v_s uint32,
	restrictions map[uint32]map[uint32]void, path map[uint32]uint32,chosen map[uint32]void, file *os.File, ordering []uint32) int {
	if _, ok := path[v_g]; ok {
		return 0
	}
	if len(Subgraph) == (len(path) + 1) {
		path[v_g] = v_s
		if file != nil {
			file.WriteString(fmt.Sprintf("%v\n", path))
		}
		delete(path, v_g)
		return 1
	}
	ret := 0
	path[v_g] = v_s
	defer delete(path, v_g)
	chosen[v_s] = void{}
	defer delete(chosen,v_s)

	self_list := restrictions[v_s]
	delete(restrictions, v_s)
	inverse_restrictions, empty := UpdateRestrictions(Graph, Subgraph, v_g, v_s, restrictions, chosen)
	inverse_restrictions[v_s] = self_list
	if !empty {
		// if true{
		if len(path) < len(ordering) {
			targets := []uint32{}
			new_v_s := uint32(0)
			new_v_s = ordering[len(path)]
			for u_instance := range restrictions[new_v_s] {
				targets = append(targets, u_instance)
			}
			fmt.Println("targets size", len(targets), "depth", len(path))
			for i := 0; i < len(targets); i++ {
				ret += RecursionSearch(Graph, Subgraph, targets[i], new_v_s, restrictions, path,chosen, file, ordering)
			}
		} else {
			ret += MinRestrictionsCall(Graph, Subgraph, restrictions, path,chosen, ordering, file)
		}
	}
	for u := range inverse_restrictions {
		if _, ok := inverse_restrictions[u][^uint32(0)]; ok {
			delete(restrictions, u)
		} else {
			for u_instance := range inverse_restrictions[u]{
				if restrictions[u] == nil{
					restrictions[u] = make(map[uint32]void)
				}
				restrictions[u][u_instance] = void{}
			}
		}
	}
	return ret
}

func MinRestrictionsCall(Graph graph, Subgraph graph, restrictions map[uint32]map[uint32]void,
	path map[uint32]uint32,chosen map[uint32]void, ordering []uint32, file *os.File) int {
	ret := 0
	best_length := int(^uint32(0))
	targets := []uint32{}
	new_v_s := uint32(0)
	for t := range restrictions {
		if len(restrictions[t]) < best_length {
			new_v_s = t
			best_length = len(restrictions[t])
		}
		if best_length <= 1 {
			fmt.Println("targets size", best_length, "depth", len(path), "vertex", new_v_s,"short skip")
			new_v_g := uint32(0)
			for v := range restrictions[new_v_s]{
				new_v_g = v
			}
			ret += RecursionSearch(Graph, Subgraph,new_v_g, new_v_s, restrictions, path,chosen, file, ordering)
			return ret
		}
	}
	for u_instance := range restrictions[new_v_s] {
		targets = append(targets, u_instance)
	}
	fmt.Println("targets size", len(targets), "death", len(path), "vertex", new_v_s)
	for i := 0; i < len(targets); i++ {
		ret += RecursionSearch(Graph, Subgraph, targets[i], new_v_s, restrictions, path,chosen, file, ordering)
	}
	return ret
}

func UpdateRestrictions(G graph, S graph, v_g uint32, v_s uint32,
	restrictions map[uint32]map[uint32]void, chosen map[uint32]void) (map[uint32]map[uint32]void, bool) {
	empty := false
	inverse_restrictions := make(map[uint32]map[uint32]void)
	for u := range S[v_s].neighborhood {
		if _, ok := chosen[u]; !ok {
			if _, ok := restrictions[u]; !ok {
				restrictions[u] = ColoredNeighborhood(G, v_g, S[u].attribute.color)
				inverse_restrictions[u] = make(map[uint32]void)
				inverse_restrictions[u][^uint32(0)] = void{}
			} else {
					// _, ok := G[v_g].neighborhood[u_instance]
				for u_instance := range restrictions[u]{
					if _, ok := G[v_g].neighborhood[u_instance]; !ok{
						if inverse_restrictions[u] == nil{
							inverse_restrictions[u] = map[uint32]void{}
						}
						inverse_restrictions[u][u_instance] = void{}
						delete(restrictions[u],u_instance)
					}
				}				
				
			}
			if len(restrictions[u]) == 0 {
				empty = true
			}
		}
	}
	return inverse_restrictions, empty
}

func ColoredNeighborhood(Graph graph, u uint32, c uint16) map[uint32]void {
	output := make(map[uint32]void)
	for v := range Graph[u].neighborhood {
		if Graph[v].attribute.color == c {
			output[v] = void{}
		}
	}
	return output
}

func SplitList(l *list, which discriminator) (*list, *list) {
	l1 := &list{nil, nil, 0}
	l2 := &list{nil, nil, 0}
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
	l1.length += l2.length
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
	l.length += 1
	el.next = nil
	if l.start == nil {
		l.start = el
		l.end = el
		return
	}
	l.end.next = el
	l.end = el
}

func (Graph graph) AddVertex(u uint32, c uint16) {
	if _, ok := Graph[u]; !ok {
		Graph[u] = vertex{neighborhood: make(map[uint32]void), attribute: att{color: c}}
	}
}

func (Graph graph) AddEdge(u uint32, v uint32) {
	Graph[u].neighborhood[v] = void{}
	Graph[v].neighborhood[u] = void{}
}

func Gnp(n uint32, p float32) graph {
	Graph := make(graph)
	for i := uint32(0); i < n; i++ {
		color := rand.N(5)
		Graph.AddVertex(i, uint16(color))
	}

	for i := uint32(0); i < n; i++ {
		for j := i + 1; j < n; j++ {
			if rand.Float32() <= p {
				Graph.AddEdge(i, j)
			}
		}
	}
	return Graph
}

func PrintList(l *list) {
	for u_instance := l.start; u_instance != nil; u_instance = u_instance.next {
		fmt.Print(u_instance.value, u_instance.next, u_instance, ',')
	}
	fmt.Print("\tlength", l.length)
	fmt.Println()
}

// func reverseMap(m map[uint32]uint32) map[uint32]uint32 {
// 	n := make(map[uint32]uint32, len(m))
// 	for k, v := range m {
// 		n[v] = k
// 	}
// 	return n
// }
