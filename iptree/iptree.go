package iptree

// Leaf refers to the very end of a branch of the tree, contains the raw data.
type Leaf struct {
	Data []byte
}

// Node refers to a binary tree node, either left or right might be one of Node
// or Leaf.
type Node struct {
	Left, Right interface{}
}
