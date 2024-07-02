package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
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

	f, err := os.Create("cpu.pprof")
	if err != nil {
		panic(err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sig := range c {
			// sig is a ^C (interrupt), handle it
			if sig == os.Interrupt {
				pprof.StopCPUProfile()
				os.Exit(0)
			}
		}
	}()

	// runtime.GOMAXPROCS(64)

	i := os.Args[1]
	j := os.Args[2]
	num_errors, err := strconv.Atoi(os.Args[3])
	if err != nil {
		// ... handle error
		panic(err)
	}

	G := ReadGraph(fmt.Sprintf("inputs/graph%v.json", i))
	S := ReadGraph(fmt.Sprintf("inputs/graph%v.json", j))
	fmt.Println(len(G), len(S))
	// ordering := ReadOrdering(fmt.Sprintf("inputs/ordering_%v_%v.json", i, j))
	start := time.Now()
	// FindAllSubgraphPathgraph(G, S, ordering, fmt.Sprintf("output%v_%v", i, j))
	IncompleteFindAll(G, S, 5, uint32(num_errors), fmt.Sprintf("output_%v_%v_withskips_%v", i, j, num_errors))
	algo_time := time.Since(start)
	fmt.Println("done", algo_time.Seconds())

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

func IncompleteFindAll(Graph graph, Subgraph graph, root uint32, threshold uint32, fname string) uint64 {
	var wg sync.WaitGroup
	var ops atomic.Uint64
	f, err := os.Create("dat/" + fname + ".txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// start_time := time.Now()
	for u := range Graph {
		if Graph[u].attribute.color == Subgraph[root].attribute.color {
			wg.Add(1)
			go func(u uint32) {
				ret := IncompleteRecusionSearch(Graph, Subgraph, u, root, make(map[uint32]map[uint32]uint32),
					make(map[uint32]uint32), make(map[uint32]void), threshold, f)

				// fmt.Println("done run",u,time.Since(start_time))
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

func IncompleteUpdateRestrictions(G graph, S graph, v_g uint32, v_s uint32, restrictions map[uint32]map[uint32]uint32,
	chosen map[uint32]void, threshold uint32) map[uint32]map[uint32]uint32 {

	inverse_restrictions := make(map[uint32]map[uint32]uint32)
	for u := range S[v_s].neighborhood {
		if _, ok := chosen[u]; !ok {
			if _, ok := restrictions[u]; !ok {
				//the restrictions is uninitialized
				restrictions[u] = PriorityColoredNeighborhood(G, v_g, S[u].attribute.color)
				inverse_restrictions[u] = make(map[uint32]uint32)
				inverse_restrictions[u][0] = ^uint32(0) //special value to denote restrictions is new
			} else {
				for u_instance := range restrictions[u] {
					if _, ok := G[v_g].neighborhood[u_instance]; !ok {
						if inverse_restrictions[u] == nil {
							inverse_restrictions[u] = make(map[uint32]uint32)
						}
						inverse_restrictions[u][u_instance] = restrictions[u][u_instance]
						restrictions[u][u_instance] += 1
						if restrictions[u][u_instance] > threshold {
							delete(restrictions[u], u_instance)
						}
					}
				}
			}
		}
	}
	return inverse_restrictions
}

func IncompleteRecusionSearch(Graph graph, Subgraph graph, v_g uint32, v_s uint32,
	restrictions map[uint32]map[uint32]uint32, path map[uint32]uint32,
	chosen map[uint32]void, threshold uint32, file *os.File) int {
	if _, ok := path[v_g]; ok {
		return 0
	}
	if len(Subgraph) == (len(chosen) + 1) {
		if v_g != ^uint32(0) {
			path[v_g] = v_s
			defer delete(path, v_g)
		}
		if file != nil {
			file.WriteString(fmt.Sprintf("%v\n", path))
		}
		return 1
	}
	ret := 0
	chosen[v_s] = void{}
	defer delete(chosen, v_s)

	self_rest := restrictions[v_s]
	delete(restrictions, v_s)

	if v_g != ^uint32(0) {
		path[v_g] = v_s
		defer delete(path, v_g)

		inverse_restrictions := IncompleteUpdateRestrictions(Graph, Subgraph, v_g, v_s, restrictions, chosen, threshold)
		defer IncompleteReverseRestrictions(restrictions, inverse_restrictions)
	}

	// temp_ordering := []uint32{5, 6, 7, 8, 9, 11, 2, 0, 1, 3, 4, 10, 12}
	// new_v_s := temp_ordering[len(chosen)]
	new_v_s := IncompleteChooseNext(restrictions)
	fmt.Println("depth", len(chosen), len(restrictions[new_v_s]), "skips", len(chosen)-len(path))
	for target := range restrictions[new_v_s] {
		if restrictions[new_v_s][target] <= threshold {
			IncompleteRecusionSearch(Graph, Subgraph, target, new_v_s, restrictions, path, chosen, threshold, file)
		}
	}
	//skip call!
	fmt.Println("skip call")
	skip_deg := uint32(len(Subgraph[new_v_s].neighborhood))
	if skip_deg <= threshold {
		IncompleteRecusionSearch(Graph, Subgraph, ^uint32(0), new_v_s, restrictions, path, chosen, threshold-skip_deg, file)
	}

	restrictions[v_s] = self_rest
	return ret
}

func IncompleteChooseNext(restrictions map[uint32]map[uint32]uint32) uint32 {
	//we want the max number of errors, but also min length
	min_score := float32(math.MaxFloat32)
	idx := uint32(0)
	for u := range restrictions {
		score := RestrictionScore(restrictions[u])
		if score < min_score {
			min_score = score
			idx = u
		}
	}
	return idx
}

func RestrictionScore(input map[uint32]uint32) float32 {
	// score := float32(0)
	// for u_instance := range input{
	// 	if input[u_instance] == 0 {
	// 		score += 1
	// 	} else {
	// 		score += (1 / float32(input[u_instance]))
	// 	}
	// }
	// return score
	return float32(len(input))
}

func IncompleteReverseRestrictions(restrictions map[uint32]map[uint32]uint32, inverse_rest map[uint32]map[uint32]uint32) {
	for u := range inverse_rest {
		if inverse_rest[u][0] == ^uint32(0) {
			delete(restrictions, u)
		} else {
			for u_instance := range inverse_rest[u] {
				restrictions[u][u_instance] = inverse_rest[u][u_instance]
			}
		}
	}
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
	for u := range Graph {
		if Graph[u].attribute.color == Subgraph[ordering[0]].attribute.color {
			wg.Add(1)
			go func(u uint32) {
				ret := RecursionSearch(Graph, Subgraph, u, ordering[0], make(map[uint32]*list, len(Subgraph)),
					make(map[uint32]uint32), f, ordering)

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
	restrictions map[uint32]*list, path map[uint32]uint32, file *os.File, ordering []uint32) int {
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
	self_list := restrictions[v_s]
	delete(restrictions, v_s)
	inverse_restrictions, empty := UpdateRestrictions(Graph, Subgraph, v_g, v_s, restrictions, path)
	inverse_restrictions[v_s] = self_list
	if !empty {
		// if true{
		if len(path) < 1 {
			targets := []uint32{}
			new_v_s := uint32(0)
			new_v_s = ordering[len(path)]
			for u_instance := restrictions[new_v_s].start; u_instance != nil; u_instance = u_instance.next {
				targets = append(targets, u_instance.value)
			}
			fmt.Println("targets size", len(targets), "death", len(path))
			for i := 0; i < len(targets); i++ {
				ret += RecursionSearch(Graph, Subgraph, targets[i], new_v_s, restrictions, path, file, ordering)
			}
		} else {
			ret += MinRestrictionsCall(Graph, Subgraph, restrictions, path, ordering, file)
		}
	}
	for u := range inverse_restrictions {
		if inverse_restrictions[u] != nil {
			if inverse_restrictions[u].start != nil && inverse_restrictions[u].start.value == ^uint32(0) {
				delete(restrictions, u)
			} else {
				restrictions[u] = JoinLists(restrictions[u], inverse_restrictions[u])
			}
		}
	}
	delete(path, v_g)
	return ret
}

func MinRestrictionsCall(Graph graph, Subgraph graph, restrictions map[uint32]*list,
	path map[uint32]uint32, ordering []uint32, file *os.File) int {
	ret := 0
	best_length := ^uint32(0)
	targets := []uint32{}
	new_v_s := uint32(0)
	for t := range restrictions {
		if restrictions[uint32(t)].length < best_length {
			new_v_s = uint32(t)
			best_length = restrictions[uint32(t)].length
		}
		if best_length == 1 {
			fmt.Println("targets size", best_length, "death", len(path), "vertex", new_v_s)
			ret += RecursionSearch(Graph, Subgraph, restrictions[new_v_s].start.value, new_v_s, restrictions, path, file, ordering)
			return ret
		}
	}
	for u_instance := restrictions[new_v_s].start; u_instance != nil; u_instance = u_instance.next {
		targets = append(targets, u_instance.value)
	}
	fmt.Println("targets size", len(targets), "death", len(path), "vertex", new_v_s)
	for i := 0; i < len(targets); i++ {
		ret += RecursionSearch(Graph, Subgraph, targets[i], new_v_s, restrictions, path, file, ordering)
	}
	return ret
}

func UpdateRestrictions(G graph, S graph, v_g uint32, v_s uint32,
	restrictions map[uint32]*list, path map[uint32]uint32) (map[uint32]*list, bool) {
	empty := false
	inverse_restrictions := make(map[uint32]*list, len(S))
	rev_path := reverseMap(path)
	for u := range S[v_s].neighborhood {
		if _, ok := rev_path[u]; !ok {
			if _, ok := restrictions[u]; !ok {
				restrictions[u] = ColoredNeighborhood(G, v_g, S[u].attribute.color)
				el := element{^uint32(0), nil}
				inverse_restrictions[u] = &list{&el, &el, 0}
			} else {
				var dis discriminator = func(u_instance uint32) bool {
					_, ok := G[v_g].neighborhood[u_instance]
					return ok
				}
				restrictions[u], inverse_restrictions[u] = SplitList(restrictions[u], dis)
			}
			if restrictions[u].length == 0 {
				empty = true
			}
		}
	}
	return inverse_restrictions, empty
}

func ColoredNeighborhood(Graph graph, u uint32, c uint16) *list {
	output := list{nil, nil, 0}
	for v := range Graph[u].neighborhood {
		if Graph[v].attribute.color == c {
			el := element{v, nil}
			ListAppend(&output, &el)
		}
	}
	return &output
}

func PriorityColoredNeighborhood(Graph graph, u uint32, c uint16) map[uint32]uint32 {
	output := make(map[uint32]uint32)
	for v := range Graph[u].neighborhood {
		if Graph[v].attribute.color == c {
			output[v] = 0
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

func reverseMap(m map[uint32]uint32) map[uint32]uint32 {
	n := make(map[uint32]uint32, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}
