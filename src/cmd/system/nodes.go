package main

import (
	"encoding/json"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"log"
	"math/rand"
)

type LabelNode struct {
	Id       int      `json:id`
	Label    string   `json:label`
	Rule     []string `json:rule,omitempty`
	Children []int    `json:children`
	Visited  bool
}

func load(inputFile string) []*LabelNode {
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	var nodes []*LabelNode
	err = json.Unmarshal(data, &nodes)
	if err != nil {
		log.Fatal(err)
	}
	return nodes
}

func output(graph []*LabelNode, outputFile string) {
	b, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(outputFile, b, 0777)
}

func generateRandom(size int, seed int64) []*LabelNode {
	r := rand.New(rand.NewSource(seed))

	numNodes := r.Intn(size)
	var nodes []*LabelNode
	var labels []string
	for i := 0; i < numNodes; i++ {
		newNode := &LabelNode{
			Id:    i,
			Label: uuid.NewV4().String(),
		}
		numChildren := r.Intn(numNodes) / 2
		for j := 0; j < numChildren; j++ {
			newNode.Children = append(newNode.Children, r.Intn(numNodes))
		}
		labels = append(labels, newNode.Label)
		nodes = append(nodes, newNode)
	}
	return nodes
}
