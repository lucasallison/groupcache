package prefetcher

// TODO visability
import (
	"fmt"
)

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
		fmt.Print(c.URI, " ")
	}
	fmt.Println(" ")
}

// TODO naming; this is a bit more than just a trie
/* suffix trie */
type Trie struct {
	root        *node
	curentTrace circulairArray
	savedTrace  []string
	uid         string
}

func NewTrie(uid string) *Trie {
	t := Trie{
		root:        nil,
		curentTrace: newCirculairArray(),
		uid:         uid,
	}
	t.buildTrie()
	return &t
}

func (t *Trie) SaveTrie() error {
	return writeUserTrace(t.uid, t.curentTrace.getDataAsSlice())
}

func (t *Trie) ProcessRequest(URI string) {
	t.curentTrace.pushBack(URI)

	t.predictNext(true)
}

func (t *Trie) buildTrie() {

	t.savedTrace, _ = getUserTrace(t.uid)

	t.root = newNode("root")

	for i := 0; i < len(t.savedTrace); i++ {
		t.addSuffix(i)
	}
}

func (t *Trie) addSuffix(pos int) {

	currentNode := t.root
	var nextNode *node

	// TODO is using slices efficient enough? see also predict next
	for _, URI := range (t.savedTrace)[pos:] {
		nextNode = currentNode.getChild(URI)

		if nextNode == nil {
			nextNode = newNode(URI)
			currentNode.addChild(nextNode)
		}

		currentNode = nextNode
	}
}

// TODO refactor
func (t *Trie) predictNext(recursive bool) {

	trace := t.curentTrace.getDataAsSlice()
	var currentNode *node
	var matched bool

	fmt.Println(t.curentTrace.getDataAsSlice())

	for i := 0; i < len(trace); i++ {

		if len(trace)-i < MIN_URI_MATCHES {
			break
		}

		currentNode = t.root
		matched = true

		fmt.Println("Matching: ", trace[i:])
		for _, URI := range trace[i:] {
			currentNode = currentNode.getChild(URI)
			if currentNode == nil {
				matched = false
				break
			}

		}

		if !recursive || matched {
			fmt.Print("Matched in ", i, "th iteration")
			break
		}
	}

	if currentNode == nil {
		fmt.Println("pattern not matched")
	} else {
		currentNode.printChilderen()
	}
}
