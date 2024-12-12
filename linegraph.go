package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type void struct{}

type graph map[uint64]vertex

type vertex struct {
	attribute    att
	neighborhood map[uint64]void
}

type att struct {
	color uint32
}

type context struct {
	Graph        graph
	Subgraph     graph
	restrictions map[uint64]map[uint64]void
	path         map[uint64]uint64
	chosen       map[uint64]void
	prior        map[uint64]float32
	prior_policy int
}

type metric struct {
	time  time.Duration
	calls uint64
}

var recolor_policy *int
var output_file *os.File
var depth_file *os.File
var branching_file *os.File
var branching []float32
var logging_mu sync.Mutex
var branching_counter []float32
var calls uint64
var start_time time.Time
var depths map[uint64]metric

func main() {
	runtime.GOMAXPROCS(32)                        //regularization, keeps cpu under control
	debug.SetMaxStack(10 * 128 * 1024 * 1024)     //GBit
	debug.SetMemoryLimit(200 * 128 * 1024 * 1024) //GBit

	out_fname := flag.String("out", "dat/output.txt", "output location")
	input_fmt := flag.String("fmt", "json", "The file format to read\njson node-link,folder to .edges,.labels")
	input_parse := flag.String("parse", "%d\t%d", "The parse format of reading from file, used only for folder fmt")
	prior_policy := flag.Int("prior", 0, "the prior of the information we gain from vertex, based on our method S=0,G=1 or Constant=2, Random=3,Gready based on S=4")
	subset_size := flag.Int64("subset", -1, "take as subset of this size from G, to be the Subgraph")
	print_subset := flag.Bool("subout", false, "if subset, output it at that folder")
	recolor_policy = flag.Int("recolor", -1, "recolor value policy, defualt is base on read,else is rand.N")
	profile := flag.Bool("prof", false, "profile the program")
	depth_log := flag.String("depth", "", "fname to log the deapth over time")
	branching_log := flag.String("branching", "", "fname to  log the branching factor over time")

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

	if *depth_log != "" {
		depth_file, err = os.Create(*depth_log)
		if err != nil {
			panic(err)
		}
	}
	if *branching_log != "" {
		branching_file, err = os.Create(*branching_log)
		if err != nil {
			panic(err)
		}
	}

	gra_fname := flag.Args()[0]

	G := ReadGraph(gra_fname, *input_fmt, *input_parse)
	var S graph
	if *subset_size == -1 {
		sub_fname := flag.Args()[1]
		S = ReadGraph(sub_fname, *input_fmt, *input_parse)
	} else {
		S = reduceGraph(G, int(*subset_size))
		for int64(len(S)) < *subset_size { //in golang this is a while loop
			S = reduceGraph(G, int(*subset_size))
		}
		if *print_subset {
			printGraph(S, "dat/subgraphs/sub")
		}
	}
	fmt.Println(len(G), len(S))

	prior := make(map[uint64]float32)
	switch *prior_policy {
	case 0: //d^2 in S
		for v := range S {
			// prior[v] += float32(len(Subgraph[v].neighborhood))
			for u := range S[v].neighborhood {
				prior[v] += float32(len(S[u].neighborhood))
			}
		}
	case 1: //d^2 in G
		for v := range G {
			// prior[v] += float32(len(Graph[v].neighborhood))
			for u := range G[v].neighborhood {
				prior[v] += float32(len(G[u].neighborhood))
			}
		}
	case 4: //d in S
		for v := range S {
			prior[v] = float32(len(S[v].neighborhood))
		}
	}

	start := time.Now()
	var matches uint64 = FindAll(G, S, prior, *prior_policy)
	algo_time := time.Since(start)
	fmt.Println("done", algo_time.Seconds())
	fmt.Println("matches", matches)
}

