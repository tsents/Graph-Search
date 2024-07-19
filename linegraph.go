package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"
)

type void struct{}

type discriminator func(uint64) bool

type graph map[uint64]vertex

type element struct {
	value uint64
	next  *element
}

type list struct {
	start  *element
	end    *element
	length uint64
}

type vertex struct {
	attribute    att
	neighborhood map[uint64]void
}

type att struct {
	color uint16
}

var prior_policy *int
var missing_policy *int
var start_point *int64
var output_file *os.File

// var next_function func(Subgraph graph,rest map[uint64]map[uint64]uint64,u uint64,prior uint64)uint64;

func main() {
	runtime.GOMAXPROCS(32) //regularization, keeps cpu under control

	out_fname := flag.String("out", "dat/output.txt", "output location")
	cmd_error := flag.Int("err", 0, "number of errors in the search\ndefault is exact isomorphism (default 0)")
	input_fmt := flag.String("fmt", "json", "The file format to read\njson node-link,folder to .edges,.labels")
	input_parse := flag.String("parse", "%d\t%d", "The parse format of reading from file, used only for folder fmt")
	prior_policy = flag.Int("prior", 0, "the prior of the information we gain from vertex, based on S=0,G=1 or Constant=2")
	subset_size := flag.Int64("subset", -1, "take as subset of this size from G, to be the Subgraph")
	missing_policy = flag.Int("missing", -1, "missing value policy, defualt is 'any color',else is rand.N")
	profile := flag.Bool("prof", false, "profile the program")
	start_point = flag.Int64("start", 0, "the starting point of the search")

	flag.Parse()

	if *profile {
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
	}

	fmt.Println("output ->", *out_fname)
	fmt.Println("parsing :", *input_parse)
	fmt.Println("prior :", *prior_policy)
	var err error
	output_file, err = os.Create(*out_fname)
	if err != nil {
		panic(err)
	}
	defer output_file.Close()

	gra_fname := flag.Args()[0]
	num_errors := *cmd_error

	G := ReadGraph(gra_fname, *input_fmt, *input_parse)
	var S graph
	if *subset_size == -1 {
		sub_fname := flag.Args()[1]
		S = ReadGraph(sub_fname, *input_fmt, *input_parse)
	} else {
		S = reduceGraph(G, int(*subset_size))
	}
	fmt.Println(len(G), len(S))
	// ordering := ReadOrdering(fmt.Sprintf("inputs/ordering_%v_%v.json", i, j))
	start := time.Now()
	// FindAllSubgraphPathgraph(G, S, ordering, fmt.Sprintf("output%v_%v", i, j))
	IncompleteFindAll(G, S, uint64(num_errors))
	algo_time := time.Since(start)
	fmt.Println("done", algo_time.Seconds())
}

// func colorDist(Graph graph) {
// 	bins := make(map[uint16]uint64)
// 	for _, v := range Graph {
// 		bins[v.attribute.color] += 1
// 	}
// 	fmt.Println(bins)
// }

func reduceGraph(Graph graph, size int) graph {
	m := make(map[uint64]void)
	for i := range Graph {
		m[i] = void{}
	}
	subset := connectedComponentOfSizeK(Graph, uint64(*start_point), size)
	return graphSubset(Graph, subset)
}

func sizeBFS(Graph graph, node uint64, visited map[uint64]void, component map[uint64]void, k *int) {
	queue := make(map[uint64]void)
	for neighbor := range Graph[node].neighborhood {
		if _, ok := visited[neighbor]; !ok {
			*k -= 1
			queue[neighbor] = void{}
			component[neighbor] = void{}
			visited[neighbor] = void{}
			if *k <= 0 {
				return
			}
		}
	}
	for neighbor := range queue {
		if *k > 0 {
			sizeBFS(Graph, neighbor, visited, component, k)
		}
	}
}

func connectedComponentOfSizeK(Graph graph, startNode uint64, k int) map[uint64]void {
	visited := make(map[uint64]void)
	component := make(map[uint64]void)
	visited[startNode] = void{}
	component[startNode] = void{}
	k -= 1
	sizeBFS(Graph, startNode, visited, component, &k)
	return component
}

func graphSubset(Graph graph, subset map[uint64]void) graph {
	cpy := make(graph)
	for v := range subset {
		new_neighborhood := make(map[uint64]void)
		for u := range Graph[v].neighborhood {
			if _, ok := subset[u]; ok {
				new_neighborhood[u] = void{}
			}
		}
		cpy[v] = vertex{Graph[v].attribute, new_neighborhood}
	}
	return cpy
}

