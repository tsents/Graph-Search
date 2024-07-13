package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func loadJSON[T any](filename string) (T, error) {
	var data T
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return data, err
	}
	return data, json.Unmarshal(fileData, &data)
}

func ReadGraph(filename string, input_fmt string,input_parse string) graph {
	switch input_fmt {
	case "json":
		if !strings.HasSuffix(filename, ".json") {
			panic("incorrect suffix")
		}
		return readJsonGrpah(filename)
	case "folder":
		return readFolderGraph(filename,input_parse)
	}
	panic("no valid format")
}
func readJsonGrpah(filename string) graph {
	dat, _ := loadJSON[JSON_graph](filename)
	var output graph = make(graph)
	for i := 0; i < len(dat.Nodes); i++ {
		output.AddVertex(uint64(dat.Nodes[i].ID), uint16(dat.Nodes[i].Color))
	}
	for i := 0; i < len(dat.Links); i++ {
		output.AddEdge(dat.Links[i].Source, dat.Links[i].Target)
	}
	return output
}

func readFolderGraph(dirname string,input_parse string) graph {
	var output graph = make(graph)
	var vertex_fname string
	var edges_fname string
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// fmt.Printf("dir: %v: name: %s\n", info.IsDir(), path)
		if !info.IsDir() {
			if strings.HasSuffix(path, ".node_labels") {
				vertex_fname = strings.Clone(path)
			}
			if strings.HasSuffix(path, ".edges") {
				edges_fname = strings.Clone(path)
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(edges_fname, vertex_fname)
	vertex_file, err := os.Open(vertex_fname)
	if err != nil {
		panic(err)
	}
	defer vertex_file.Close()
	scanner := bufio.NewScanner(vertex_file)
	for scanner.Scan() {
		var vertex uint64
		var color uint16
		_, err := fmt.Sscanf(scanner.Text(), input_parse, &vertex, &color)
		if err != nil{
			panic(err)
		}
		output.AddVertex(vertex,color)
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	edges_file, err := os.Open(edges_fname)
	if err != nil {
		panic(err)
	}
	defer edges_file.Close()
	scanner = bufio.NewScanner(edges_file)
	for scanner.Scan() {
		var u, v uint64
		n, err := fmt.Sscanf(scanner.Text(), input_parse, &u, &v)
		if err != nil || n != 2 {
			fmt.Println("Error parsing line:", scanner.Text())
			continue
		}
		output.AddEdge(u, v)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return output
}

func ReadOrdering(filename string) []uint64 {
	dat, _ := loadJSON[JSON_ordering](filename)
	return dat.Ordering
}
