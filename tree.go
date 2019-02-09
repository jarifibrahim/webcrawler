package main

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	ErrNodeNil = "URlNode cannot be nil"
)

// URLNodeType represents a Node in the URL tree
type URLNodeType struct {
	url      string         // the actual URL
	children []*URLNodeType // all URLs reachable from actual URL
}

// NewNode return a new node of URLNodeType
func NewNode(url string) *URLNodeType {
	return &URLNodeType{url: url}
}

// AddChild adds a new child to the given node.
// Returns the newly created child
func (node *URLNodeType) AddChild(childURL string) (*URLNodeType, error) {
	if node == nil {
		return nil, errors.New(ErrNodeNil)
	}
	newChild := URLNodeType{url: childURL}
	node.children = append(node.children, &newChild)
	return &newChild, nil
}

func (node *URLNodeType) WriteTreeToFile(file *os.File) {

	if _, err := file.Write([]byte(node.GenerateTree())); err != nil {
		log.Error(err)
	}

}

func (node *URLNodeType) GenerateTree() string {
	return node.generateTree(0)
}

// GenerateTree returns the ...
func (node *URLNodeType) generateTree(tabSize int) string {
	// Generate items in depth first search manner
	subTree := ""
	for _, child := range node.children {
		line := strings.Repeat("\t", tabSize)
		line += "└── "
		subTree += line + child.generateTree(tabSize+1)
	}
	return node.url + "\n" + subTree
}
