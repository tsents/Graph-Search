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
	calls        uint64
	start_time   time.Time
	depths       map[uint64]metric
}

type metric struct {
	time  time.Duration
	calls uint64
}

var prior_policy *int
var recolor_policy *int
var start_point *int64
var output_file *os.File
var depth_file *os.File
var branching_file *os.File
var branching []float32
var branching_counter []float32
var branching_mu sync.Mutex
var hair_counter []uint64
var hair_file *os.File
var crit_log *int
var crit_file *os.File
var deg_file *os.File

func main() {
	runtime.GOMAXPROCS(32)                        //regularization, keeps cpu under control
	debug.SetMaxStack(10 * 128 * 1024 * 1024)     //GBit
	debug.SetMemoryLimit(200 * 128 * 1024 * 1024) //GBit

	out_fname := flag.String("out", "dat/output.txt", "output location")
	cmd_error := flag.Int("err", 0, "number of errors in the search\ndefault is exact isomorphism (default 0)")
	input_fmt := flag.String("fmt", "json", "The file format to read\njson node-link,folder to .edges,.labels")
	input_parse := flag.String("parse", "%d\t%d", "The parse format of reading from file, used only for folder fmt")
	prior_policy = flag.Int("prior", 0, "the prior of the information we gain from vertex, based on our method S=0,G=1 or Constant=2, Random=3,Gready based on S=4")
	subset_size := flag.Int64("subset", -1, "take as subset of this size from G, to be the Subgraph")
	print_subset := flag.Bool("subout", false, "if subset, output it at that folder")
	recolor_policy = flag.Int("recolor", -1, "recolor value policy, defualt is base on read,else is rand.N")
	profile := flag.Bool("prof", false, "profile the program")
	start_point = flag.Int64("start", 1, "the starting point of the search")
	depth_log := flag.String("depth", "", "fname to log the deapth over time")
	branching_log := flag.String("branching", "", "fname to  log the branching factor over time")
	crit_log = flag.Int("crit", -1, "log the degrees at given critical point in the algorithm")
	crit_fname := flag.String("critfile", "", "fname for the file for the crit option")
	hair_log := flag.String("hair", "", "fname to log the number of vertecies with small degree that are left to choose vs algorithm depth")
	degs_log := flag.String("deg", "", "fname to log the degree distribution")
	sparse := flag.Float64("sparse", 0, "sparsify the subgraph")

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

	if *depth_log != "" {
		depth_file, err = os.Create(*depth_log)
		if err != nil {
			panic(err)
		}
	}

	if *crit_fname != "" {
		crit_file, err = os.Create(*crit_fname)
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

	if *hair_log != "" {
		hair_file, err = os.Create(*hair_log)
		if err != nil {
			panic(err)
		}
	}
	if *degs_log != "" {
		deg_file, err = os.Create(*degs_log)
		if err != nil {
			panic(err)
		}
	}

	gra_fname := flag.Args()[0]
	num_errors := *cmd_error

	G := ReadGraph(gra_fname, *input_fmt, *input_parse)
	var S graph
	if *subset_size == -1 {
		sub_fname := flag.Args()[1]
		S = ReadGraph(sub_fname, *input_fmt, *input_parse)
	} else {
		S = reduceGraph(G, int(*subset_size))
		if int64(len(S)) < *subset_size {
			m := make(map[uint64]void)
			for v := range G {
				m[v] = void{}
			}
			var componnents []map[uint64]void = ConnectedComponents(G, m) //array of sets of vertecies

			largest := len(componnents[0])
			largest_component := componnents[0]
			for i := 1; i < len(componnents); i++ {
				fmt.Println("componnentsize", len(componnents[i]))
				if len(componnents[i]) > largest {
					largest = len(componnents[i])
					largest_component = componnents[i]
				}
			}
			for idx := range largest_component {
				*start_point = int64(idx)
				break
			}
			S = reduceGraph(G, int(*subset_size))
		}
		S = Sparsify(S, float32(*sparse))
		if *print_subset {
			if err := os.MkdirAll(fmt.Sprintf("dat/subgraphs/subs%d", *start_point), os.ModePerm); err != nil {
				panic(err)
			}
			sub_vertecies, err := os.Create(fmt.Sprintf("dat/subgraphs/subs%d/sub.node_labels", *start_point))
			if err != nil {
				panic(err)
			}
			for v := range S {
				fmt.Fprintln(sub_vertecies, v, S[v].attribute.color)
			}
			sub_vertecies.Close()
			sub_edges, err := os.Create(fmt.Sprintf("dat/subgraphs/subs%d/sub.edges", *start_point))
			if err != nil {
				panic(err)
			}
			for v := range S {
				for u := range S[v].neighborhood {
					if u >= v {
						fmt.Fprintln(sub_edges, v, u)
					}
				}
			}
			sub_edges.Close()
		}
	}
	fmt.Println(len(G), len(S))
	colorDist(G)
	colorDist(S)

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
	// FindAllSubgraphPathgraph(G, S, ordering, fmt.Sprintf("output%v_%v", i, j))

	// if num_errors != 0{
	// 	IncompleteFindAll(G, S, uint64(num_errors))
	// } else {
	// 	FindAllSubgraphPathgraph(G,S,[]uint64{uint64(*start_point)},"bla")
	// }
	var matches uint64
	if num_errors == 0 {
		matches = FindAll(G, S, prior)
	} else {
		matches = IncompleteFindAll(G, S, uint64(num_errors), prior)
	}
	algo_time := time.Since(start)
	fmt.Println("done", algo_time.Seconds())
	fmt.Println("matches", matches)
}

func colorDist(Graph graph) {
	bins := make(map[uint32]uint64)
	for _, v := range Graph {
		bins[v.attribute.color] += 1
	}
	fmt.Println(bins)
}

func degDist(Graph graph, inv_subset map[uint64]void, file *os.File) {
	if inv_subset == nil {
		bins := make(map[uint64]uint64)
		for _, v := range Graph {
			bins[uint64(len(v.neighborhood))] += 1
		}
		file.WriteString(fmt.Sprintf("%v\n", bins))
		return
	}
	bins := make(map[uint64]uint64)
	for v := range Graph {
		if _, ok := inv_subset[v]; !ok {
			bins[uint64(len(Graph[v].neighborhood))] += 1
		}
	}
	file.WriteString(fmt.Sprintf("%v\n", bins))
}

func reduceGraph(Graph graph, size int) graph {
	subset := connectedComponentOfSizeK(Graph, uint64(*start_point), size)
	return graphSubset(Graph, subset)
}

func Sparsify(G graph, p float32) graph {
	total := 0
	for v := range G {
		total += len(G[v].neighborhood)
	}
	fmt.Println("S total", total)

	mst := minimumSpanningTree(G, uint64(*start_point))

	for u := range G {
		for v := range G[u].neighborhood {
			if u > v && rand.Float32() > p { //keep edge at prob 1-p
				mst.AddEdge(u, v)
			}
		}
	}

	total = 0
	for v := range mst {
		total += len(mst[v].neighborhood)
	}
	fmt.Println("sparse total", total)

	return mst
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

func IncompleteFindAll(Graph graph, Subgraph graph, threshold uint64, prior map[uint64]float32) uint64 {
	m := make(map[uint64]void)
	for v := range Subgraph {
		m[v] = void{}
	}
	var ret uint64 = 0
	ret += IncompleteFindWithRoot(Graph, Subgraph, uint64(*start_point), threshold, make(map[uint64]void), prior)
	ret += IncompleteCaller(Graph, Subgraph, uint64(*start_point), threshold, make(map[uint64]void), m, prior)
	return ret
}

func IncompleteCaller(Graph graph, Subgraph graph, v_start uint64, threshold uint64, ignore map[uint64]void, componnent map[uint64]void, prior map[uint64]float32) uint64 {
	if uint64(len(Subgraph[v_start].neighborhood)) > threshold {
		return 0
	}
	delete(componnent, v_start)
	var ret uint64 = 0
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

			ret += IncompleteFindWithRoot(Graph, Subgraph, new_v, threshold, ignore, prior)

			ret += IncompleteCaller(Graph, Subgraph, new_v, threshold, ignore, new_components[i], prior)
		}
	}
	return ret
}

