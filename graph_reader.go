package main

import (
	"encoding/json"
	"os"
)

type JSON_ordering struct {
	Ordering []uint64 `json:"ordering"`
}

type JSON_graph struct {
	Nodes []struct {
		Color uint16 `json:"color"`
		ID    uint64 `json:"id"`
	} `json:"nodes"`
	Links []struct {
		Source uint64 `json:"source"`
		Target uint64 `json:"target"`
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
		output.AddVertex(uint64(dat.Nodes[i].ID),uint16(dat.Nodes[i].Color))
	}
	for i := 0; i < len(dat.Links); i++ {
		output.AddEdge(dat.Links[i].Source, dat.Links[i].Target)
	}
	return output
}

func ReadOrdering(filename string) []uint64{
	dat, _ := LoadJSON[JSON_ordering](filename)
	return dat.Ordering
}
