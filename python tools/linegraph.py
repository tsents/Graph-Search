import networkx as nx
import random
import matplotlib.pyplot as plt
import numpy as np
from networkx.algorithms import isomorphism
import time
from to_line_graph import *

from multiprocessing.pool import ThreadPool as Pool


#Debugging:
def print_graph(graph,colormap,G):
    # Draw the graph with colored nodes
    pos = nx.spring_layout(G, seed=42)  # Position nodes using spring layout
    # pos = {i:[i,pow(-1,i)] for i in range(100)} displays subgraphs nicely
    nx.draw(graph, pos, node_color=list(colormap.values()), cmap=plt.cm.get_cmap("viridis"), with_labels=True)

    nodes_data_list = list(graph.nodes(data=True))
    print(nodes_data_list)

    # Show the plot
    plt.show()

def format_as_array(graph_dict):
    output = [i for i in range((len(graph_dict)))]
    for key,val in graph_dict.items():
        output[val] = key
    return output

#with no multiple finds

def search_all_no_repetition(Graph,Subgraph,ordering,common_colors):
    all_subgraphs = []

    def wrapper(node):
        if Graph.nodes[node]["color"] == Subgraph.nodes[ordering[0]]["color"]:
            output = recursion_search_ordered(Graph,Subgraph,node,ordering[0],{},{},ordering) #we can run this part in parallel, to massivly boost speed
            all_subgraphs.extend(output)
    # pool = Pool(64)

    # for node in Graph.nodes():
    #     pool.apply_async(wrapper, (node,))    

    # pool.close()
    # pool.join()
    for node in Graph.nodes():
        wrapper(node)
    return all_subgraphs

#modifies G
def recursion_search_no_repetition(G,S,node_g,node_s,restrictions,path,ordering,common_colors):
    if node_g in path:
        return []
    if len(path) >= len(ordering) - 1: #len(S.nodes()) was bugged? not now?
        #we want to lower memory so for now we wont store the answers
        copy = path.copy()
        copy[node_g] = node_s
        G.remove_nodes_from(copy.keys())
        return [copy]
    output = []
    # print("before update",restrictions)
    inverse_restrictions, empty_set = restriction_update(G,S,node_g,node_s,restrictions)
    # print("after update",restrictions,inverse_restrictions)
    path[node_g] = node_s

    if not empty_set:
        for u in restrictions[ordering[len(path)]]: #the next node in the ordering! in the extention of the algoritms this is a part we will change
            recursion_output = recursion_search_ordered(G,S,u,ordering[len(path)],restrictions,path,ordering)
            output.extend(recursion_output)
            if len(recursion_output) > 0:
                return output

    #do the union of the changes, i.e reverse the new imposed retrictions
    # print("before inverse",restrictions,inverse_restrictions)
    for u in inverse_restrictions:
        if len(inverse_restrictions[u]) > 0 and inverse_restrictions[u][0] == "G":
            restrictions.pop(u)
        else:
            restrictions[u].extend(inverse_restrictions[u]) 
            #Problem!! this is not constant, but O(n), should move to custom linked lists!
            #In profiling this part wasnt even visible, very minor
    # print("after inverse",restrictions,inverse_restrictions)
    path.pop(node_g)
    return output

#for any ordering
def search_all_subgraphs_orderd(Graph,Subgraph,ordering):
    all_subgraphs = []

    def wrapper(node):
        if Graph.nodes[node]["color"] == Subgraph.nodes[ordering[0]]["color"]:
            print("in",node)
            output = recursion_search_ordered(Graph,Subgraph,node,ordering[0],{},{},ordering) #we can run this part in parallel, to massivly boost speed
            all_subgraphs.extend(output)
            print("done",node)

    # pool = Pool(64)

    # for node in Graph.nodes():
    #     pool.apply_async(wrapper, (node,))    

    # pool.close()
    # pool.join()
    for node in Graph.nodes():
        wrapper(node)
    #     if Graph.nodes[node]["color"] == Subgraph.nodes[ordering[0]]["color"]:
    #         output = recursion_search_ordered(Graph,Subgraph,node,ordering[0],{},{},ordering) #we can run this part in parallel, to massivly boost speed
    #         all_subgraphs.extend(output)
    #         # print("started from", node, output,len(output)) #the bug is somewhere here!
    return all_subgraphs

