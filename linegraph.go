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

var prior_policy *int
var recolor_policy *int
var start_point *int64
var output_file *os.File

// var next_function func(Subgraph graph,rest map[uint64]map[uint64]uint64,u uint64,prior uint64)uint64;

func main() {
	runtime.GOMAXPROCS(32)                        //regularization, keeps cpu under control
	debug.SetMaxStack(2 * 128 * 1024 * 1024)      //GBit
	debug.SetMemoryLimit(200 * 128 * 1024 * 1024) //GBit

	out_fname := flag.String("out", "dat/output.txt", "output location")
	cmd_error := flag.Int("err", 0, "number of errors in the search\ndefault is exact isomorphism (default 0)")
	input_fmt := flag.String("fmt", "json", "The file format to read\njson node-link,folder to .edges,.labels")
	input_parse := flag.String("parse", "%d\t%d", "The parse format of reading from file, used only for folder fmt")
	prior_policy = flag.Int("prior", 0, "the prior of the information we gain from vertex, based on S=0,G=1 or Constant=2")
	subset_size := flag.Int64("subset", -1, "take as subset of this size from G, to be the Subgraph")
	recolor_policy = flag.Int("recolor", -1, "recolor value policy, defualt is base on read,else is rand.N")
	profile := flag.Bool("prof", false, "profile the program")
	start_point = flag.Int64("start", 1, "the starting point of the search")

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
	colorDist(G)
	colorDist(S)
	// ordering := ReadOrdering(fmt.Sprintf("inputs/ordering_%v_%v.json", i, j))
	start := time.Now()
	// FindAllSubgraphPathgraph(G, S, ordering, fmt.Sprintf("output%v_%v", i, j))

	// if num_errors != 0{
	// 	IncompleteFindAll(G, S, uint64(num_errors))
	// } else {
	// 	FindAllSubgraphPathgraph(G,S,[]uint64{uint64(*start_point)},"bla")
	// }
	var matches uint64 = IncompleteFindAll(G, S, uint64(num_errors))
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

func IncompleteFindAll(Graph graph, Subgraph graph, threshold uint64) uint64 {
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
				ret := RecursionSearch(Graph, Subgraph, u, ordering[0], make(map[uint64]map[uint64]void, len(Subgraph)),
					make(map[uint64]uint64), make(map[uint64]void), f, ordering)

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
	restrictions map[uint64]map[uint64]void, path map[uint64]uint64, chosen map[uint64]void, file *os.File, ordering []uint64) int {
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
	defer delete(chosen, v_s)

	self_list := restrictions[v_s]
	delete(restrictions, v_s)
	inverse_restrictions, empty := UpdateRestrictions(Graph, Subgraph, v_g, v_s, restrictions, chosen)
	inverse_restrictions[v_s] = self_list
	if !empty {
		if len(path) < len(ordering) {
			targets := []uint64{}
			new_v_s := uint64(0)
			new_v_s = ordering[len(path)]
			for u_instance := range restrictions[new_v_s] {
				targets = append(targets, u_instance)
			}
			fmt.Println("targets size", len(targets), "depth", len(path))
			for i := 0; i < len(targets); i++ {
				ret += RecursionSearch(Graph, Subgraph, targets[i], new_v_s, restrictions, path, chosen, file, ordering)
			}
		} else {
			ret += MinRestrictionsCall(Graph, Subgraph, restrictions, path, chosen, ordering, file)
		}
	}
	for u := range inverse_restrictions {
		if _, ok := inverse_restrictions[u][^uint64(0)]; ok {
			delete(restrictions, u)
		} else {
			for u_instance := range inverse_restrictions[u] {
				if restrictions[u] == nil {
					restrictions[u] = make(map[uint64]void)
				}
				restrictions[u][u_instance] = void{}
			}
		}
	}
	return ret
}

func MinRestrictionsCall(Graph graph, Subgraph graph, restrictions map[uint64]map[uint64]void,
	path map[uint64]uint64, chosen map[uint64]void, ordering []uint64, file *os.File) int {
	ret := 0
	best_length := ^uint64(0)
	targets := []uint64{}
	new_v_s := uint64(0)
	for t := range restrictions {
		if uint64(len(restrictions[t])) < best_length {
			new_v_s = t
			best_length = uint64(len(restrictions[t]))
		}
		if best_length <= 1 {
			fmt.Println("targets size", best_length, "depth", len(path), "vertex", new_v_s, "short skip")
			new_v_g := uint64(0)
			for v := range restrictions[new_v_s] {
				new_v_g = v
			}
			ret += RecursionSearch(Graph, Subgraph, new_v_g, new_v_s, restrictions, path, chosen, file, ordering)
			return ret
		}
	}
	for u_instance := range restrictions[new_v_s] {
		targets = append(targets, u_instance)
	}
	fmt.Println("targets size", len(targets), "death", len(path), "vertex", new_v_s)
	for i := 0; i < len(targets); i++ {
		ret += RecursionSearch(Graph, Subgraph, targets[i], new_v_s, restrictions, path, chosen, file, ordering)
	}
	return ret
}

func UpdateRestrictions(G graph, S graph, v_g uint64, v_s uint64,
	restrictions map[uint64]map[uint64]void, chosen map[uint64]void) (map[uint64]map[uint64]void, bool) {
	empty := false
	inverse_restrictions := make(map[uint64]map[uint64]void)
	for u := range S[v_s].neighborhood {
		if _, ok := chosen[u]; !ok {
			if _, ok := restrictions[u]; !ok {
				restrictions[u] = ColoredNeighborhood(G, v_g, S[u].attribute.color)
				inverse_restrictions[u] = make(map[uint64]void)
				inverse_restrictions[u][^uint64(0)] = void{}
			} else {
				// _, ok := G[v_g].neighborhood[u_instance]
				for u_instance := range restrictions[u] {
					if _, ok := G[v_g].neighborhood[u_instance]; !ok {
						if inverse_restrictions[u] == nil {
							inverse_restrictions[u] = map[uint64]void{}
						}
						inverse_restrictions[u][u_instance] = void{}
						delete(restrictions[u], u_instance)
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

func ColoredNeighborhood(Graph graph, u uint64, c uint32) map[uint64]void {
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
