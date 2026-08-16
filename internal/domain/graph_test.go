package domain

import (
	"errors"
	"testing"
)

func TestGraphTopoOrder(t *testing.T) {
	t.Parallel()

	g := NewGraph()
	if err := g.AddEdge(1, 2); err != nil {
		t.Fatalf("AddEdge(1,2): %v", err)
	}
	if err := g.AddEdge(2, 3); err != nil {
		t.Fatalf("AddEdge(2,3): %v", err)
	}
	if err := g.AddEdge(1, 3); err != nil {
		t.Fatalf("AddEdge(1,3): %v", err)
	}

	order, err := g.TopoOrder()
	if err != nil {
		t.Fatalf("TopoOrder: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(order))
	}

	position := make(map[uint]int, len(order))
	for i, n := range order {
		position[n] = i
	}
	if position[1] > position[2] || position[2] > position[3] {
		t.Fatalf("order violates dependencies: %v", order)
	}
}

func TestGraphAddEdgeRejectsCycle(t *testing.T) {
	t.Parallel()

	g := NewGraph()
	if err := g.AddEdge(1, 2); err != nil {
		t.Fatalf("AddEdge(1,2): %v", err)
	}
	if err := g.AddEdge(2, 3); err != nil {
		t.Fatalf("AddEdge(2,3): %v", err)
	}
	if err := g.AddEdge(3, 1); !errors.Is(err, ErrGraphCycle) {
		t.Fatalf("expected ErrGraphCycle, got %v", err)
	}
	if g.EdgeCount() != 2 {
		t.Fatalf("cycle edge must not be inserted, got %d edges", g.EdgeCount())
	}
}

func TestGraphSelfDependency(t *testing.T) {
	t.Parallel()

	g := NewGraph()
	if err := g.AddEdge(1, 1); !errors.Is(err, ErrGraphCycle) {
		t.Fatalf("expected ErrGraphCycle for self edge, got %v", err)
	}
}

func TestGraphHasCycle(t *testing.T) {
	t.Parallel()

	g := NewGraph()
	g.AddNode(1)
	g.AddNode(2)
	g.AddNode(3)
	g.edges[1] = map[uint]struct{}{2: {}}
	g.edges[2] = map[uint]struct{}{3: {}}
	g.edges[3] = map[uint]struct{}{1: {}}

	if !g.HasCycle() {
		t.Fatal("expected HasCycle to be true")
	}
	if _, err := g.TopoOrder(); !errors.Is(err, ErrGraphCycle) {
		t.Fatalf("expected ErrGraphCycle from TopoOrder, got %v", err)
	}
}

func TestGraphPredecessors(t *testing.T) {
	t.Parallel()

	g := NewGraph()
	if err := g.AddEdge(1, 3); err != nil {
		t.Fatalf("AddEdge(1,3): %v", err)
	}
	if err := g.AddEdge(2, 3); err != nil {
		t.Fatalf("AddEdge(2,3): %v", err)
	}

	deps := g.Predecessors(3)
	if len(deps) != 2 {
		t.Fatalf("expected 2 predecessors, got %d", len(deps))
	}
	if g.Predecessors(1) != nil {
		t.Fatalf("expected no predecessors for node 1")
	}
}

func TestGraphDiamond(t *testing.T) {
	t.Parallel()

	g := NewGraph()
	// 1 -> 2, 1 -> 3, 2 -> 4, 3 -> 4
	if err := g.AddEdge(1, 2); err != nil {
		t.Fatalf("AddEdge(1,2): %v", err)
	}
	if err := g.AddEdge(1, 3); err != nil {
		t.Fatalf("AddEdge(1,3): %v", err)
	}
	if err := g.AddEdge(2, 4); err != nil {
		t.Fatalf("AddEdge(2,4): %v", err)
	}
	if err := g.AddEdge(3, 4); err != nil {
		t.Fatalf("AddEdge(3,4): %v", err)
	}

	order, err := g.TopoOrder()
	if err != nil {
		t.Fatalf("TopoOrder: %v", err)
	}
	if len(order) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(order))
	}
	if g.HasCycle() {
		t.Fatal("diamond must not contain a cycle")
	}
}
