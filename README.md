# Subgraph Isomorphism Solver

This project provides a solver for subgraph isomorphism problems (both induced and non-induced). It includes a Go implementation for efficient graph processing. **This is the minimal branch, optimized for speed, and not the debug version.** It includes support for both directed and induced subgraph isomorphism.

## Compiling

You need Go version 1.22.3 or higher.

To compile the project, run:
```
$ go build -o subgraph_isomorphism
```

This will generate an executable named `subgraph_isomorphism`.

## Testing

To run the tests, use:
```
$ go test
```

The tests include random graph generation (Gnp), so there is a very low probability of test failures due to randomness. If a test fails, rerunning it can verify the result.

## Running

To execute the solver:
```
$ ./subgraph_isomorphism input_graph_file subgraph_file
```

This will find subgraph isomorphisms of the `subgraph_file` inside the `input_graph_file`. The program expects the input files to be provided in the correct format (see below) and outputs results to the `dat/` directory.

### Example:
```
$ ./subgraph_isomorphism inputs/graph0.json inputs/graph1.json
```

### Help:
To see all available flags and options, use:
```
$ ./subgraph_isomorphism -h
```

### Additional Flags:
- `-prior`: Specify the prior policy for vertex selection. Options include:
  - `0`: Degree squared in the subgraph.
  - `1`: Degree squared in the input graph.
  - `2`: Constant.
  - `3`: Random.
  - `4`: Degree in the subgraph.
  - `5`: Combined strategies.
- `-directed`: Enable directed graph processing.
- `-induced`: Enable induced subgraph isomorphism.

## Expected Input and Output Formats

### Input Graphs:
- **JSON Format** (`-fmt=json`):
  - **Structure**: Node-link JSON (compatible with NetworkX).
  - **Node IDs**: `uint32`.
  - **Color IDs**: `uint16`.
  - **Example**:
    ```json
    {
      "nodes": [
        {"id": 0, "color": 1},
        {"id": 1, "color": 2}
      ],
      "links": [
        {"source": 0, "target": 1}
      ]
    }
    ```

- **Folder Format** (`-fmt=folder`):
  - **Structure**: A directory containing two files:
    - `<name>.node_labels`: Each line specifies a node ID and its color, e.g., `0 1`.
    - `<name>.edges`: Each line specifies an edge between two nodes, e.g., `0 1`.
  - **Example**:
    ```
    folder/
    ├── graph.node_labels
    └── graph.edges
    ```

### Outputs:
Results are saved in the `dat/` directory and include:
- Subgraph isomorphism matches.
- Depth and branching factor logs (if enabled).
- Profiling data (if enabled).

## Notes

- The project includes a `.gitignore` file to exclude temporary files, logs, and outputs.

This project is designed for research and experimentation with subgraph isomorphism problems. Contributions and feedback are welcome!
