iptree
======

`iptree` makes us to easily construct the idea of IPDB, but I have doubted if it was useful while developing. The initial idea comes to me was that I was going to make static rules for China IP addresses, specially routed. To improve performance, the program was not going to use the common implementation, but a new one that cut (or simplified) the branches. See below.

I would like to hear ideas about `iptree`, as many as possible, to provide something more useful.

```go
package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/GreenYun/ipdb"
)

func main() {
	root, err := initTree()
	if err != nil {
		log.Fatal(err)
	}
	getCNSegments(root, uint32(0), 0)
}

func initTree() (*node, error) {
	db, err := ipdb.NewIPDb("/path/to/db")
	if err != nil {
		return nil, err
	}

	root, err := getRoot(db)
	if err != nil {
		return nil, err
	}
	return root, nil
}

func getCNSegments(n *node, ip uint32, depth int) {
	if n == nil {
		return
	}
	if v, ok := n.Left.(*leaf); ok && v != nil {
		if v.IsCn {
			netAddr := make(net.IP, 4)
			binary.BigEndian.PutUint32(netAddr, ip)
			fmt.Printf("%s/%d\n", netAddr, depth+1)
		}
	} else {
		if x, ok := n.Left.(*node); ok && x != nil {
			getCNSegments(x, ip, depth+1)
		} else {
			return
		}
	}
	if v, ok := n.Right.(*leaf); ok && v != nil {
		if v.IsCn {
			netAddr := make(net.IP, 4)
			binary.BigEndian.PutUint32(netAddr, ip+uint32(1)<<(32-depth-1))
			fmt.Printf("%s/%d\n", netAddr, depth+1)
		}
	} else {
		if x, ok := n.Right.(*node); ok && x != nil {
			getCNSegments(x, ip+uint32(1)<<(32-depth-1), depth+1)
		} else {
			return
		}
	}
}

type node struct {
	Left, Right interface{}
}

type leaf struct {
	IsCn bool
}

var (
	errNotLeaf = errors.New("Not a Leaf at that position")
	errNotNode = errors.New("Not a Node at that position")
)

func getRoot(d *ipdb.IPDb) (*node, error) {
	nodeCount := d.Metadata.NodeCount
	offset := uint32(0)
	if d.IsIPv4Db() {
		for i := 0; i < 96 && offset < nodeCount; i++ {
			if i >= 80 {
				pos := offset*8 + 4
				offset = binary.BigEndian.Uint32(d.Data[pos : pos+4])
			} else {
				pos := offset * 8
				offset = binary.BigEndian.Uint32(d.Data[pos : pos+4])
			}
		}
	} else {
		if !d.IsIPv6Db() {
			return nil, fmt.Errorf("file error")
		}
	}

	node, err := getNode(d, offset)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func getNode(d *ipdb.IPDb, offset uint32) (*node, error) {
	nodeCount := d.Metadata.NodeCount
	if offset > nodeCount {
		return nil, errNotNode
	}

	pos := offset * 8
	leftOffset := binary.BigEndian.Uint32(d.Data[pos : pos+4])
	pos += 4
	rightOffset := binary.BigEndian.Uint32(d.Data[pos : pos+4])

	var left, right interface{}

	var err error
	if leftOffset > nodeCount {
		left, err = getLeaf(d, leftOffset)
		if err != nil {
			return nil, err
		}
	} else {
		v, err := getNode(d, leftOffset)
		if err != nil {
			return nil, err
		}

		left = v
		if v != nil {
			if l, ok := v.Left.(*leaf); ok {
				if r, ok := v.Right.(*leaf); ok {
					if l == nil && r == nil {
						left = nil
					} else {
						if l != nil && r != nil && l.IsCn == r.IsCn {
							left = &leaf{
								IsCn: l.IsCn,
							}
						}
					}
				}
			}
		}
	}

	if rightOffset > nodeCount {
		right, err = getLeaf(d, rightOffset)
		if err != nil {
			return nil, err
		}
	} else {
		v, err := getNode(d, rightOffset)
		if err != nil {
			return nil, err
		}

		right = v
		if v != nil {
			if l, ok := v.Left.(*leaf); ok {
				if r, ok := v.Right.(*leaf); ok {
					if l == nil && r == nil {
						right = nil
					} else {
						if l != nil && r != nil && l.IsCn == r.IsCn {
							right = &leaf{
								IsCn: l.IsCn,
							}
						}
					}
				}
			}
		}
	}

	if left == nil && right == nil {
		return nil, nil
	}

	n := &node{
		Left:  left,
		Right: right,
	}
	return n, nil
}

func getLeaf(d *ipdb.IPDb, offset uint32) (*leaf, error) {
	if offset <= d.Metadata.NodeCount {
		return nil, errNotLeaf
	}

	dbDataLen := uint32(len(d.Data))

	pos := d.Metadata.NodeCount*7 + offset
	if pos >= dbDataLen {
		return nil, fmt.Errorf("file error")
	}

	len := binary.BigEndian.Uint16(d.Data[pos : pos+2])
	if pos+2+uint32(len) > dbDataLen {
		return nil, fmt.Errorf("file error")
	}

	data := d.Data[pos+2 : pos+2+uint32(len)]
	cc := ""
	zeroCount := 0
	for _, b := range data {
		if b == 9 {
			zeroCount++
			continue
		}
		if zeroCount == 11 {
			cc += string(b)
		}
	}

	if cc == "" {
		return nil, nil
	}

	leaf := &leaf{
		IsCn: strings.ToUpper(cc) == "CN",
	}

	return leaf, nil
}
```
