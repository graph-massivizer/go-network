package model

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type NewGraph struct {
	Nodes map[string]NewNode
	Edges map[int]NewEdge
	Type  string
}

type NewNode struct {
	Attributes map[string]interface{}
	ID         string
}

type NewEdge struct {
	First_node  NewNode
	Second_node NewNode
	Attributes  map[string]interface{}
}

func BasicGraph() NewGraph {
	return NewGraph{
		Nodes: map[string]NewNode{},
		Edges: map[int]NewEdge{},
		Type:  "graph",
	}
}

func DiGraph() NewGraph {
	return NewGraph{
		Nodes: map[string]NewNode{},
		Edges: map[int]NewEdge{},
		Type:  "digraph",
	}
}

func MultiGraph() NewGraph {
	return NewGraph{
		Nodes: map[string]NewNode{},
		Edges: map[int]NewEdge{},
		Type:  "multigraph",
	}
}

// Transorfms NewGraph to byte data
func (g *NewGraph) ToJSON() ([]byte, error) {
	return json.Marshal(g)
}

// WriteToFile writes the graph data to a JSON file
func (g *NewGraph) WriteToFile(filename string) error {
	data, err := g.ToJSON()
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// Compares two graphs
func (g NewGraph) IsEqual(other NewGraph) bool {

	if g.Type != other.Type || len(g.Nodes) != len(other.Nodes) || len(g.Edges) != len(other.Edges) {
		return false
	}

	for key := range g.Nodes {
		_, ok := other.Nodes[key]
		if !ok {
			return false
		}
	}

	for i := 0; i < len(g.Edges); i++ {
		if !g.CompareEdges(g.Edges[i], other.Edges[i]) {
			return false
		}
	}

	return true
}

// Removes from graph all nodes that have only one neighbor
func (g *NewGraph) RemoveLeaves() {
	leaves := []NewNode{}

	for _, node := range g.Nodes {
		if len(g.Neighbors(node)) == 1 {
			leaves = append(leaves, node)
		}
	}

	k := g
	for _, leave := range leaves {
		k.RemoveNode(leave)
	}
}

// Combines nodes in graph, that have only one neighbor, with their neighbor
func (g *NewGraph) CombineLeaves() {
	leaves := []NewNode{}

	for _, node := range g.Nodes {
		if len(g.Neighbors(node)) == 1 {
			leaves = append(leaves, node)
		}
	}

	N := make(map[string]NewNode)
	for _, leave := range leaves {
		x := g.Neighbors(leave)[0]
		n := g.CombineNodes(leave, x, StrategyArray{}, StrategyArray{})
		g.Nodes[x.ID] = n
		for _, node := range g.Neighbors(x) {
			e, edge_idx := g.GetEdgeByNodes(x, node)
			if e.First_node.ID == leave.ID || e.Second_node.ID == leave.ID {
				g.RemoveEdge(e)
			} else {
				if e.First_node.ID == x.ID {
					e.First_node = n
				} else {
					e.Second_node = n
				}
			}
			g.Edges[edge_idx] = e

		}
		*g = g.RemoveNode(leave)
		*g = g.RemoveNode(x)
	}
	for _, value := range g.Nodes {
		N[value.ID] = value
	}
	g.Nodes = N
}

// Combines two graphs into one
func (g NewGraph) Combine(h NewGraph) NewGraph {
	if g.Type != h.Type {
		fmt.Println("Graphs must be the same type!")
		return NewGraph{}
	}

	k := NewGraph{
		Nodes: map[string]NewNode{},
		Edges: map[int]NewEdge{},
		Type:  g.Type,
	}

	for key, value := range g.Nodes {
		k.Nodes[key] = value
	}
	for key, value := range h.Nodes {
		k.Nodes[key] = value
	}
	for key, value := range g.Edges {
		k.Edges[key] = value
	}
	for key, value := range h.Edges {
		k.Edges[key] = value
	}
	return k
}

// Returns a string description of a graph
func (g NewGraph) ToString() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("Type: %s\n", g.Type))
	str.WriteString(fmt.Sprintf("Nodes: %d\n", len(g.Nodes)))
	for key, value := range g.Nodes {
		str.WriteString(fmt.Sprintf("- ID: %s, attributes: %+v\n", key, value.Attributes))
	}
	str.WriteString("Edges:\n")
	for _, value := range g.Edges {
		str.WriteString(fmt.Sprintf("- node1_ID: %s, node2_ID: %s, attributes: %+v\n", value.First_node.ID, value.Second_node.ID, value.Attributes))
	}
	return str.String()
}

