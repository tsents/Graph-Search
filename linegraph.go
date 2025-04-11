package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

type void struct{}

type graph map[uint64]vertex

type vertex struct {
	attribute        att
	neighborhood     map[uint64]void
	neighborhood_in  map[uint64]void
	neighborhood_out map[uint64]void
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
}

var prior_policy *int
var output_file *os.File
var directed *bool
var induced *bool
var Finf float32 = math.Float32frombits(0x7F800000)

func main() {
	runtime.GOMAXPROCS(32)                        //regularization, keeps cpu under control
	debug.SetMaxStack(10 * 128 * 1024 * 1024)     //GBit
	debug.SetMemoryLimit(200 * 128 * 1024 * 1024) //GBit

	out_fname := flag.String("out", "dat/output.txt", "output location")
	input_fmt := flag.String("fmt", "json", "The file format to read\njson node-link,folder to .edges,.labels")
	input_parse := flag.String("parse", "%d\t%d", "The parse format of reading from file, used only for folder fmt")
	prior_policy = flag.Int("prior", 0, "the prior of the information we gain from vertex, based on our method S=0,G=1 or Constant=2, Random=3,Gready based on S=4,Combined Method S=5")
	directed = flag.Bool("directed", false, "change to directed graphs, default is automaticly undirected")
	induced = flag.Bool("induced", false, "change to induced subgraph, default is automaticly non-induced")
	flag.Parse()

	fmt.Println("output ->", *out_fname)
	fmt.Println("parsing :", *input_parse)
	fmt.Println("prior :", *prior_policy)
	var err error
	output_file, err = os.Create(*out_fname)
	if err != nil {
		panic(err)
	}
	defer output_file.Close()
	G := ReadGraph(flag.Args()[0], *input_fmt, *input_parse)
	S := ReadGraph(flag.Args()[1], *input_fmt, *input_parse)

	var wg sync.WaitGroup
	if *prior_policy == 5 {
		wg.Add(2)
		prior := calculatePrior(S, G, 0)
		go func() {
			FindAll(G, S, prior)
			wg.Done()
		}()
		prior = calculatePrior(S, G, 4)
		go func() {
			FindAll(G, S, prior)
			wg.Done()
		}()
		wg.Wait()
		fmt.Println("done")
		return
	}

	prior := calculatePrior(S, G, *prior_policy)
	matches := FindAll(G, S, prior)
	fmt.Println("matches", matches)
}

func calculatePrior(S, G graph, prior_policy int) map[uint64]float32 {
	prior := make(map[uint64]float32)
	switch prior_policy {
	case 0: // d^2 in S
		for v := range S {
			for u := range S[v].neighborhood {
				prior[v] += float32(len(S[u].neighborhood))
			}
		}
	case 1: // d^2 in G
		for v := range G {
			for u := range G[v].neighborhood {
				prior[v] += float32(len(G[u].neighborhood))
			}
		}
	case 4: // d in S
		for v := range S {
			prior[v] = float32(len(S[v].neighborhood))
		}
	}
	return prior
}

func ChooseNext[T any](restrictions map[uint64]map[uint64]T, chosen map[uint64]void, Subgraph graph, prior map[uint64]float32) uint64 {
	//we want the max number of errors, but also min length
	max_score := -Finf //negitive inf for floats
	idx := ^uint64(0)

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
	max_score := -Finf //negitive inf for floats
	var idx uint64
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
		return -score
	case 2:
		return float32(-len(rest[u]))
	case 3:
		return rand.Float32()
	case 4:
		return prior[u]
	case 5:
		return prior[u]
	}
	return 0
}

