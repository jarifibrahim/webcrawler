package tree

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	ErrNodeNil = "URlNode cannot be nil"
)

// URLNode represents a Node in the URL tree
type URLNode struct {
	url      string     // the actual URL
	children []*URLNode // all URLs reachable from actual URL
}

// NewNode return a new node of URLNode
func NewNode(url string) *URLNode {
	return &URLNode{url: url}
}

// AddChild adds a new child to the given node.
// Returns the newly created child
func (node *URLNode) AddChild(childURL string) *URLNode {
	// Don't add child node if root node is nil
	// The root node (and all the following nodes) will be nil if --show-tree
	// flag is set to false
	if node == nil {
		return nil
	}
	newChild := URLNode{url: childURL}
	node.children = append(node.children, &newChild)
	return &newChild
}

func (node *URLNode) WriteTreeToFile(file *os.File) {
	if _, err := file.Write([]byte(node.GenerateTree())); err != nil {
		log.Error(err)
	}

}

func (node *URLNode) GenerateTree() string {
	return node.generateTree(0)
}

// GenerateTree returns the ...
func (node *URLNode) generateTree(tabSize int) string {
	// Generate items in depth first search manner
	subTree := ""
	for _, child := range node.children {
		line := strings.Repeat("\t", tabSize)
		line += "└── "
		subTree += line + child.generateTree(tabSize+1)
	}
	return node.url + "\n" + subTree
}