func IncompleteFindWithRoot(Graph graph, Subgraph graph, root uint64, threshold uint64, ignore map[uint64]void, prior map[uint64]float32) uint64 {
	var wg sync.WaitGroup
	var ops atomic.Uint64
	start_time := time.Now()
	for u := range Graph {
		if Graph[u].attribute.color == Subgraph[root].attribute.color ||
			Graph[u].attribute.color == ^uint32(0) || Subgraph[root].attribute.color == ^uint32(0) {
			wg.Add(1)
			go func(u uint64) {
				var depths map[uint64]uint64 = nil
				if depth_file != nil {
					depths = make(map[uint64]uint64)
				}
				fmt.Println("start")
				ret := IncompleteRecursionSearch(Graph, Subgraph, u, root, make(map[uint64]map[uint64]uint64),
					make(map[uint64]uint64), deepCopy(ignore), threshold, prior, 0, depths)

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
		*rest = PriorityColoredNeighborhood(G, v_g, S[u].attribute.color, len(S[u].neighborhood)-int(threshold))
		// when doing the degree constraint, we need the degree to be above d_u - threshold, because up to K edges may be gone
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

// Graph,Subgraph,v_g,v_s,restrictions,path,chosen,threshold and prior are the context
// calls and depths are debug only
func IncompleteRecursionSearch(Graph graph, Subgraph graph, v_g uint64, v_s uint64,
	restrictions map[uint64]map[uint64]uint64, path map[uint64]uint64,
	chosen map[uint64]void, threshold uint64, prior map[uint64]float32, calls uint64, depths map[uint64]uint64) int {
	if depths != nil {
		calls += 1
		if _, ok := depths[uint64(len(chosen))]; !ok {
			depths[uint64(len(chosen))] = calls
		}
		if len(Subgraph) <= (len(chosen) + 1) {
			depth_file.WriteString(fmt.Sprintf("%v\n", depths))
			depths = nil
		}
	}

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

	new_v_s := ChooseNext(restrictions, chosen, Subgraph, prior)
	fmt.Println("depth", len(chosen), "target size", len(restrictions[new_v_s]), "open", len(restrictions), "skips", len(chosen)-len(path))

	// cpy := deepCopy(restrictions[new_v_s])
	for target := range restrictions[new_v_s] {
		if restrictions[new_v_s][target] <= threshold {
			ret += IncompleteRecursionSearch(Graph, Subgraph, target, new_v_s, restrictions, path, chosen, threshold-restrictions[new_v_s][target], prior, calls+1, depths)
		}
	}
	//skip call!
	// fmt.Println("skip call")
	skip_deg := uint64(len(Subgraph[new_v_s].neighborhood))
	if skip_deg <= threshold {
		ret += IncompleteRecursionSearch(Graph, Subgraph, ^uint64(0), new_v_s, restrictions, path, chosen, threshold-skip_deg, prior, calls+1, depths)
	}

	return ret
}

func ChooseNext[T any](restrictions map[uint64]map[uint64]T, chosen map[uint64]void, Subgraph graph, prior map[uint64]float32) uint64 {
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
			score := RestrictionScore(restrictions, prior, u)
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

func ChooseStart(Subgraph graph, prior map[uint64]float32) uint64 {
	if *prior_policy == 2 || *prior_policy == 1 {
		for idx := range Subgraph {
			return idx //arbotery
		}
	}
	max_score := float32(0)
	idx := uint64(*start_point)
	for u := range Subgraph {
		score := RestrictionScore[uint64](nil, prior, u)
		if score > max_score {
			max_score = score
			idx = u
		}
	}
	return idx
}

func RestrictionScore[T any](rest map[uint64]map[uint64]T, prior map[uint64]float32, u uint64) float32 {
	switch *prior_policy {
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

func FindAll(Graph graph, Subgraph graph, prior map[uint64]float32) uint64 {
	var wg sync.WaitGroup
	var ops atomic.Uint64
	// t := time.Now()
	// f, err := os.Create("dat/" + t.Format("2006-01-02 15:04:05.999999") + ".txt")

	//debug stuff
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	to_track := []*map[uint64]metric{}
	branching = make([]float32, len(Subgraph))
	if hair_file == nil {
		hair_counter = nil
	} else {
		hair_counter = make([]uint64, len(Subgraph))
	}
	branching_counter = make([]float32, len(Subgraph))
	go func() {
		for sig := range c {
			// sig is a ^C (interrupt), handle it
			if sig == os.Interrupt || sig == syscall.SIGINT {
				time.Sleep(2000000000)
				for i := 0; i < len(to_track); i++ {
					printDepths(*to_track[i], depth_file)
				}
				fmt.Println("printing at interrupt")
				printBranching(branching, branching_file)
				hair_file.WriteString(fmt.Sprintf("%v\n", hair_counter))
				pprof.StopCPUProfile()
				os.Exit(0)
			}
		}
	}()

	//functionality
	v_0 := ChooseStart(Subgraph, prior)
	for u := range Graph {
		if Graph[u].attribute.color == Subgraph[v_0].attribute.color {
			wg.Add(1)
			depths := make(map[uint64]metric)
			to_track = append(to_track, &depths)
			context := context{Graph: Graph,
				Subgraph:     Subgraph,
				restrictions: make(map[uint64]map[uint64]void, len(Subgraph)),
				path:         make(map[uint64]uint64),
				chosen:       make(map[uint64]void),
				prior:        prior,
				calls:        0,
				start_time:   time.Now(),
				depths:       depths}
			go func(u uint64) {
				ret := RecursionSearch(&context, u, v_0)
				ops.Add(uint64(ret))
				wg.Done()
			}(u)
		}
		if u%512 == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
	printBranching(branching, branching_file)
	hair_file.WriteString(fmt.Sprintf("%v\n", hair_counter))
	return ops.Load()
}

func RecursionSearch(context *context, v_g uint64, v_s uint64) int {
	//debug
	context.calls++
	if context.depths != nil {
		if _, ok := context.depths[uint64(len(context.chosen))]; !ok {
			context.depths[uint64(len(context.chosen))] = metric{time.Since(context.start_time), context.calls}
		}
		if len(context.Subgraph) == (len(context.path) + 1) {
			printDepths(context.depths, depth_file)
			// depth_file.WriteString(fmt.Sprintf("%v\n", context.depths))
			context.depths = nil
		}
	}
	if len(context.chosen) == *crit_log {
		degDist(context.Subgraph, context.chosen, crit_file)
		*crit_log = -1
	}

	if hair_counter != nil && hair_counter[len(context.chosen)] == 0 {
		degDist(context.Subgraph, context.chosen, deg_file)
		for v := range context.chosen {
			if len(context.Subgraph[v].neighborhood) < 4 {
				hair_counter[len(context.chosen)]++
			}
		}
	}

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
		new_v_s := ChooseNext(context.restrictions, context.chosen, context.Subgraph, context.prior)

		//debug
		fmt.Println("depth", len(context.chosen), "target size", len(context.restrictions[new_v_s]), "open", len(context.restrictions))
		branching_mu.Lock()
		branching[len(context.chosen)] += float32(len(context.restrictions[new_v_s]))
		branching_counter[len(context.chosen)]++
		branching_mu.Unlock()
		//functionality
		for u_instance := range context.restrictions[new_v_s] {
			ret += RecursionSearch(context, u_instance, new_v_s)
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

// func MinRestrictionsCall(Graph graph, Subgraph graph, restrictions map[uint64]map[uint64]void,
// 	path map[uint64]uint64, chosen map[uint64]void, ordering []uint64, file *os.File) int {
// 	ret := 0
// 	best_length := ^uint64(0)
// 	targets := []uint64{}
// 	new_v_s := uint64(0)
// 	for t := range restrictions {
// 		if uint64(len(restrictions[t])) < best_length {
// 			new_v_s = t
// 			best_length = uint64(len(restrictions[t]))
// 		}
// 		if best_length <= 1 {
// 			fmt.Println("targets size", best_length, "depth", len(path), "vertex", new_v_s, "short skip")
// 			new_v_g := uint64(0)
// 			for v := range restrictions[new_v_s] {
// 				new_v_g = v
// 			}
// 			ret += RecursionSearch(Graph, Subgraph, new_v_g, new_v_s, restrictions, path, chosen)
// 			return ret
// 		}
// 	}
// 	for u_instance := range restrictions[new_v_s] {
// 		targets = append(targets, u_instance)
// 	}
// 	fmt.Println("targets size", len(targets), "death", len(path), "vertex", new_v_s)
// 	for i := 0; i < len(targets); i++ {
// 		ret += RecursionSearch(Graph, Subgraph, targets[i], new_v_s, restrictions, path, chosen)
// 	}
// 	return ret
// }

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
		*single_rest = ColoredNeighborhood(context.Graph, v_g, context.Subgraph[u].attribute.color)
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
		if Graph[v].attribute.color == c {
			output[v] = void{}
		}
	}
	return output
}

func PriorityColoredNeighborhood(Graph graph, u uint64, c uint32, deg int) map[uint64]uint64 {
	output := make(map[uint64]uint64, len(Graph[u].neighborhood))
	for v := range Graph[u].neighborhood {
		if len(Graph[v].neighborhood) >= deg {
			if Graph[v].attribute.color == c || Graph[v].attribute.color == ^uint32(0) {
				output[v] = 0
			}
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

func minimumSpanningTree(G graph, start uint64) graph {
	visited := make(map[uint64]bool)
	mst := make(graph)
	var dfs func(uint64)

	for v := range G {
		mst.AddVertex(v, G[v].attribute.color)
	}
	dfs = func(v uint64) {
		visited[v] = true

		for u := range G[v].neighborhood {
			if !visited[u] {
				// Add this edge to the MST
				mst.AddEdge(v, u)
				dfs(u)
			}
		}
	}

	dfs(start)
	return mst
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
