package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
