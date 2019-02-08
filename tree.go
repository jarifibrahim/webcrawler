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

	if _, err := file.Write([]byte(node.GenerateTree(0))); err != nil {
		log.Error(err)
	}

}

// GenerateTree returns the ...
func (node *URLNodeType) GenerateTree(tabSize int) string {
	// Generate items in depth first search manner
	subTree := ""
	for _, child := range node.children {
		line := strings.Repeat("\t", tabSize)
		line += "└── "
		subTree += line + child.GenerateTree(tabSize+1)
	}
	return node.url + "\n" + subTree
}