func IncompleteFindAll(Graph graph, Subgraph graph, threshold uint64) {
	m := make(map[uint64]void)
	for v := range Subgraph {
		m[v] = void{}
	}

	prior := make(map[uint64]float32)
	switch *prior_policy {
	case 0: //d^2 in S
		for v := range Subgraph {
			// prior[v] += float32(len(Subgraph[v].neighborhood))
			for u := range Subgraph[v].neighborhood {
				prior[v] += float32(len(Subgraph[u].neighborhood))
			}
		}
	case 1: //d^2 in G
		for v := range Graph {
			// prior[v] += float32(len(Graph[v].neighborhood))
			for u := range Graph[v].neighborhood {
				prior[v] += float32(len(Graph[u].neighborhood))
			}
		}
	}
	IncompleteFindWithRoot(Graph, Subgraph, uint64(*start_point), threshold, make(map[uint64]void), prior)
	IncompleteCaller(Graph, Subgraph, uint64(*start_point), threshold, make(map[uint64]void), m, prior)
}

func IncompleteCaller(Graph graph, Subgraph graph, v_start uint64, threshold uint64, ignore map[uint64]void, componnent map[uint64]void, prior map[uint64]float32) {
	if uint64(len(Subgraph[v_start].neighborhood)) > threshold {
		return
	}
	delete(componnent, v_start)

	ignore[v_start] = void{}
	defer delete(ignore, v_start)

	threshold -= uint64(len(Subgraph[v_start].neighborhood))

	new_components := ConnectedComponents(Subgraph, componnent)

	for i := 0; i < len(new_components); i++ {
		if len(new_components[i]) > 0 {
			new_v := uint64(0)
			for v := range new_components[i] {
				new_v = v
				break
			}
			fmt.Println("call", new_v, threshold, ignore)

			IncompleteFindWithRoot(Graph, Subgraph, new_v, threshold, ignore, prior)

			IncompleteCaller(Graph, Subgraph, new_v, threshold, ignore, new_components[i], prior)
		}
	}
}

func IncompleteFindWithRoot(Graph graph, Subgraph graph, root uint64, threshold uint64, ignore map[uint64]void, prior map[uint64]float32) uint64 {
	var wg sync.WaitGroup
	var ops atomic.Uint64
	start_time := time.Now()
	for u := range Graph {
		if Graph[u].attribute.color == Subgraph[root].attribute.color ||
			Graph[u].attribute.color == ^uint16(0) || Subgraph[root].attribute.color == ^uint16(0) {
			wg.Add(1)
			go func(u uint64) {
				ret := IncompleteRecursionSearch(Graph, Subgraph, u, root, make(map[uint64]map[uint64]uint64),
					make(map[uint64]uint64), deepCopy(ignore), threshold, prior)

				// fmt.Println("done run", u, time.Since(start_time))
				ops.Add(uint64(ret))
				wg.Done()
			}(u)
		}
		if u%512 == 0 { // regulrization, keeps memory under control
			wg.Wait()
		}
	}
	wg.Wait()
	fmt.Println("done run", time.Since(start_time), ops.Load())
	return ops.Load()
}

func IncompleteUpdateRestrictions(G graph, S graph, v_g uint64, v_s uint64, restrictions map[uint64]map[uint64]uint64,
	chosen map[uint64]void, threshold uint64) map[uint64]map[uint64]uint64 {
	inverse_restrictions := make(map[uint64]map[uint64]uint64)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for u := range S[v_s].neighborhood {
		wg.Add(1)
		go func(u uint64) {
			if _, ok := chosen[u]; ok {
				wg.Done()
				return
			}
			mu.Lock()
			cur_rest := restrictions[u]
			inv_rest := inverse_restrictions[u]
			mu.Unlock()

			incompleteSingleUpdate(G, S, u, v_g, threshold, &(cur_rest), &(inv_rest))

			mu.Lock()
			restrictions[u] = cur_rest
			inverse_restrictions[u] = inv_rest
			mu.Unlock()
			wg.Done()
		}(u)
	}
	wg.Wait()
	return inverse_restrictions
}

func incompleteSingleUpdate(G graph, S graph, u uint64, v_g uint64, threshold uint64, rest *map[uint64]uint64, inv_rest *map[uint64]uint64) {
	if *rest == nil {
		//the restrictions is uninitialized
		*rest = PriorityColoredNeighborhood(G, v_g, S[u].attribute.color, len(S[u].neighborhood))
		*inv_rest = make(map[uint64]uint64)
		(*inv_rest)[0] = ^uint64(0) //special value to denote restrictions is new
		return
	}
	for u_instance := range *rest {
		if _, ok := G[v_g].neighborhood[u_instance]; !ok {
			if *inv_rest == nil {
				*inv_rest = make(map[uint64]uint64)
			}
			(*inv_rest)[u_instance] = (*rest)[u_instance]
			(*rest)[u_instance] += 1
			if (*rest)[u_instance] > threshold {
				delete(*rest, u_instance)
			}
		}
	}
}

