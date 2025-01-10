package citation

/*
This file extracts data for citation network from turtle (.ttl) file. It not only extracts connections between different
citations, but also extracts some of the key elements from each link. It extracts keywords and text which is transformed
and saved as an embedding.
*/

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/nvkp/turtle"
)

type Triple struct {
	Subject   string `turtle:"subject"`
	Predicate string `turtle:"predicate"`
	Object    string `turtle:"object"`
}

func OnPage(link string) map[string]interface{} {
	res, err := http.Get(link)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	// Read the response body
	content, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Create a map to store the parsed JSON
	var result map[string]interface{}

	// Parse the JSON into the map
	err = json.Unmarshal(content, &result)
	if err != nil {
		log.Fatal("Error parsing JSON:", err)
	}

	return result
}

func reconstructText(wordPositions map[string]interface{}) string {
	// Slice to hold words in order of positions
	var words []string

	// Collect positions and words
	positionWordMap := make(map[float64]string)
	for word, positions := range wordPositions {
		positions := positions.([]interface{})
		for _, pos := range positions {
			x := pos.(float64)
			positionWordMap[x] = word
		}
	}

	// Sort positions and build the ordered text
	var positions []float64
	for pos := range positionWordMap {
		positions = append(positions, pos)
	}
	sort.Float64s(positions)

	for _, pos := range positions {
		words = append(words, positionWordMap[pos])
	}

	return strings.Join(words, " ")
}

func GetKeywords(api_map map[string]interface{}) []string {
	x := api_map["keywords"]
	var words []string

	if intSlice, ok := x.([]interface{}); ok {
		for _, value := range intSlice {
			if slice, ok := value.(map[string]interface{}); ok {
				str, ok := slice["display_name"].(string)
				if ok {
					words = append(words, str)
				}
			}
		}
	} else {
		fmt.Println("x is not a slice of ints")
	}
	return words
}

func GetText(api_map map[string]interface{}) []float64 {
	inverted_text, ok := api_map["abstract_inverted_index"].(map[string]interface{})
	var text string
	if ok {
		text = reconstructText(inverted_text)
	}

	cmd := exec.Command("python", "text_embedding_model.py", text)
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Error getting text embedding: ", err)
	}

	outputStr := strings.TrimSpace(string(output))

	outputStr = strings.ReplaceAll(outputStr, "[", "")
	outputStr = strings.ReplaceAll(outputStr, "]", "")

	parts := strings.Split(outputStr, ", ")

	var result []float64
	for _, part := range parts {
		value, err := strconv.ParseFloat(part, 64)
		if err != nil {
			log.Fatal("Error transformint string to float64: ", err)
		}
		result = append(result, value)
	}
	return result
}

func GetNodeAttributes(api_map map[string]interface{}) map[string]interface{} {
	attr := map[string]interface{}{}
	attr["keywords"] = GetKeywords(api_map)
	attr["text_embedding"] = GetText(api_map)
	return attr
}

func Extract(file_name string) string {
	file, err := os.Open(file_name) //"citation_network_tiny.ttl"
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close()

	var triples []Triple

	byteFile, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error converting file:", err)
		return ""
	}
	err = turtle.Unmarshal(byteFile, &triples)
	if err != nil {
		fmt.Println("Error unmarshaling Turtle data:", err)
		return ""
	}

	allAttributes := make(map[string]map[string]interface{})
	for _, triple := range triples {
		_, ok1 := allAttributes[triple.Subject]
		_, ok2 := allAttributes[triple.Object]
		if !ok1 {
			str := strings.Replace(triple.Subject, "https://semopenalex.org/work", "https://api.openalex.org/works", 1)
			api_map := OnPage(str)
			allAttributes[triple.Subject] = /*map[string]interface{}{}*/ GetNodeAttributes(api_map)
		}

		if !ok2 {
			str := strings.Replace(triple.Object, "https://semopenalex.org/work", "https://api.openalex.org/works", 1)
			api_map := OnPage(str)
			allAttributes[triple.Object] = /*map[string]interface{}{}*/ GetNodeAttributes(api_map)
		}
		_, exists := allAttributes[triple.Subject]["neighbors"]

		if !exists {
			var neighbors []string
			allAttributes[triple.Subject]["neighbors"] = neighbors
		}

		allAttributes[triple.Subject]["neighbors"] = append(allAttributes[triple.Subject]["neighbors"].([]string), triple.Object)
	}

	newfile_name := "citation_network_tiny_extracted_data.json"

	newfile, err := os.Create(newfile_name)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return ""
	}
	defer newfile.Close()

	encoder := json.NewEncoder(newfile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(allAttributes); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return ""
	}

	fmt.Println("Data extraction succesful!")
	return newfile_name
}
