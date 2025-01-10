package citation

/*
This file creates graph from already extracted data +, that is saved in a map[string]map[string]interface{} and exported to json file.
It returns completed graph.
*/

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jmCodeCraft/go-network/model"
)

func Create_graph(file_name string) *model.NewGraph {
	file, err := os.Open(file_name)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	// Declare a map to hold the data
	var all_atribute_map map[string]map[string]interface{}

	// Decode the JSON data into the map
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&all_atribute_map); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return nil
	}

	g := model.NewGraph{
		Nodes: map[string]model.NewNode{},
		Edges: map[int]model.NewEdge{},
	}

	for key, value := range all_atribute_map {
		g.AddNode(model.NewNode{
			ID:         key,
			Attributes: value,
		})
	}

	for _, node := range g.Nodes {
		if neighbors, ok := node.Attributes["neighbors"].([]string); ok {
			for _, neighbor := range neighbors {
				g.AddEdge(model.NewEdge{
					First_node:  node,
					Second_node: g.Nodes[neighbor],
					Attributes:  map[string]interface{}{},
				})
			}
		}
		delete(node.Attributes, "neighbor")
	}

	fmt.Println("Graph creation successful!")

	return &g

}