// Returns a string description of a node
func (g NewGraph) NodeToString(node NewNode) string {
	var str strings.Builder
	str.WriteString("Nodes:\n")
	str.WriteString(fmt.Sprintf("- ID: %s, attributes: %+v\n", node.ID, node.Attributes))
	return str.String()
}

// DFS algorithm
func (g NewGraph) NewDFS(startNode NewNode, visited map[string]bool) NewGraph {
	_, exists := g.Nodes[startNode.ID]
	if !exists {
		return NewGraph{}
	}
	vis := make(map[string]bool)
	visitedGraph := NewGraph{
		Nodes: make(map[string]NewNode),
		Edges: make(map[int]NewEdge),
	}
	g.NewDfsUtil(startNode, vis, &visitedGraph)
	for key, value := range vis {
		visited[key] = value
	}
	return visitedGraph
}

func (g NewGraph) NewDfsUtil(node NewNode, visited map[string]bool, visitedGraph *NewGraph) {
	visited[node.ID] = true
	visitedGraph.AddNode(node)

	for _, neighbor := range g.Neighbors(node) {

		if !visited[neighbor.ID] {
			e, _ := g.GetEdgeByNodes(node, neighbor)
			visitedGraph.AddEdge(NewEdge{First_node: node, Second_node: neighbor, Attributes: e.Attributes})
			g.NewDfsUtil(neighbor, visited, visitedGraph)
		}
	}
}

// returns a slice of all components, separated (each graph in a slice only has one component)
func (g NewGraph) GetComponents() []NewGraph {
	visited := make(map[string]bool)
	for key := range g.Nodes {
		visited[key] = false
	}
	components := []NewGraph{}
	x := GetTrueString(visited)
	for x != "" {
		components = append(components, g.NewDFS(*g.GetNode(x), visited))
		x = GetTrueString(visited)
	}
	return components
}

func GetTrueString(x map[string]bool) string {
	for key, value := range x {
		if !value {
			return key
		}
	}
	return ""
}

// Compares nodes - Will be updated with different versions of graphs
func CompareNodes(n1, n2 NewNode) bool {
	return n1.ID == n2.ID
	/*if n1.ID != n2.ID || len(n1.Attributes) != len(n2.Attributes) {
		return false
	}

	for key, value1 := range n1.Attributes {
		value2, exists := n2.Attributes[key]
		if !exists || value1 != value2 {
			return false
		}
	}
	return true*/
}

// Adds a node to a graph
func (g *NewGraph) AddNode(node NewNode) {
	for key := range g.Nodes {
		if node.ID == key {
			for key, value := range node.Attributes {
				existingvalue, isfound := g.GetNode(node.ID).Attributes[key]
				if isfound {
					v, ok := existingvalue.(int)
					u, okk := value.(int)
					if ok && okk {
						g.GetNode(node.ID).Attributes[key] = v + u
					} else {
						g.GetNode(node.ID).Attributes[key] = value
					}
				} else {
					g.GetNode(node.ID).Attributes[key] = value
				}
			}
			return
		}
	}
	g.Nodes[node.ID] = node
}

// Adds nodes from a slice to a graph
func (g NewGraph) AddNodesFrom(arr []NewNode) NewGraph {
	for _, node := range arr {
		g.AddNode(node)
	}
	return g
}

// Checks if graph already has node n
func (g NewGraph) HasNode(n NewNode) bool {
	for _, node := range g.Nodes {
		if CompareNodes(n, node) {
			return true
		}
	}
	return false
}

// Returns a node in graph with provided id
func (g NewGraph) GetNode(id string) *NewNode {
	for _, node := range g.Nodes {
		if node.ID == id {
			return &node
		}
	}
	return nil
}

// Adds an attribute to a node with corresponding id
func (g NewGraph) AddNodeAttribute(node_id string, att string, value interface{}) NewGraph {
	if g.GetNode(node_id) == nil {
		fmt.Println("this graph does not have node with id:", node_id)
		return g
	}
	g.GetNode((node_id)).Attributes[att] = value
	return g
}

