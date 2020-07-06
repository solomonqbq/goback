package main

import (
	"errors"
	"fmt"
	"github.com/cespare/xxhash/v2"
	"hash"
	"hash/adler32"
	"hash/crc32"
	"hash/fnv"
	"math"
	"sort"
)

func main() {
	nodeIds := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ring := newRing(nodeIds, crc32.NewIEEE())
	for i := 1; i < 200; i++ {
		ring.addNode(10 + i)
	}

	test(newRing(nodeIds, crc32.NewIEEE()))
	test(newRing(nodeIds, fnv.New32()))
	test(newRing(nodeIds, adler32.New()))
	test(newRing(nodeIds, &xxhashAdapter{}))

}

type xxhashAdapter struct {
	xxhash.Digest
}

func (a *xxhashAdapter) Sum32() uint32 {
	return uint32(a.Sum64())
}

func test(ring *Ring) {
	keyCount := 1000000
	counter := make(map[int]int)
	for i := 0; i < keyCount; i++ {
		n, _ := ring.getNode(fmt.Sprintf("%s%d", "key", i))
		if counter[n] == 0 {
			counter[n] = 1
		} else {
			counter[n] = counter[n] + 1
		}
	}
	fmt.Println(counter)
	//计算标准差
	average := keyCount / len(ring.actualNodes)
	var sum float64
	for _, v := range counter {
		sum += math.Pow(float64(v)-float64(average), 2)
	}
	fmt.Printf(fmt.Sprintf("%T 标准差:%v\n", ring.hash, math.Sqrt(sum)))
}

type Ring struct {
	size            int
	virtualCount    int
	virtualNodes    []*Node
	virtualNodesMap map[int][]*Node
	actualNodes     []*Node
	step            int
	hash            hash.Hash32
}

type Node struct {
	id    int
	index int
}

func newRing(nodeIds []int, hash hash.Hash32) *Ring {
	ring := &Ring{
		virtualCount:    200,
		virtualNodes:    make([]*Node, 0),
		actualNodes:     make([]*Node, 0),
		virtualNodesMap: map[int][]*Node{},
		hash:            hash,
		size:            1000000,
	}
	for nodeId := range nodeIds {
		ring.actualNodes = append(ring.actualNodes, &Node{id: nodeId})
	}
	ring.step = 1000000 / ring.virtualCount
	index := 0
	for i := 0; i < ring.virtualCount; {
		for _, nodeId := range nodeIds {
			index = i * ring.step
			n := &Node{id: nodeId, index: index}
			ring.virtualNodesMap[n.id] = append(ring.virtualNodesMap[n.id], n)
			i++
		}
	}
	ring.rebuildVirtualNodes()
	return ring
}

func (r *Ring) getNode(key string) (node int, error error) {
	if len(r.virtualNodes) == 0 {
		return 0, errors.New("no actual nodes")
	}
	r.hash.Reset()
	r.hash.Write([]byte(key))
	//fmt.Println(hash.Sum32())
	index := int(r.hash.Sum32() % uint32(r.size))
	//fmt.Println(index)
	for _, an := range r.virtualNodes {
		if index > an.index {
			continue
		} else {
			return an.id, nil
		}
	}
	return r.virtualNodes[0].id, nil
}

func (r *Ring) addNode(newNodeId int) error {
	if len(r.actualNodes) == 0 {
		return errors.New("no actual nodes")
	}
	//rebalance的逻辑:新增节点讨'百家饭'，从'大户'开始吃,一家一口，吃到新节点virtual node count达到virtual nodes count/actual nodes count均值为止
	//不考虑并发
	newNodes := make([]*Node, 0)
	nodeSorter := r.getNodeSorter()
	sort.Sort(sort.Reverse(nodeSorter))
	//sort.Sort(nodeSorter)
	count := len(r.virtualNodes) / (len(r.actualNodes) + 1) / len(r.actualNodes)
	for i := 0; i < count; i++ {
		for _, nodes := range nodeSorter {
			n := &Node{newNodeId, nodes[i].index}
			newNodes = append(newNodes, n)
			r.virtualNodesMap[nodes[i].id] = r.virtualNodesMap[nodes[i].id][i+1:]
		}
	}
	count = len(r.virtualNodes) / (len(r.actualNodes) + 1) % len(r.actualNodes)
	for i, nodes := range nodeSorter {
		if i < count {
			n := &Node{newNodeId, nodes[0].index}
			newNodes = append(newNodes, n)
			r.virtualNodesMap[nodes[0].id] = r.virtualNodesMap[nodes[0].id][1:]
		} else {
			break
		}
	}

	r.actualNodes = append(r.actualNodes, &Node{id: newNodeId})
	r.virtualNodesMap[newNodeId] = newNodes
	r.rebuildVirtualNodes()
	return nil
}

func (r *Ring) rebuildVirtualNodes() {
	nodes := make([]*Node, 0)
	for _, v := range r.virtualNodesMap {
		nodes = append(nodes, v...)
	}
	sort.Sort(VirtualNodeSorter(nodes))
	r.virtualNodes = nodes
}

type NodeSorter [][]*Node

func (r *Ring) getNodeSorter() NodeSorter {
	virtualNodes := make([][]*Node, 0)
	for _, vn := range r.virtualNodesMap {
		virtualNodes = append(virtualNodes, vn)
	}
	return virtualNodes
}

// Len is the number of elements in the collection.
func (ns NodeSorter) Len() int {
	return len(ns)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (ns NodeSorter) Less(i, j int) bool {
	if len(ns[i]) == len(ns[j]) && len(ns[i]) != 0 && len(ns[j]) != 0 {
		return ns[i][0].index < ns[j][0].index
	} else {
		return len(ns[i]) < len(ns[j])
	}
}

// Swap swaps the elements with indexes i and j.
func (ns NodeSorter) Swap(i, j int) {
	tmp := ns[i]
	ns[i] = ns[j]
	ns[j] = tmp
}

type VirtualNodeSorter []*Node

func (vns VirtualNodeSorter) Len() int {
	return len(vns)
}

func (vns VirtualNodeSorter) Less(i, j int) bool {
	return vns[i].index < vns[j].index
}

func (vns VirtualNodeSorter) Swap(i, j int) {
	tmp := vns[i]
	vns[i] = vns[j]
	vns[j] = tmp
}
