package domain

// Graph is a directed graph used to validate and order roadmap dependencies.
type Graph struct {
	nodes map[uint]struct{}
	edges map[uint]map[uint]struct{}
}

// NewGraph creates an empty graph.
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[uint]struct{}),
		edges: make(map[uint]map[uint]struct{}),
	}
}

// AddNode registers a node in the graph.
func (g *Graph) AddNode(id uint) {
	g.nodes[id] = struct{}{}
}

// HasNode reports whether the node is present.
func (g *Graph) HasNode(id uint) bool {
	_, ok := g.nodes[id]
	return ok
}

// NodeCount returns the number of nodes.
func (g *Graph) NodeCount() int {
	return len(g.nodes)
}

// AddEdge adds a dependency from -> to, rejecting edges that create a cycle.
func (g *Graph) AddEdge(from, to uint) error {
	g.AddNode(from)
	g.AddNode(to)
	if g.reaches(to, from) {
		return ErrGraphCycle
	}
	if g.edges[from] == nil {
		g.edges[from] = make(map[uint]struct{})
	}
	g.edges[from][to] = struct{}{}
	return nil
}

// EdgeCount returns the number of edges.
func (g *Graph) EdgeCount() int {
	count := 0
	for _, tos := range g.edges {
		count += len(tos)
	}
	return count
}

// Predecessors returns the nodes that directly depend on the given node
// being completed, i.e. edges where from_node = node.
func (g *Graph) Predecessors(node uint) []uint {
	var result []uint
	for from, tos := range g.edges {
		if _, ok := tos[node]; ok {
			result = append(result, from)
		}
	}
	return result
}

// HasCycle reports whether the graph contains a cycle.
func (g *Graph) HasCycle() bool {
	_, err := g.TopoOrder()
	return err != nil
}

// TopoOrder returns the nodes in dependency order using Kahn's algorithm.
func (g *Graph) TopoOrder() ([]uint, error) {
	indegree := make(map[uint]int, len(g.nodes))
	for node := range g.nodes {
		indegree[node] = 0
	}
	for _, tos := range g.edges {
		for to := range tos {
			indegree[to]++
		}
	}

	queue := make([]uint, 0, len(g.nodes))
	for node, degree := range indegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	order := make([]uint, 0, len(g.nodes))
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)
		for to := range g.edges[node] {
			indegree[to]--
			if indegree[to] == 0 {
				queue = append(queue, to)
			}
		}
	}

	if len(order) != len(g.nodes) {
		return nil, ErrGraphCycle
	}
	return order, nil
}

// reaches reports whether target is reachable from start following edges.
func (g *Graph) reaches(start, target uint) bool {
	seen := make(map[uint]struct{})
	stack := []uint{start}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if node == target {
			return true
		}
		if _, ok := seen[node]; ok {
			continue
		}
		seen[node] = struct{}{}
		for next := range g.edges[node] {
			stack = append(stack, next)
		}
	}
	return false
}