// Returns the number of nodes in graph
func (g NewGraph) NumberOfNodes() int {
	return len(g.Nodes)
}

// Returns a slice, which contains all of the neighbors of a provided node
func (g NewGraph) Neighbors(n NewNode) []NewNode {
	neighbors := []NewNode{}

	for _, edge := range g.Edges {
		if CompareNodes(edge.First_node, n) {
			neighbors = append(neighbors, edge.Second_node)
		} else if CompareNodes(edge.Second_node, n) {
			neighbors = append(neighbors, edge.First_node)
		}
	}

	return neighbors

}

// Removes node from a graph
func (g NewGraph) RemoveNode(n NewNode) NewGraph {
	for i, node := range g.Nodes {
		if CompareNodes(node, n) {
			delete(g.Nodes, i)
			for j, edge := range g.Edges {
				if edge.First_node.ID == i || edge.Second_node.ID == i {
					delete(g.Edges, j)
				}
			}
		}
	}
	return g
}

// Returns the number of neighbors this node has.
func (g NewGraph) NodeDegree(n NewNode) int {
	counter := 0
	for _, edge := range g.Edges {
		if CompareNodes(n, edge.First_node) || CompareNodes(n, edge.Second_node) {
			counter++
		}
	}
	return counter
}

// Deletes a node from a graph and connects all of the neighbours of deleted node
func (g NewGraph) ContractNewNode(n NewNode) {
	if !g.HasNode(n) {
		fmt.Println("This graph does not have this node.")
		return
	}
	neighbors := g.Neighbors(n)
	for i := 0; i < len(neighbors); i++ {
		for j := i + 1; j < len(neighbors); j++ {
			e := NewEdge{
				First_node:  neighbors[i],
				Second_node: neighbors[j],
				Attributes:  map[string]interface{}{},
			}
			g.AddEdge(e)
		}
	}
	g.RemoveNode(n)
}

// Next we have different strategies for combining two nodes
type CombineStrategy interface {
	CombineInt(n1, n2 interface{}) interface{}
	CombineFloat32(n1, n2 interface{}) interface{}
	CombineFloat64(n1, n2 interface{}) interface{}
	CombineString(n1, n2 interface{}) interface{}
}

type StrategyAvgNum struct{}
type StrategyArray struct{}
type StrategyRetainMax struct{}
type StrategyRetainMin struct{}

func (s StrategyRetainMin) CombineInt(v1, v2 interface{}) interface{} {
	n1, _ := v1.(int)
	n2, _ := v2.(int)
	if n1 <= n2 {
		return n1
	}
	return n2
}

func (s StrategyRetainMin) CombineFloat32(v1, v2 interface{}) interface{} {
	n1, _ := v1.(float32)
	n2, _ := v2.(float32)
	if n1 <= n2 {
		return n1
	}
	return n2
}

func (s StrategyRetainMin) CombineFloat64(v1, v2 interface{}) interface{} {
	n1, _ := v1.(float64)
	n2, _ := v2.(float64)
	if n1 <= n2 {
		return n1
	}
	return n2
}

func (s StrategyRetainMin) CombineString(v1, v2 interface{}) interface{} {
	n1, _ := v1.(string)
	n2, _ := v2.(string)
	if n1 <= n2 {
		return n1
	}
	return n2
}

func (s StrategyRetainMax) CombineInt(v1, v2 interface{}) interface{} {
	n1, _ := v1.(int)
	n2, _ := v2.(int)
	if n1 >= n2 {
		return n1
	}
	return n2
}

func (s StrategyRetainMax) CombineFloat32(v1, v2 interface{}) interface{} {
	n1, _ := v1.(float32)
	n2, _ := v2.(float32)
	if n1 >= n2 {
		return n1
	}
	return n2
}

func (s StrategyRetainMax) CombineFloat64(v1, v2 interface{}) interface{} {
	n1, _ := v1.(float64)
	n2, _ := v2.(float64)
	if n1 >= n2 {
		return n1
	}
	return n2
}

func (s StrategyRetainMax) CombineString(v1, v2 interface{}) interface{} {
	n1, _ := v1.(string)
	n2, _ := v2.(string)
	if n1 >= n2 {
		return n1
	}
	return n2
}

func (s StrategyAvgNum) CombineInt(v1, v2 interface{}) interface{} {
	n1, _ := v1.(int)
	n2, _ := v2.(int)
	return (n1 + n2) / 2
}

