package utils

import (
	"errors"
	"fmt"
	"sync"
)

type Node struct {
	id    interface{}
	value interface{}
}

func NewNode(id, value interface{}) *Node {
	return &Node{id: id, value: value}
}

type Graph struct {
	nodes  []*Node // 节点集
	nodeId map[interface{}]uint32
	edges  map[Node][]*Node // 邻接表表示的无向图
	lock   sync.RWMutex     // 保证线程安全}
}

func NewGraph() *Graph {
	return &Graph{
		nodes:  make([]*Node, 0),
		nodeId: make(map[interface{}]uint32, 0),
		edges:  make(map[Node][]*Node, 0),
		lock:   sync.RWMutex{},
	}
}

// 增加节点
func (g *Graph) AddNode(n *Node) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.addNode(n)
}

func (g *Graph) addNode(n *Node) {
	index, ok := g.nodeId[n.id]
	if ok {
		g.nodes[index] = n
	} else {
		g.nodes = append(g.nodes, n)
		g.nodeId[n.id] = uint32(len(g.nodes) - 1)
	}
}

// 增加边
func (g *Graph) AddEdge(u, v *Node) {
	g.lock.Lock()
	defer g.lock.Unlock()
	// 首次建立图
	if g.edges == nil {
		g.edges = make(map[Node][]*Node)
	}
	g.addNode(u)
	g.addNode(v)
	g.edges[*u] = append(g.edges[*u], v) // 建立 u->v 的边
	g.edges[*v] = append(g.edges[*v], u) // 由于是无向图，同时存在 v->u 的边
}

// 输出图
func (g *Graph) String() {
	g.lock.RLock()
	defer g.lock.RUnlock()
	str := ""
	for _, iNode := range g.nodes {
		str += iNode.String() + " -> "
		nexts := g.edges[*iNode]
		for _, next := range nexts {
			str += next.String() + " -> "
		}
		str += "\n"
	}
	fmt.Println(str)
}

// 实现 BFS 遍历
func (g *Graph) FindNodePath(node1, node2 *Node) ([][]*Node, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	key := NewPathKey()
	key.Add(node1.id)
	return g.path(node1, node2, []*Node{}, key)
}

// 实现 BFS 遍历
func (g *Graph) path(node1, node2 *Node, path []*Node, key *PathKey) ([][]*Node, error) {
	paths := make([][]*Node, 0)
	path = append(path, node1)
	index1, ok := g.nodeId[node1.id]
	if !ok {
		return paths, errors.New("no exist")
	}
	node := g.nodes[index1]
	nexts := g.edges[*node]
	for _, next := range nexts {
		if next.id == node2.id {
			newPath := copyPath(path)
			newPath = append(newPath, next)
			paths = append(paths, newPath)
		} else {
			if !key.Exist(next.id) {
				newKey := key.Copy()
				newKey.Add(next.id)
				newPaths, err := g.path(next, node2, path, newKey)
				if err == nil {
					for _, newPath := range newPaths {
						paths = append(paths, newPath)
					}
				}
			}
		}
	}

	return paths, nil
}

func pathKey(nodes []*Node) string {
	key := ""
	for _, node := range nodes {
		key = fmt.Sprintf("%s%v", key, node.id)
	}
	return key
}

func copyPath(path []*Node) []*Node {
	newPath := make([]*Node, len(path))
	for i, node := range path {
		newPath[i] = node
	}
	return newPath
}

// 输出节点
func (n *Node) String() string {
	return fmt.Sprintf("%v", n.id)
}

type NodeQueue struct {
	nodes []Node
	lock  sync.RWMutex
}

// 实现 BFS 遍历
func (g *Graph) BFS(f func(node *Node)) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	// 初始化队列
	q := NewNodeQueue()
	// 取图的第一个节点入队列
	head := g.nodes[0]
	q.Enqueue(*head)
	// 标识节点是否已经被访问过
	visited := make(map[*Node]bool)
	visited[head] = true
	// 遍历所有节点直到队列为空
	for {
		if q.IsEmpty() {
			break
		}
		node := q.Dequeue()
		visited[node] = true
		nexts := g.edges[*node]
		// 将所有未访问过的邻接节点入队列
		for _, next := range nexts {
			// 如果节点已被访问过
			if visited[next] {
				continue
			}
			q.Enqueue(*next)
			visited[next] = true
		}
		// 对每个正在遍历的节点执行回调
		if f != nil {
			f(node)
		}
	}
}

// 生成节点队列
func NewNodeQueue() *NodeQueue {
	q := NodeQueue{}
	q.lock.Lock()
	defer q.lock.Unlock()
	q.nodes = []Node{}
	return &q
}

// 入队列
func (q *NodeQueue) Enqueue(n Node) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.nodes = append(q.nodes, n)
}

// 出队列
func (q *NodeQueue) Dequeue() *Node {
	q.lock.Lock()
	defer q.lock.Unlock()
	node := q.nodes[0]
	q.nodes = q.nodes[1:]
	return &node
}

// 判空
func (q *NodeQueue) IsEmpty() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.nodes) == 0
}

type PathKey []interface{}

func NewPathKey() *PathKey {
	return &PathKey{}
}

func (p *PathKey) Add(id interface{}) {
	*p = append(*p, id)
}

func (p *PathKey) Copy() *PathKey {
	x := NewPathKey()
	for _, i := range *p {
		x.Add(i)
	}
	return x
}

func (p *PathKey) Exist(id interface{}) bool {
	for _, i := range *p {
		if i == id {
			return true
		}
	}
	return false
}