func IncompleteRecursionSearch(Graph graph, Subgraph graph, v_g uint64, v_s uint64,
	restrictions map[uint64]map[uint64]uint64, path map[uint64]uint64,
	chosen map[uint64]void, threshold uint64, prior map[uint64]float32) int {
	if _, ok := path[v_g]; ok {
		return 0
	}
	if len(Subgraph) <= (len(chosen) + 1) {
		if v_g != ^uint64(0) {
			path[v_g] = v_s
			defer delete(path, v_g)
		}
		output_file.WriteString(fmt.Sprintf("%v\n", path))
		return 1
	}
	ret := 0
	chosen[v_s] = void{}
	defer delete(chosen, v_s)

	if v_g != ^uint64(0) {
		path[v_g] = v_s
		defer delete(path, v_g)

		inverse_restrictions := IncompleteUpdateRestrictions(Graph, Subgraph, v_g, v_s, restrictions, chosen, threshold)
		defer IncompleteReverseRestrictions(restrictions, inverse_restrictions)
		self_rest := restrictions[v_s]
		delete(restrictions, v_s)
		defer func() { restrictions[v_s] = self_rest }()
	}

	new_v_s := IncompleteChooseNext(restrictions, chosen, Subgraph, prior)
	fmt.Println("depth", len(chosen), "target size", len(restrictions[new_v_s]), "open", len(restrictions), "skips", len(chosen)-len(path))

	// cpy := deepCopy(restrictions[new_v_s])
	for target := range restrictions[new_v_s] {
		if restrictions[new_v_s][target] <= threshold {
			ret += IncompleteRecursionSearch(Graph, Subgraph, target, new_v_s, restrictions, path, chosen, threshold-restrictions[new_v_s][target], prior)
		}
	}
	//skip call!
	// fmt.Println("skip call")
	skip_deg := uint64(len(Subgraph[new_v_s].neighborhood))
	if skip_deg <= threshold {
		ret += IncompleteRecursionSearch(Graph, Subgraph, ^uint64(0), new_v_s, restrictions, path, chosen, threshold-skip_deg, prior)
	}

	return ret
}

func IncompleteChooseNext(restrictions map[uint64]map[uint64]uint64, chosen map[uint64]void, Subgraph graph, prior map[uint64]float32) uint64 {
	//we want the max number of errors, but also min length
	max_score := float32(0)
	idx := ^uint64(0)
	// for u := range restrictions {
	// 	if _, ok := chosen[u]; !ok {
	// 		if len(restrictions[u]) == 0 {
	// 			return u
	// 		}
	// 	}
	// }
	for u := range restrictions {
		if _, ok := chosen[u]; !ok {
			if len(restrictions[u]) <= 1 {
				return u
			}
			score := RestrictionScore(restrictions, prior, u, len(chosen))
			if score > max_score {
				max_score = score
				idx = u
			}
		}
	}
	if idx == ^uint64(0) {
		for v := range Subgraph {
			if _, ok := chosen[v]; !ok {
				return v
			}
		}
	}
	return idx
}

func RestrictionScore(rest map[uint64]map[uint64]uint64, prior map[uint64]float32, u uint64, depth int) float32 {
	switch *prior_policy {
	case 0:
		return prior[u]
	case 1:
		score := float32(0)
		for u_instance := range rest[u] {
			score += prior[u_instance]
		}
		return score
	case 2:
		return 1024 / float32(len(rest[u]))
	}
	return 0
}

func IncompleteReverseRestrictions(restrictions map[uint64]map[uint64]uint64, inverse_rest map[uint64]map[uint64]uint64) {
	for u := range inverse_rest {
		if inverse_rest[u][0] == ^uint64(0) {
			delete(restrictions, u)
		} else {
			for u_instance := range inverse_rest[u] {
				restrictions[u][u_instance] = inverse_rest[u][u_instance]
			}
		}
	}
}

func FindAllSubgraphPathgraph(Graph graph, Subgraph graph, ordering []uint64, fname string) uint64 {
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
			go func(u uint64) {
				ret := RecursionSearch(Graph, Subgraph, u, ordering[0], make(map[uint64]*list, len(Subgraph)),
					make(map[uint64]uint64), f, ordering)

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

func RecursionSearch(Graph graph, Subgraph graph, v_g uint64, v_s uint64,
	restrictions map[uint64]*list, path map[uint64]uint64, file *os.File, ordering []uint64) int {
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
			targets := []uint64{}
			new_v_s := uint64(0)
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
			if inverse_restrictions[u].start != nil && inverse_restrictions[u].start.value == ^uint64(0) {
				delete(restrictions, u)
			} else {
				restrictions[u] = JoinLists(restrictions[u], inverse_restrictions[u])
			}
		}
	}
	delete(path, v_g)
	return ret
}