func reduceGraph(Graph graph, size int) graph {
	subset := connectedComponentOfSizeK(Graph, randomVertex(Graph), size)
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

func ChooseNext[T any](restrictions map[uint64]map[uint64]T, chosen map[uint64]void, Subgraph graph, prior map[uint64]float32, prior_policy int) uint64 {
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
			score := RestrictionScore(restrictions, prior, u, prior_policy)
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

func ChooseStart(Subgraph graph, prior map[uint64]float32, prior_policy int) uint64 {
	if prior_policy == 3 || prior_policy == 2 || prior_policy == 1 {
		max_deg := 0
		idx := ^uint64(0)
		for u := range Subgraph {
			if len(Subgraph[u].neighborhood) > max_deg {
				max_deg = len(Subgraph[u].neighborhood)
				idx = u
			}
		}
		return idx
	}
	max_score := float32(0)
	idx := ^uint64(0)
	for u := range Subgraph {
		score := RestrictionScore[uint64](nil, prior, u, prior_policy)
		if score > max_score {
			max_score = score
			idx = u
		}
	}
	return idx
}

func RestrictionScore[T any](rest map[uint64]map[uint64]T, prior map[uint64]float32, u uint64, prior_policy int) float32 {
	switch prior_policy {
	case 0:
		return prior[u]
	case 1:
		score := float32(0)
		for u_instance := range rest[u] {
			score += prior[u_instance]
		}
		return 1 / score
	case 2:
		return 1024 / float32(len(rest[u]))
	case 3:
		return rand.Float32()
	case 4:
		return prior[u]
	}
	return 0
}

func FindAll(Graph graph, Subgraph graph, prior map[uint64]float32, prior_policy int) uint64 {
	var wg sync.WaitGroup
	var ops atomic.Uint64
	// t := time.Now()
	// f, err := os.Create("dat/" + t.Format("2006-01-02 15:04:05.999999") + ".txt")

	//debug stuff
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	branching = make([]float32, len(Subgraph))

	branching_counter = make([]float32, len(Subgraph))
	go func() {
		for sig := range c {
			// sig is a ^C (interrupt), handle it
			if sig == os.Interrupt || sig == syscall.SIGINT {
				fmt.Println("printing at interrupt")
				logging_mu.Lock()
				printDepths(depths, depth_file)
				printBranching(branching, branching_file)
				logging_mu.Unlock()
				pprof.StopCPUProfile()
				os.Exit(0)
			}
		}
	}()
	depths = make(map[uint64]metric)
	calls = 0
	start_time = time.Now()
	//functionality
	v_0 := ChooseStart(Subgraph, prior, prior_policy)
	for u := range Graph {
		if Graph[u].attribute.color == Subgraph[v_0].attribute.color && len(Graph[u].neighborhood) >= len(Subgraph[v_0].neighborhood) {
			wg.Add(1)
			context := context{Graph: Graph,
				Subgraph:     Subgraph,
				restrictions: make(map[uint64]map[uint64]void, len(Subgraph)),
				path:         make(map[uint64]uint64),
				chosen:       make(map[uint64]void),
				prior:        prior,
				prior_policy: prior_policy,
			}
			func(u uint64) {
				ret := RecursionSearch(&context, u, v_0)
				ops.Add(uint64(ret))
				wg.Done()
			}(u)
		}
		if ops.Load() > 0 {
			return ops.Load()
		}
		if u%512 == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
	printBranching(branching, branching_file)
	printDepths(depths, depth_file)
	return ops.Load()
}

func RecursionSearch(context *context, v_g uint64, v_s uint64) int {
	//debug
	logging_mu.Lock()
	calls++
	if depths != nil {
		if _, ok := depths[uint64(len(context.chosen))]; !ok {
			depths[uint64(len(context.chosen))] = metric{time.Since(start_time), calls}
		}
	}
	logging_mu.Unlock()
	// if len(context.chosen) == 8700 {
	// 	printGraph(graphSubset(context.Subgraph, context.chosen), "dat/subgraphs/partial_s")
	// 	var m map[uint64]void = make(map[uint64]void)
	// 	for _, v := range context.path {
	// 		m[v] = void{}
	// 	}
	// 	fmt.Println("about to print G")
	// 	printGraph(graphSubset(context.Graph, m), "dat/subgraphs/partial_g")
	// 	os.Exit(0)
	// }
	//functionality
	if _, ok := context.path[v_g]; ok {
		return 0
	}
	if len(context.Subgraph) == (len(context.path) + 1) {
		context.path[v_g] = v_s
		output_file.WriteString(fmt.Sprintf("%v\n", context.path))
		delete(context.path, v_g)
		return 1
	}
	ret := 0
	context.path[v_g] = v_s
	defer delete(context.path, v_g)
	context.chosen[v_s] = void{}
	defer delete(context.chosen, v_s)

	self_list := context.restrictions[v_s]
	delete(context.restrictions, v_s)
	inverse_restrictions, empty := UpdateRestrictions(context, v_g, v_s)
	inverse_restrictions[v_s] = self_list
	if !empty {
		new_v_s := ChooseNext(context.restrictions, context.chosen, context.Subgraph, context.prior, context.prior_policy)

		//debug
		fmt.Println("depth", len(context.chosen), "target size", len(context.restrictions[new_v_s]), "open", len(context.restrictions), "deg", len(context.Subgraph[new_v_s].neighborhood))
		if branching_file != nil {
			logging_mu.Lock()
			branching[len(context.chosen)] += float32(len(context.restrictions[new_v_s]))
			branching_counter[len(context.chosen)]++
			logging_mu.Unlock()
		}
		//functionality
		for u_instance := range context.restrictions[new_v_s] {
			ret += RecursionSearch(context, u_instance, new_v_s)
			// fmt.Println("ret ", ret)
			// if ret > 0 {
			// 	return ret
			// }
		}
		// ret += MinRestrictionsCall(Graph, Subgraph, restrictions, path, chosen)
	}
	for u := range inverse_restrictions {
		if _, ok := inverse_restrictions[u][^uint64(0)]; ok {
			delete(context.restrictions, u)
		} else {
			for u_instance := range inverse_restrictions[u] {
				if context.restrictions[u] == nil {
					context.restrictions[u] = make(map[uint64]void)
				}
				context.restrictions[u][u_instance] = void{}
			}
		}
	}
	return ret
}

func UpdateRestrictions(context *context, v_g uint64, v_s uint64) (map[uint64]map[uint64]void, bool) {
	empty := false
	inverse_restrictions := make(map[uint64]map[uint64]void)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for u := range context.Subgraph[v_s].neighborhood {
		wg.Add(1)
		go func() {
			if _, ok := context.chosen[u]; ok {
				wg.Done()
				return
			}
			mu.Lock()
			var single_rest map[uint64]void = context.restrictions[u]
			var single_inverse map[uint64]void = inverse_restrictions[u]
			mu.Unlock()

			SingleUpdate(context, u, v_g, &single_inverse, &single_rest)
			mu.Lock()
			context.restrictions[u] = single_rest
			inverse_restrictions[u] = single_inverse
			mu.Unlock()
			if len(single_rest) == 0 {
				empty = true
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return inverse_restrictions, empty
}
func SingleUpdate(context *context, u uint64, v_g uint64, single_inverse *map[uint64]void, single_rest *map[uint64]void) {
	if *single_rest == nil {
		*single_rest = ColoredNeighborhood(context.Graph, v_g, context.Subgraph[u].attribute.color, len(context.Subgraph[u].neighborhood))
		*single_inverse = make(map[uint64]void)
		(*single_inverse)[^uint64(0)] = void{}
	} else {
		// _, ok := G[v_g].neighborhood[u_instance]
		for u_instance := range *single_rest {
			if _, ok := context.Graph[v_g].neighborhood[u_instance]; !ok {
				if *single_inverse == nil {
					*single_inverse = make(map[uint64]void)
				}
				(*single_inverse)[u_instance] = void{}
				delete(*single_rest, u_instance)
			}
		}

	}
}

func ColoredNeighborhood(Graph graph, u uint64, c uint32, deg int) map[uint64]void {
	output := make(map[uint64]void)
	for v := range Graph[u].neighborhood {
		if Graph[v].attribute.color == c && len(Graph[v].neighborhood) >= deg {
			output[v] = void{}
		}
	}
	return output
}

func (Graph graph) AddVertex(u uint64, c uint32) {
	if _, ok := Graph[u]; ok {
		return
	}
	if *recolor_policy == -1 {
		Graph[u] = vertex{neighborhood: make(map[uint64]void), attribute: att{color: c}}
		return
	}
	Graph[u] = vertex{neighborhood: make(map[uint64]void), attribute: att{color: uint32(rand.N(*recolor_policy))}}
}

func (Graph graph) AddEdge(u uint64, v uint64) {
	if u == v {
		// fmt.Println("Ignores self loops")
		return
	}
	if _, ok := Graph[u]; !ok {
		Graph.AddVertex(u, ^uint32(0))
	}
	if _, ok := Graph[v]; !ok {
		Graph.AddVertex(v, ^uint32(0))
	}
	Graph[u].neighborhood[v] = void{}
	Graph[v].neighborhood[u] = void{}
}

func Gnp(n uint64, p float32) graph {
	Graph := make(graph)
	for i := uint64(0); i < n; i++ {
		color := rand.N(5)
		Graph.AddVertex(i, uint32(color))
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

func printDepths(depths map[uint64]metric, file *os.File) {
	file.WriteString("Depth,Time,Calls\n")
	for point := range depths { //print as csv instead!
		file.WriteString(fmt.Sprintf("%v,%v,%v\n", point, depths[point].time, depths[point].calls))
	}
}

func printBranching(branching []float32, file *os.File) {
	file.WriteString("Depth,BranchingFactor\n")
	for i := len(branching) - 1; i > 0; i-- { //print as csv instead!
		branching[i] /= branching_counter[i]
		file.WriteString(fmt.Sprintf("%v,%v\n", i, branching[i]))
	}
}

func randomVertex(Graph graph) uint64 {
	r := rand.IntN(len(Graph))
	for v := range Graph {
		if r == 0 {
			return v
		}
		r--
	}
	panic("no random")
}

func printGraph(G graph, dir string) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err)
	}
	sub_vertecies, err := os.Create(dir + "/sub.node_labels")
	if err != nil {
		panic(err)
	}
	for v := range G {
		fmt.Fprintln(sub_vertecies, v, G[v].attribute.color)
	}
	sub_vertecies.Close()
	sub_edges, err := os.Create(dir + "/sub.edges")
	if err != nil {
		panic(err)
	}
	for v := range G {
		for u := range G[v].neighborhood {
			if u >= v {
				fmt.Fprintln(sub_edges, v, u)
			}
		}
	}
	sub_edges.Close()
}
