package main

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Disable logger. We don't want noisy logs when running tests
func init() {
	logrus.SetLevel(logrus.PanicLevel)

}
func TestAddChild(t *testing.T) {
	rootURL := "foo"
	root := &URLNodeType{url: rootURL}
	validateNewChild(t, root, rootURL)

	t.Run("Success", func(t *testing.T) {
		// root shouldn't have any children
		t.Run("add child 1", func(t *testing.T) {
			// Root
			//  | - Child
			childURL := "child of foo"
			child, err := root.AddChild(childURL)
			assert.Nil(t, err)
			validateNewChild(t, child, childURL)
			// check if new child was added
			assert.Len(t, root.children, 1)
			assert.Equal(t, root.children[0], child)
			assert.Equal(t, root.children[0].url, childURL)

			t.Run("add grandchild", func(T *testing.T) {
				// Root
				//  | - Child
				//    | - GrandChild
				grandChildURL := "grandchild of foo"
				grandChild, err := child.AddChild(grandChildURL)
				assert.Nil(t, err)

				validateNewChild(t, grandChild, grandChildURL)
				// check if new grandchild child was added to child
				assert.Len(t, child.children, 1)
				assert.Equal(t, child.children[0], grandChild)
				assert.Equal(t, child.children[0].url, grandChildURL)

				// check if new grandchild child was not added root
				// root should have only one child
				assert.Len(t, root.children, 1)
				assert.Equal(t, root.children[0], child)

			})

			t.Run("add sibling of child1", func(t *testing.T) {
				// Root
				//  | - Child 1
				//  | - Child 2
				siblingURL := "sibling of foo"
				siblingOfChild, err := root.AddChild(siblingURL)
				assert.Nil(t, err)
				validateNewChild(t, siblingOfChild, siblingURL)

				// Root should have 2 children
				assert.Len(t, root.children, 2)
				assert.Equal(t, root.children[0], child)
				assert.Equal(t, root.children[1], siblingOfChild)
			})
		})

	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Add Child to nil node", func(t *testing.T) {
			var nilNode *URLNodeType
			child, err := nilNode.AddChild("foo child")
			assert.Error(t, err, ErrNodeNil)
			assert.Nil(t, child)

		})
	})
}

// validateNewChild checks if the newly created child has correct URL and
func validateNewChild(t *testing.T, node *URLNodeType, nodeURL string) {
	assert.Nil(t, node.children)
	assert.Equal(t, node.url, nodeURL)
	assert.Len(t, node.children, 0)
}