func (s StrategyAvgNum) CombineFloat32(v1, v2 interface{}) interface{} {
	n1, _ := v1.(float32)
	n2, _ := v2.(float32)
	return (n1 + n2) / 2
}

func (s StrategyAvgNum) CombineFloat64(v1, v2 interface{}) interface{} {
	n1, _ := v1.(float64)
	n2, _ := v2.(float64)
	return (n1 + n2) / 2
}

func (s StrategyArray) CombineInt(v1, v2 interface{}) interface{} {
	if !(reflect.TypeOf(v1).Kind() == reflect.Slice) {
		nv := v1.(int)
		v1 = []int{nv}
	} else {
		n1 := v1
		v1, _ = n1.([]int)
	}
	if !(reflect.TypeOf(v2).Kind() == reflect.Slice) {
		nv := v2.(int)
		v2 = []int{nv}
	} else {
		n2 := v2
		v2 = n2.([]int)
	}
	return append(v1.([]int), v2.([]int)...)
}

func (s StrategyArray) CombineFloat32(n1, n2 interface{}) interface{} {
	if !(reflect.TypeOf(n1).Kind() == reflect.Slice) {
		nv := n1.(float32)
		n1 = []float32{nv}
	} else {
		v := n1
		n1 = v.([]float32)
	}
	if !(reflect.TypeOf(n2).Kind() == reflect.Slice) {
		nv := n2.(float32)
		n2 = []float32{nv}
	} else {
		v := n2
		n2 = v.([]float32)
	}
	return append(n1.([]float32), n2.([]float32)...)
}

func (s StrategyArray) CombineFloat64(n1, n2 interface{}) interface{} {
	if !(reflect.TypeOf(n1).Kind() == reflect.Slice) {
		nv := n1.(float64)
		n1 = []float64{nv}
	} else {
		v := n1
		n1 = v.([]float64)
	}
	if !(reflect.TypeOf(n2).Kind() == reflect.Slice) {
		nv := n2.(float64)
		n2 = []float64{nv}
	} else {
		v := n1
		n1 = v.([]float64)
	}
	return append(n1.([]float64), n2.([]float64)...)
}

func (s StrategyArray) CombineString(n1, n2 interface{}) interface{} {
	if !(reflect.TypeOf(n1).Kind() == reflect.Slice) {
		nv := n1.(string)
		n1 = []string{nv}
	} else {
		v := n1
		n1 = v.([]string)
	}
	if !(reflect.TypeOf(n2).Kind() == reflect.Slice) {
		nv := n2.(string)
		n2 = []string{nv}
	} else {
		v := n1
		n1 = v.([]string)
	}
	return append(n1.([]string), n2.([]string)...)
}

// Combines two nodes based on provided strategies
func (g NewGraph) CombineNodes(n1, n2 NewNode, strat_num, strat_string CombineStrategy) NewNode {
	id := ""
	if n1.ID < n2.ID {
		id += n1.ID + "_" + n2.ID
	} else {
		id += n2.ID + "_" + n1.ID
	}

	att := map[string]interface{}{}
	for key, value := range n1.Attributes {
		att[key] = value
	}
	for key, newValue := range n2.Attributes {
		if existingValue, found := att[key]; found {
			switch v := existingValue.(type) {
			case int:
				if nv, ok := newValue.(int); ok {
					att[key] = strat_num.CombineInt(v, nv)
				}
			case float64:
				if nv, ok := newValue.(float64); ok {
					att[key] = strat_num.CombineFloat64(v, nv)
				}
			case string:
				if nv, ok := newValue.(string); ok {
					strat_string.CombineString(v, nv)
				}
			}
		} else {
			att[key] = newValue
		}
	}

	return NewNode{ID: id, Attributes: att}
}

// Deletes an edge in graph and combines the nodes that it was connecting
func (g *NewGraph) ContractNewEdge(e NewEdge, strategy_num CombineStrategy, strategy_string CombineStrategy) {
	if !g.HasEdge(e) {
		fmt.Println("This graph does not have this edge.")
		return
	}
	n := g.CombineNodes(e.First_node, e.Second_node, strategy_num, strategy_string)
	g.AddNode(n)

	first := g.Neighbors(e.First_node)
	sec := g.Neighbors(e.Second_node)

	for x := 0; x < len(first); x++ {
		g.AddEdge(NewEdge{
			First_node:  n,
			Second_node: first[x],
			Attributes:  map[string]interface{}{},
		})
	}

	for x := 0; x < len(sec); x++ {
		g.AddEdge(NewEdge{
			First_node:  n,
			Second_node: sec[x],
			Attributes:  map[string]interface{}{},
		})
	}

	g.RemoveNode(e.First_node)
	g.RemoveNode(e.Second_node)

}

