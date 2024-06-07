package main

import (
	"encoding/json"
	"os"
)

type JSON_ordering struct {
	Ordering []uint32 `json:"ordering"`
}

type JSON_graph struct {
	Directed   bool `json:"directed"`
	Multigraph bool `json:"multigraph"`
	Graph      struct {
	} `json:"graph"`
	Nodes []struct {
		Color uint8 `json:"color"`
		ID    uint32 `json:"id"`
	} `json:"nodes"`
	Links []struct {
		Source uint32 `json:"source"`
		Target uint32 `json:"target"`
	} `json:"links"`
}

func LoadJSON[T any](filename string) (T, error) {
	var data T
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return data, err
	}
	return data, json.Unmarshal(fileData, &data)
}

func ReadGraph(filename string) graph {
	dat, _ := LoadJSON[JSON_graph](filename)
	var output graph = make(graph)
	for i := 0; i < len(dat.Nodes); i++ {
		output.AddVertex(uint32(dat.Nodes[i].ID),uint8(dat.Nodes[i].Color))
	}
	for i := 0; i < len(dat.Links); i++ {
		output.AddEdge(dat.Links[i].Source, dat.Links[i].Target)
	}
	return output
}

func ReadOrdering(filename string) []uint32{
	dat, _ := LoadJSON[JSON_ordering](filename)
	return dat.Ordering
}