def recursion_search_ordered(G,S,node_g,node_s,restrictions,path,ordering):
    if node_g in path:
        return []
    if len(path) >= len(ordering) - 1: #len(S.nodes()) was bugged? not now?
        #we want to lower memory so for now we wont store the answers
        copy = path.copy()
        copy[node_g] = node_s
        return [copy]
    output = []
    # print("before update",restrictions)
    inverse_restrictions, empty_set = restriction_update(G,S,node_g,node_s,restrictions)
    # print("after update",restrictions,inverse_restrictions)
    path[node_g] = node_s

    if not empty_set:
        for u in restrictions[ordering[len(path)]]: #the next node in the ordering! in the extention of the algoritms this is a part we will change
            output.extend(recursion_search_ordered(G,S,u,ordering[len(path)],restrictions,path,ordering))

    #do the union of the changes, i.e reverse the new imposed retrictions
    # print("before inverse",restrictions,inverse_restrictions)
    for u in inverse_restrictions:
        if len(inverse_restrictions[u]) > 0 and inverse_restrictions[u][0] == "G":
            restrictions.pop(u)
        else:
            restrictions[u].extend(inverse_restrictions[u]) 
            #Problem!! this is not constant, but O(n), should move to custom linked lists!
            #In profiling this part wasnt even visible, very minor
    # print("after inverse",restrictions,inverse_restrictions)
    path.pop(node_g)
    return output

def search_all_subgraphs(Graph,Subgraph):
    all_subgraphs = []
    for node in Graph.nodes():
        if Graph.nodes[node]["color"] == Subgraph.nodes[0]["color"]:
            output = recursion_search(Graph,Subgraph,node,0,{},{}) #we can run this part in parallel, to massivly boost speed
            all_subgraphs.extend(output)
            # print("started from", node, output,len(output)) #the bug is somewhere here!
    return all_subgraphs

# Graph and Subgraph dont change, node_g is the current node in G,node_s is the corresponding position in S
# restriction - dictionary of lists, path - set of chosen nodes
def recursion_search(G,S,node_g,node_s,restrictions,path): 
    if node_g in path:
        return []
    if len(path) >= len(S.nodes()) - 1: #len(S.nodes()) was bugged? not now?
        return []
        #we want to lower memory so for now we wont store the answers
        copy = path.copy()
        copy[node_g] = node_s
        return [copy]
    output = []
    # print("before update",restrictions)
    inverse_restrictions, empty_set = restriction_update(G,S,node_g,node_s,restrictions)
    # print("after update",restrictions,inverse_restrictions)
    path[node_g] = node_s

    if not empty_set:
        for u in restrictions[node_s + 1]: #the next node in the ordering! in the extention of the algoritms this is a part we will change
            output.extend(recursion_search(G,S,u,node_s+1,restrictions,path))

    #do the union of the changes, i.e reverse the new imposed retrictions
    # print("before inverse",restrictions,inverse_restrictions)
    for u in inverse_restrictions:
        if len(inverse_restrictions[u]) > 0 and inverse_restrictions[u][0] == "G":
            restrictions.pop(u)
        else:
            restrictions[u].extend(inverse_restrictions[u]) 
            #Problem!! this is not constant, but O(n), should move to custom linked lists!
            #In profiling this part wasnt even visible, very minor
    # print("after inverse",restrictions,inverse_restrictions)
    path.pop(node_g)
    return output

def restriction_update(G,S,node_g,node_s,restrictions):
    empty = False
    inverse_restrictions = {}
    for u in S.neighbors(node_s):
        # if u > node_s: #Optional! makes it quicker to retreat in the DFS
        if u in restrictions: #if there are restrictions on u
            marginal_rest, inverse_marginal_rest = [], [] #A_i in the paper
            for u_instance in restrictions[u]:
                (inverse_marginal_rest, marginal_rest)[G.has_edge(node_g,u_instance)].append(u_instance)
            if len(marginal_rest) == 0:
                empty = True
            restrictions[u] = marginal_rest
            inverse_restrictions[u] = inverse_marginal_rest
        else:
            restrictions[u] = neighboorhood_with_color(node_g,G,S.nodes[u]["color"])
            inverse_restrictions[u] = ["G"]
    return inverse_restrictions,empty

def neighboorhood_with_color(node,Graph,color):
    return [n for n in Graph.neighbors(node) if Graph.nodes[n]["color"] == color]

def colors_match(n1_attrib,n2_attrib):
    '''returns False if either does not have a color or if the colors do not match'''
    try:
        return n1_attrib['color']==n2_attrib['color']
    except KeyError:
        return False