func FindAll(Graph graph, Subgraph graph, prior map[uint64]float32) uint64 {
	var wg sync.WaitGroup
	var ops atomic.Uint64
	// t := time.Now()
	// f, err := os.Create("dat/" + t.Format("2006-01-02 15:04:05.999999") + ".txt")
	//functionality
	v_0 := ChooseStart(Subgraph, prior)
	for u := range Graph {
		if Graph[u].attribute.color == Subgraph[v_0].attribute.color {
			wg.Add(1)
			context := context{Graph: Graph,
				Subgraph:     Subgraph,
				restrictions: make(map[uint64]map[uint64]void, len(Subgraph)),
				path:         make(map[uint64]uint64),
				chosen:       make(map[uint64]void),
				prior:        prior}
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
	return ops.Load()
}

func RecursionSearch(context *context, v_g uint64, v_s uint64) int {
	//functionality
	if _, ok := context.path[v_g]; ok {
		return 0
	}
	if len(context.Subgraph) == (len(context.chosen) + 1) {
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

			SingleUpdate(context, u, v_s, v_g, &single_inverse, &single_rest)

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
func SingleUpdate(context *context, u uint64, v_s uint64, v_g uint64, single_inverse *map[uint64]void, single_rest *map[uint64]void) {
	if *single_rest == nil {
		*single_inverse = make(map[uint64]void)
		(*single_inverse)[^uint64(0)] = void{}
		if !*directed {
			*single_rest = ColoredNeighborhood(context.Graph, v_g, context.Subgraph[u].attribute.color, len(context.Subgraph[u].neighborhood))
			return
		}

		//directed
		if _, ok := context.Subgraph[v_s].neighborhood_out[u]; ok { //if v_s has an edge to u
			*single_rest = ColoredNeighborhoodOut(context.Graph, v_g, context.Subgraph[u].attribute.color, len(context.Subgraph[u].neighborhood_out))
		}
		if _, ok := context.Subgraph[v_s].neighborhood_in[u]; ok { //if v_s has an edge from u
			if _, ok := context.Subgraph[v_s].neighborhood_out[u]; !ok { //and v_s doesnt have edge to u then finish with one sided
				*single_rest = ColoredNeighborhoodIn(context.Graph, v_g, context.Subgraph[u].attribute.color, len(context.Subgraph[u].neighborhood_in))
				return
			}
			// if both directions exist, compute the intersection
			intersection := make(map[uint64]void)
			for key := range *single_rest {
				if _, ok := ColoredNeighborhoodIn(context.Graph, v_g, context.Subgraph[u].attribute.color, len(context.Subgraph[u].neighborhood_in))[key]; !ok {
					intersection[key] = void{}
				}
			}
			*single_rest = intersection
		}
		return
	}
	// _, ok := G[v_g].neighborhood[u_instance]
	if !*directed {
		for u_instance := range *single_rest {
			if _, ok := context.Graph[v_g].neighborhood[u_instance]; !ok {
				if *single_inverse == nil {
					*single_inverse = make(map[uint64]void)
				}
				(*single_inverse)[u_instance] = void{}
				delete(*single_rest, u_instance)
			}
		}
		return
	}
	if _, ok := context.Subgraph[v_s].neighborhood_out[u]; ok {
		for u_instance := range *single_rest {
			if _, ok := context.Graph[v_g].neighborhood_out[u_instance]; !ok {
				if *single_inverse == nil {
					*single_inverse = make(map[uint64]void)
				}
				(*single_inverse)[u_instance] = void{}
				delete(*single_rest, u_instance)
			}
		}
	}
	if _, ok := context.Subgraph[v_s].neighborhood_in[u]; ok {
		for u_instance := range *single_rest {
			if _, ok := context.Graph[v_g].neighborhood_in[u_instance]; !ok {
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
			if (!(*induced) && deg <= len(Graph[v].neighborhood)) || deg == len(Graph[v].neighborhood) { //if not induced allow the easier way
				output[v] = void{}
			}
		}
	}
	return output
}

func ColoredNeighborhoodOut(Graph graph, u uint64, c uint32, deg int) map[uint64]void {
	output := make(map[uint64]void)
	for v := range Graph[u].neighborhood_out {
		if Graph[v].attribute.color == c {
			if (!(*induced) && deg <= len(Graph[v].neighborhood_out)) || deg == len(Graph[v].neighborhood_out) { //if not induced allow the easier way
				output[v] = void{}
			}
		}
	}
	return output
}

func ColoredNeighborhoodIn(Graph graph, u uint64, c uint32, deg int) map[uint64]void {
	output := make(map[uint64]void)
	for v := range Graph[u].neighborhood_in {
		if Graph[v].attribute.color == c {
			if (!(*induced) && deg <= len(Graph[v].neighborhood_in)) || deg == len(Graph[v].neighborhood_in) { //if not induced allow the easier way
				output[v] = void{}
			}
		}
	}
	return output
}

func (Graph graph) AddVertex(u uint64, c uint32) {
	if _, ok := Graph[u]; ok {
		return
	}
	Graph[u] = vertex{neighborhood: make(map[uint64]void), attribute: att{color: c}, neighborhood_in: make(map[uint64]void), neighborhood_out: make(map[uint64]void)}
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
	if *directed {
		Graph[u].neighborhood_out[v] = void{}
		Graph[v].neighborhood_in[u] = void{}
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