func MinRestrictionsCall(Graph graph, Subgraph graph, restrictions map[uint64]*list,
	path map[uint64]uint64, ordering []uint64, file *os.File) int {
	ret := 0
	best_length := ^uint64(0)
	targets := []uint64{}
	new_v_s := uint64(0)
	for t := range restrictions {
		if restrictions[uint64(t)].length < best_length {
			new_v_s = uint64(t)
			best_length = restrictions[uint64(t)].length
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

func UpdateRestrictions(G graph, S graph, v_g uint64, v_s uint64,
	restrictions map[uint64]*list, path map[uint64]uint64) (map[uint64]*list, bool) {
	empty := false
	inverse_restrictions := make(map[uint64]*list, len(S))
	rev_path := reverseMap(path)
	for u := range S[v_s].neighborhood {
		if _, ok := rev_path[u]; !ok {
			if _, ok := restrictions[u]; !ok {
				restrictions[u] = ColoredNeighborhood(G, v_g, S[u].attribute.color)
				el := element{^uint64(0), nil}
				inverse_restrictions[u] = &list{&el, &el, 0}
			} else {
				var dis discriminator = func(u_instance uint64) bool {
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

func ColoredNeighborhood(Graph graph, u uint64, c uint16) *list {
	output := list{nil, nil, 0}
	for v := range Graph[u].neighborhood {
		if Graph[v].attribute.color == c || Graph[v].attribute.color == ^uint16(0) {
			el := element{v, nil}
			ListAppend(&output, &el)
		}
	}
	return &output
}

func PriorityColoredNeighborhood(Graph graph, u uint64, c uint16, deg int) map[uint64]uint64 {
	output := make(map[uint64]uint64, len(Graph[u].neighborhood))
	for v := range Graph[u].neighborhood {
		if len(Graph[v].neighborhood) >= deg {
			if Graph[v].attribute.color == c || Graph[v].attribute.color == ^uint16(0) {
				output[v] = 0
			}
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

func (Graph graph) AddVertex(u uint64, c uint16) {
	if _, ok := Graph[u]; !ok {
		Graph[u] = vertex{neighborhood: make(map[uint64]void), attribute: att{color: c}}
	}
}

func (Graph graph) AddEdge(u uint64, v uint64) {
	if u == v {
		// fmt.Println("Ignores self loops")
		return
	}
	if _, ok := Graph[u]; !ok {
		if *missing_policy == -1 {
			Graph.AddVertex(u, ^uint16(0))
		} else {
			Graph.AddVertex(u, uint16(rand.N(*missing_policy)))
		}
	}
	if _, ok := Graph[v]; !ok {
		if *missing_policy == -1 {
			Graph.AddVertex(v, ^uint16(0))
		} else {
			Graph.AddVertex(v, uint16(rand.N(*missing_policy)))
		}
	}
	Graph[u].neighborhood[v] = void{}
	Graph[v].neighborhood[u] = void{}
}

func Gnp(n uint64, p float32) graph {
	Graph := make(graph)
	for i := uint64(0); i < n; i++ {
		color := rand.N(5)
		Graph.AddVertex(i, uint16(color))
	}

	for i := uint64(0); i < n; i++ {
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

func reverseMap(m map[uint64]uint64) map[uint64]uint64 {
	n := make(map[uint64]uint64, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}

func deepCopy[T, V comparable](m map[T]V) map[T]V {
	cpy := make(map[T]V)
	for k, v := range m {
		cpy[k] = v
	}
	return cpy
}

func ComponentsDFS(Graph graph, v_g uint64, visited map[uint64]void, component map[uint64]void, subset map[uint64]void) {
	if _, ok := subset[v_g]; !ok {
		return
	}
	visited[v_g] = void{}
	component[v_g] = void{}

	for neighbor := range Graph[v_g].neighborhood {
		if _, ok := visited[neighbor]; !ok {
			ComponentsDFS(Graph, neighbor, visited, component, subset)
		}
	}
}

func ConnectedComponents(Graph graph, subset map[uint64]void) []map[uint64]void {
	visited := make(map[uint64]void)
	components := make([]map[uint64]void, 0)

	for v_g := range subset {
		if _, ok := visited[v_g]; !ok {
			component := make(map[uint64]void)
			ComponentsDFS(Graph, v_g, visited, component, subset)
			components = append(components, component)
		}
	}

	return components
}
