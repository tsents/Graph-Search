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

type fisset [][]uint32

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

	inverse_restrictions, empty := update_restrictions(Graph, Subgraph, v_g, v_s, restrictions)
	defer reverse_restrictions(restrictions, inverse_restrictions)

	if !empty {
		for i := 0; i < len(restrictions[v_s+1]); i++ {
			recursion_search(Graph, Subgraph, restrictions[v_s+1][i], v_s+1, restrictions, path, file) // for some resone modifies the retrictions?! but doesnt reverse currectly
		}
	}
	delete(path, v_g)
}

func reverse_restrictions(restrictions fisset, inverse_restrictions fisset) {
	for i := 0; i < len(inverse_restrictions); i++{
		if len(inverse_restrictions[i]) > 0 && inverse_restrictions[i][0] == ^uint32(0){
			restrictions[i] = nil
		} else {
			restrictions[i] = append(restrictions[i], inverse_restrictions[i]...)
		}
	} 
}

func update_restrictions(G graph, S graph, v_g uint32, v_s uint32, restrictions fisset) (fisset, bool) {
	empty := false
	inverse_restrictions := make(fisset, len(S))
	for u := range S[v_s].neighborhood {
		if restrictions[u] == nil {
			restrictions[u] = neighboorhood_with_color(G, v_g, S[u].attribute.color)
			inverse_restrictions[u] = []uint32{^uint32(0)}
		} else {
			restrictions[u], inverse_restrictions[u] = uninitialized_update(G, restrictions, u, v_g)
			if len(restrictions[u]) == 0 {
				empty = true
			}
		}
	}
	return inverse_restrictions, empty
}

func uninitialized_update(G graph, restrictions fisset, u uint32, v_g uint32) ([]uint32,[]uint32) {
	temp_retriction := []uint32{}
	temp_inverse := []uint32{}
	for i := 0; i < len(restrictions[u]); i++{
		if _, ok := G[v_g].neighborhood[restrictions[u][i]]; ok {
			temp_retriction = append(temp_retriction, restrictions[u][i])
		} else {
			temp_inverse = append(temp_inverse, restrictions[u][i])
		}
	}
	return temp_retriction, temp_inverse
}

func neighboorhood_with_color(Graph graph, u uint32, c uint8) []uint32 {
	output := []uint32{}
	for v := range Graph[u].neighborhood {
		if Graph[v].attribute.color == c {
			output = append(output, v)
		}
	}
	return output
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

func ReverseMap(m map[uint32]uint32) map[uint32]uint32 {
	n := make(map[uint32]uint32, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}