def test_matching(size_g,size_s,k,p_g,p_s):
    G = nx.gnp_random_graph(size_g, p_g)

    color_map = {node: random.randint(1, k) for node in G.nodes()}
    nx.set_node_attributes(G, color_map, "color")


    Subgraph = nx.path_graph(size_s)
    added_edges = nx.gnp_random_graph(size_s,p_s)
    Subgraph = nx.compose(Subgraph,added_edges)
    subgraph_color_map = {node: random.randint(1, k) for node in Subgraph.nodes()}
    nx.set_node_attributes(Subgraph, subgraph_color_map, "color")

    G = Subgraph

    start1 = time.time()
    output1 = search_all_subgraphs(G,Subgraph)
    time1 = time.time() - start1

    for g in output1:
        for key,val in g.items():
            if not G.nodes[key]["color"] == Subgraph.nodes[val]["color"]:
                print("fucking up color",G.nodes[key]["color"],Subgraph.nodes[val]["color"])
        temp = format_as_array(g)
        for u,v in Subgraph.edges():
            if not G.has_edge(temp[u],temp[v]):
                print("fucking up edge")
        
    start2 = time.time()
    GM = isomorphism.GraphMatcher(G, Subgraph,node_match=colors_match)
    output2 = GM.subgraph_monomorphisms_iter()
    output2 = list(output2)
    time2 = time.time() - start2

    if len(output1) != len(output2):
        print("wrong answer :(")
    print(len(output1),len(output2),time1,time2)
    return (time1,time2)

def run_without_test(size_g,size_s,k,p_g,p_s):
    G = nx.gnp_random_graph(size_g, p_g)

    color_map = {node: random.randint(1, k) for node in G.nodes()}
    nx.set_node_attributes(G, color_map, "color")

    Subgraph = nx.path_graph(size_s)
    added_edges = nx.gnp_random_graph(size_s,p_s)
    Subgraph = nx.compose(Subgraph,added_edges)
    subgraph_color_map = {node: random.randint(1, k) for node in Subgraph.nodes()}
    nx.set_node_attributes(Subgraph, subgraph_color_map, "color")

    start1 = time.time()
    output1 = search_all_subgraphs(G,Subgraph)
    time1 = time.time() - start1


    ordering = [i for i in range(Subgraph.number_of_nodes())]
    start2 = time.time()
    output2 = search_all_subgraphs_orderd(G,Subgraph,ordering)
    time2 = time.time() - start2

    print(time1,time2,len(output1),len(output2))

    return output1

def read_json_file(filename):
    with open(filename) as f:
        js_graph = json.load(f)
    g = nx.node_link_graph(js_graph)
    g = nx.convert_node_labels_to_integers(g)
    return g

if __name__ == "__main__":
    graph = read_json_file('graph0.json')
    sub_g = read_json_file('graph2.json')
    start2 = time.time()
    GM = isomorphism.GraphMatcher(graph, sub_g,node_match=colors_match)
    output2 = GM.subgraph_monomorphisms_iter()
    output2 = list(output2)
    time2 = time.time() - start2
    print(output2,len(output2),time2)
    # graph, node_to_label_g, label_to_node_g = nx.read_gpickle("erdos_renyi_graph_p=0.5" + "_{}_classes".format(5))
    # sub_g, node_to_label_s , label_to_node_s = nx.read_gpickle("sub_erdos_renyi_graph_p=0.5" + "_{}_classes".format(5))

    # #plot_graph(sub_g, node_to_label_s, "Subgraph", "subgraph.png")
    # G_edge_ranks = rank_edges(graph, node_to_label_g)
    # hamiltonian = cheapest_hamiltonian(sub_g, node_to_label_s, G_edge_ranks)
    # print(hamiltonian)
    # print(len(hamiltonian),len(sub_g),len(graph))

    # nx.set_node_attributes(graph, node_to_label_g, "color")
    # nx.set_node_attributes(sub_g, node_to_label_s, "color")

    # start2 = time.time()
    # GM = isomorphism.GraphMatcher(graph, sub_g,node_match=colors_match)
    # output2 = GM.subgraph_monomorphisms_iter()
    # for p in output2:
    #     print(p)
    # time2 = time.time() - start2
    # print(time2,len(output2))

    # start1 = time.time()
    # search_all_subgraphs_orderd(graph,sub_g,hamiltonian)
    # time1 = time.time() - start1
    # print("TIME",time1)
    

# sum1 = 0
# sum2 = 0
# for i in range(10):
#     run_without_test(int(1e2),int(1e1),5,0.5,0.5)
#     sum1 += time1
#     sum2 += time2
# print(sum1,sum2)



