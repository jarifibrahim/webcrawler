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
	// validateNewChild checks if the newly created child has correct URL and
	validateNewChild := func(t *testing.T, node *URLNodeType, nodeURL string) {
		assert.Nil(t, node.children)
		assert.Equal(t, node.url, nodeURL)
		assert.Len(t, node.children, 0)
	}

	rootURL := "foo"
	root := NewNode(rootURL)

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

func TestGenerateTree(t *testing.T) {
	root := NewNode("root")
	// createChild creates a new child and returns it.
	// It also asserts the returned error to be nil
	createChild := func(t *testing.T, parent *URLNodeType, child string) *URLNodeType {
		createdChild, err := parent.AddChild(child)
		assert.Nil(t, err)
		return createdChild
	}
	t.Run("no children", func(t *testing.T) {
		assert.Equal(t, "root\n", root.GenerateTree())
	})
	t.Run("one child", func(t *testing.T) {
		/*
			root
			  |- child
		*/
		child1 := createChild(t, root, "child1")
		assert.Equal(t, "root\n└── child1\n", root.GenerateTree())
		t.Run("two children", func(t *testing.T) {
			/*
				root
				  | - child 1
				  | - child 2
			*/
			createChild(t, root, "child2")
			assert.Equal(t, "root\n└── child1\n└── child2\n", root.GenerateTree())
		})
		t.Run("grandchild of first child", func(t *testing.T) {
			/*
				root
				  | - child 1
				    | - grandchild 1
				  | - child 2
			*/
			createChild(t, child1, "grandchild 1")
			assert.Equal(t, "root\n└── child1\n\t└── grandchild 1\n└── child2\n", root.GenerateTree())
		})
	})
}