// returns unique string representing an edge
func (g NewGraph) EdgeID(e NewEdge) string {
	if e.First_node.ID < e.Second_node.ID || g.Type == "digraph" {
		return e.First_node.ID + "-" + e.Second_node.ID
	} else {
		return e.Second_node.ID + "-" + e.First_node.ID
	}
}

// Compares two edges
func (g NewGraph) CompareEdges(e1, e2 NewEdge) bool {
	if !((CompareNodes(e1.First_node, e2.First_node) && CompareNodes(e1.Second_node, e2.Second_node)) || (CompareNodes(e1.First_node, e2.Second_node) && CompareNodes(e1.Second_node, e2.First_node))) {
		return false
	}

	if g.Type != "multigraph" {
		return true
	}

	if len(e1.Attributes) != len(e2.Attributes) {
		return false
	}

	for key, value1 := range e1.Attributes {
		value2, exists := e2.Attributes[key]
		if !exists || value1 != value2 {
			return false
		}
	}
	return true
}

// Tells if a graph elready has this edge
func (g NewGraph) HasEdge(e NewEdge) bool {
	for _, edge := range g.Edges {
		if g.CompareEdges(e, edge) {
			return true
		}
	}
	return false
}

// Adds an edge to the graph
func (g *NewGraph) AddEdge(edge NewEdge) {
	for _, value := range g.Edges {
		if g.CompareEdges(edge, value) {
			for key, value := range edge.Attributes {
				existingvalue, isfound := g.GetEdge(edge).Attributes[key]
				if isfound {
					v, ok := existingvalue.(int)
					u, okk := value.(int)
					if ok && okk {
						g.GetEdge(edge).Attributes[key] = v + u
					} else {
						g.GetEdge(edge).Attributes[key] = value
					}
				} else {
					g.GetEdge(edge).Attributes[key] = value
				}
			}
			return
		}
	}
	if !g.HasNode(edge.First_node) {
		g.AddNode(edge.First_node)
	}
	if !g.HasNode(edge.Second_node) {
		g.AddNode(edge.Second_node)
	}
	g.Edges[len(g.Edges)] = edge
}

// Adds all edges from provided slice to the graph
func (g NewGraph) AddEdgesFrom(arr []NewEdge) NewGraph {
	for _, edge := range arr {
		g.AddEdge(edge)
	}
	return g
}

// Returns edge from graph
func (g NewGraph) GetEdge(e NewEdge) *NewEdge {
	for _, edge := range g.Edges {
		if g.CompareEdges(e, edge) {
			return &edge
		}
	}
	return nil
}

// Returns edge from the graph, that connects these two nodes
func (g NewGraph) GetEdgeByNodes(n1, n2 NewNode) (NewEdge, int) {
	for i, edge := range g.Edges {
		if (CompareNodes(edge.First_node, n1) && CompareNodes(edge.Second_node, n2)) || (CompareNodes(edge.First_node, n2) && CompareNodes(edge.Second_node, n1)) {
			return edge, i
		}
	}
	fmt.Println("this graph does not have edge between these two nodes.")
	return NewEdge{First_node: n1, Second_node: n2, Attributes: map[string]interface{}{}}, len(g.Edges)
}

// Adds an attribute to the edge
func (g NewGraph) AddEdgeAttribute(e NewEdge, att string, value interface{}) NewGraph {
	if g.GetEdge(e) == nil {
		fmt.Println("this graph does not have this edge:", e)
		return g
	}
	g.GetEdge(e).Attributes[att] = value
	return g
}

// Returns the number of edges that this graph has
func (g NewGraph) NumberOfEdges() int {
	return len(g.Edges)
}

// Removes edge from a graph
func (g NewGraph) RemoveEdge(e NewEdge) NewGraph {
	for i, edge := range g.Edges {
		if g.CompareEdges(e, edge) {
			delete(g.Edges, i)
		}
	}
	return g
}
