package main

import (
	"github.com/satori/go.uuid"
	"math/rand"
)

// Generates a random graph. The graph has a max size of size.
// Each node can have a random number of up to size/2 children.
// The graph generated is not guaranteed to be fully connected.
func generateRandomGraph(size int, seed int64) []*LabelNode {
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

type Queue struct {
	V []*LabelNode
}

func (a *Queue) Enqueue(x *LabelNode) {
	a.V = append(a.V, x)
}

func (a *Queue) Dequeue() *LabelNode {
	if len(a.V) < 1 {
		return nil
	}
	x := a.V[0]
	a.V = a.V[1:]
	return x
}

// Generates a B-Tree structure with each node having at least 1 child and at most
// a max branchingFactor provided.
// The algorithm to generate a random tree is to use a queue to enqueue each node
// generated to generate it's children. This allows us to have a more balanced tree.
func generateRandomTree(branchingFactor, size int, seed int64) []*LabelNode {
	r := rand.New(rand.NewSource(seed))
	queue := new(Queue)

	numNodes := r.Intn(size)
	var nodes []*LabelNode = []*LabelNode{
		&LabelNode{
			Id:    0,
			Label: uuid.NewV4().String()},
	}
	queue.Enqueue(nodes[0])

	nodesSoFar := 1
	for len(nodes) < numNodes {
		node := queue.Dequeue()
		if node == nil {
			break
		}
		numChildren := r.Intn(branchingFactor) + 1
		for i := 0; i < numChildren; i++ {
			node.Children = append(node.Children, nodesSoFar)
			newNode := &LabelNode{
				Id:    nodesSoFar,
				Label: uuid.NewV4().String()}
			nodesSoFar++
			nodes = append(nodes, newNode)
			queue.Enqueue(newNode)
		}
	}
	return nodes
}

// Generates a B-Tree structure similarly to generateRandomTree but with Rules
func generateRandomTreeWithRules(branchingFactor, size int, seed int64) []*LabelNode {
	r := rand.New(rand.NewSource(seed))
	queue := new(Queue)

	numNodes := r.Intn(size)
	var nodes []*LabelNode = []*LabelNode{
		&LabelNode{
			Id:    0,
			Label: uuid.NewV4().String()},
	}
	queue.Enqueue(nodes[0])

	nodesSoFar := 1
	for len(nodes) < numNodes {
		node := queue.Dequeue()
		if node == nil {
			break
		}
		numChildren := r.Intn(branchingFactor) + 1
		for i := 0; i < numChildren; i++ {
			node.Children = append(node.Children, nodesSoFar)
			newNode := &LabelNode{
				Id:    nodesSoFar,
				Label: uuid.NewV4().String()}
			if r.Int()%10 == 0 {
				for j := 0; j < r.Intn(branchingFactor); j++ {
					newNode.Rule = append(newNode.Rule, nodes[r.Intn(nodesSoFar)].Label)
				}
			}
			nodesSoFar++
			nodes = append(nodes, newNode)
			queue.Enqueue(newNode)
		}
	}
	return nodes
}
