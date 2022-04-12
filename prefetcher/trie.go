package prefetcher

import (
	"fmt"
)

var test = []string{
	"D",
	"E",
	"F",
	"A",
	"B",
	"C",
	"D",
	"E",
	"F",
	"G",
	"D",
	"E",
	"F",
	"K",
}

type node struct {
	URI        string
	AccessFreq int

	// TODO change this to a linked list?
	Children []*node
}

func newNode(URI string) *node {
	return &node{URI: URI, AccessFreq: 0}
}

func (n *node) getChild(URI string) *node {
	for _, child := range n.Children {
		if child.URI == URI {
			return child
		}
	}
	return nil
}

func (n *node) addChild(child *node) {
	n.Children = append(n.Children, child)
}

// TODO remove
func (n *node) printChilderen() {
	for _, c := range n.Children {
		fmt.Println(c.URI)
	}
}

// TODO naming; this is a bit more than just a trie
/* suffix trie */
type Trie struct {
	root  *node
	trace circulairArray
}

func NewTrie() *Trie {
	return &Trie{
		root:  nil,
		trace: newCirculairArray(),
	}
}

func (t *Trie) ProcessRequest(URI string) {
	t.trace.pushBack(URI)

	t.PredictNext()
}

func (t *Trie) BuildTrie() {

	t.root = newNode("root")

	for i := 0; i < len(test); i++ {
		t.addSuffix(&test, i)
	}
}

func (t *Trie) addSuffix(trace *[]string, pos int) {

	currentNode := t.root
	var nextNode *node

	// TODO is using slices efficient enough?
	for _, URI := range (*trace)[pos:] {
		nextNode = currentNode.getChild(URI)

		if nextNode == nil {
			nextNode = newNode(URI)
			currentNode.addChild(nextNode)
		}

		currentNode = nextNode
	}
}

func (t *Trie) PredictNext() {
	currentNode := t.root

	fmt.Println(t.trace.getAll())
	for _, URI := range t.trace.getAll() {
		currentNode = currentNode.getChild(URI)
		if currentNode == nil {
			fmt.Println("pattern not matched")
			return
		}
	}

	currentNode.printChilderen()
}
