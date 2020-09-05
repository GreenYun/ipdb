package ipdb

import (
	"encoding/binary"
	"errors"

	"github.com/GreenYun/ipdb/iptree"
)

var (
	errNotLeaf = errors.New("Not a Leaf at that position")
	errNotNode = errors.New("Not a Node at that position")
)

// GrowTree return the root Node of the iptree.
func (d *IPDb) GrowTree() (*iptree.Node, error) {
	node, err := d.getNode(d.startFrom)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (d *IPDb) getNode(offset uint32) (*iptree.Node, error) {
	nodeCount := d.Metadata.NodeCount
	if offset > nodeCount {
		return nil, errNotNode
	}

	pos := offset * 8
	leftOffset := binary.BigEndian.Uint32(d.Data[pos : pos+4])
	pos += 4
	rightOffset := binary.BigEndian.Uint32(d.Data[pos : pos+4])

	node := new(iptree.Node)

	var err error
	if leftOffset >= nodeCount {
		node.Left, err = d.getLeaf(leftOffset)
		if err != nil {
			return nil, err
		}
	} else {
		node.Left, err = d.getNode(leftOffset)
		if err != nil {
			return nil, err
		}
	}

	if rightOffset >= nodeCount {
		node.Right, err = d.getLeaf(rightOffset)
		if err != nil {
			return nil, err
		}
	} else {
		node.Right, err = d.getNode(rightOffset)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

func (d *IPDb) getLeaf(offset uint32) (*iptree.Leaf, error) {
	data, err := d.GetRaw(offset)
	if err != nil {
		return nil, err
	}

	leaf := &iptree.Leaf{
		Data: data,
	}
	return leaf, nil
}
