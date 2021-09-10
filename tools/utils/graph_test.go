package utils

import (
	"fmt"
	"testing"
)

func TestGraph_Path(t *testing.T) {
	g := NewGraph()

	g.AddEdge(NewNode(1, 0), NewNode(2, 0))
	g.AddEdge(NewNode(1, 0), NewNode(5, 0))
	g.AddEdge(NewNode(2, 0), NewNode(5, 0))
	g.AddEdge(NewNode(2, 0), NewNode(3, 0))
	g.AddEdge(NewNode(3, 0), NewNode(4, 0))
	g.AddEdge(NewNode(4, 0), NewNode(5, 0))
	paths, err := g.FindNodePath(NewNode(1, 0), NewNode(3, 0))
	if err != nil {
		t.Fatalf("error:%s", err.Error())
	}
	for _, path := range paths {
		for _, node := range path {
			fmt.Printf("%d->", node.id)
		}
		fmt.Println()
	}
}
