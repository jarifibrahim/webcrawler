package tree

import (
	"io"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// URLNode represents a Node in the URL tree
type URLNode struct {
	url      string     // the actual URL
	children []*URLNode // all URLs reachable from actual URL
	sync.Mutex
}

// NewNode return a new node of URLNode
func NewNode(url string) *URLNode {
	return &URLNode{url: url}
}

// AddChild adds a new child to the given node.
// Returns the newly created child node
func (node *URLNode) AddChild(childURL string) *URLNode {
	// Don't add child node if root node is nil
	// The root node (and all the following nodes) will be nil if --show-tree
	// flag is set to false
	if node == nil {
		return nil
	}
	node.Lock()
	newChild := URLNode{url: childURL}
	node.children = append(node.children, &newChild)
	node.Unlock()
	return &newChild
}

// WriteTree generates the tree and writes it to the writer.
func (node *URLNode) WriteTree(writer io.Writer) {
	if _, err := writer.Write([]byte(node.GenerateTree())); err != nil {
		log.Error(err)
	}
}

// GenerateTree tree builds the tree which shows links between nodes.
// Returns the complete tree pointed to by node.
func (node *URLNode) GenerateTree() string {
	return node.generateTree(0)
}

// generate tree is recursively called to build the tree.
// It generates items in depth first search manner.
func (node *URLNode) generateTree(tabSize int) string {
	subTree := ""
	for _, child := range node.children {
		line := strings.Repeat("\t", tabSize)
		line += "└── "
		subTree += line + child.generateTree(tabSize+1)
	}
	return node.url + "\n" + subTree
}
