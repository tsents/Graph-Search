import networkx as nx
import json
import time
import linegraph

def rank_edges(graph):
    # Count all optional edges in G to find rarity.
    edge_list = list(graph.edges)
    edge_ranks = {}
    for edge in edge_list:
        labeled_e = (graph.nodes[edge[0]]['color'], graph.nodes[edge[1]]['color'])
        if labeled_e in edge_ranks:
            edge_ranks[labeled_e] += 1
        elif (labeled_e[1],labeled_e[0]) in edge_ranks:
            edge_ranks[(labeled_e[1],labeled_e[0])] += 1
        else:
            edge_ranks[labeled_e] = 1
    return edge_ranks

def cheapest_hamiltonian(sub_g,edge_ranks):
    # Add ranks by rarity in G to edges in sub graph.
    edge_list = list(sub_g.edges)
    for edge in edge_list:
        labeled_e = (sub_g.nodes[edge[0]]['color'], sub_g.nodes[edge[1]]['color'])
        # If the connection between these 2 labels is in dict:
        if labeled_e in edge_ranks:
            sub_g[edge[0]][edge[1]]['weight'] = edge_ranks[labeled_e]
        elif (labeled_e[1], labeled_e[0]) in edge_ranks:
            rank = edge_ranks[(labeled_e[1], labeled_e[0])]
            sub_g[edge[0]][edge[1]]['weight'] = rank
        # If the connection between these 2 labels is not in G, exact sub isomorphism can not be found.
        else:
            return -1, edge, labeled_e
    # Find the cheapest Hamiltonian.
    tsp = nx.approximation.traveling_salesman_problem
    path = tsp(sub_g, cycle=False)
    seen = set()
    unique_path = [x for x in path if not (x in seen or seen.add(x))]
    return unique_path



def prim_hamiltonian(sub_g,edge_ranks):
    min_rank = max(edge_ranks.values())
    start = -1
    # Add ranks by rarity in G to edges in sub graph.
    edge_list = list(sub_g.edges)
    for edge in edge_list:
        labeled_e = (sub_g.nodes[edge[0]]['color'], sub_g.nodes[edge[1]]['color'])
        if labeled_e in edge_ranks:
            rank = edge_ranks[labeled_e]
        elif (labeled_e[1], labeled_e[0]) in edge_ranks:
            rank = edge_ranks[(labeled_e[1], labeled_e[0])]
        else:
            return -1, edge, labeled_e
        
        if rank < min_rank:
            start = edge[0]
            min_rank = rank
        sub_g[edge[0]][edge[1]]['weight'] = rank


    chosen = [start]
    valued_neighborhood = {}
    for i in range(len(sub_g) - 1):
        for neighbor in sub_g.neighbors(chosen[i]):
            if not neighbor in chosen:
                if not neighbor in valued_neighborhood:
                    valued_neighborhood[neighbor] = 0
                labeled_e = (sub_g.nodes[chosen[i]]['color'], sub_g.nodes[neighbor]['color'])
                if labeled_e in edge_ranks:
                    rank = edge_ranks[labeled_e]
                elif (labeled_e[1], labeled_e[0]) in edge_ranks:
                    rank = edge_ranks[(labeled_e[1], labeled_e[0])]
                valued_neighborhood[neighbor] += 1/rank
        chosen.append(max(valued_neighborhood, key=valued_neighborhood.get))
        valued_neighborhood.pop(chosen[i+1])
    print(chosen)
    return chosen

def read_json_file(filename):
    with open(filename) as f:
        js_graph = json.load(f)
    g = nx.node_link_graph(js_graph)
    g = nx.convert_node_labels_to_integers(g)
    return g

def full_pipeline(graph_name,subgraph_name,ordering_file = ""):
    G = read_json_file(graph_name)
    S = read_json_file(subgraph_name)
    start1 = time.time()
    hamiltonian = 0
    if len(ordering_file) == 0:
        G_edge_ranks = rank_edges(G)
        hamiltonian = cheapest_hamiltonian(S, G_edge_ranks)
        prim_hamiltonian(S,G_edge_ranks)
    else:
        with open(ordering_file) as f:
            d = json.load(f)
        hamiltonian = d['ordering']
    end1 = time.time() - start1
    print(end1)

    if hamiltonian[0] == -1:
        return False
    
    with open('ordering_'+graph_name[5]+'_'+subgraph_name[5] + '.json', 'w') as f:
        json.dump({"ordering":hamiltonian}, f)

    # start2 = time.time()
    # output = linegraph.search_all_no_repetition(G,S,hamiltonian,[])
    # with open('output_'+graph_name[5]+'_'+subgraph_name[5] + '.json', 'w') as f:
    #     json.dump(output, f)
    # time2 = time.time() - start2
    # print("TIME",time2)

full_pipeline('graph0.json','graph5.json')

