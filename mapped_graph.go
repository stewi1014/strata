package strata

func MakeMappedGraph[T comparable](cap int) *MappedGraph[T] {
	return &MappedGraph[T]{
		graph:    MakeGraph(0, cap),
		byObject: make(map[T]int, cap),
		objects:  make([]T, 0, cap),
	}
}

type MappedGraph[T comparable] struct {
	graph    Graph
	byObject map[T]int
	objects  []T
}

func (g *MappedGraph[T]) Len() int {
	return g.graph.Len()
}

func (g *MappedGraph[T]) Set(x, y T) {
	g.graph.Set(g.byObject[x], g.byObject[y])
}

func (g *MappedGraph[T]) Clear(x, y T) {
	g.graph.Clear(g.byObject[x], g.byObject[y])
}

func (g *MappedGraph[T]) Toggle(x, y T) {
	g.graph.Toggle(g.byObject[x], g.byObject[y])
}

func (g *MappedGraph[T]) Get(x, y T) bool {
	return g.graph.Get(g.byObject[x], g.byObject[y])
}

// Append adds nodes to the graph,
// returning the id of the first node.
// Subsequent nodes have ids incrementing from the returned id.
func (g *MappedGraph[T]) Append(nodes ...T) int {
	n := g.grow(len(nodes))

	copy(g.objects[n:], nodes)
	for i, node := range nodes {
		g.byObject[node] = i + n
	}

	return n
}

func (g *MappedGraph[T]) grow(n int) int {
	defer g.graph.grow(n)

	l := len(g.objects)

	if cap(g.objects) < l+n {
		newArray := make([]T, l+n, n+(cap(g.objects)*2))
		copy(newArray, g.objects)
		g.objects = newArray
		return l
	}

	g.objects = g.objects[:l+n]
	return l
}

func (g *MappedGraph[T]) NextChild(parent T, child *T) bool {
	childId := g.byObject[*child]

	result := g.graph.NextChild(g.byObject[parent], &childId)
	if result {
		*child = g.objects[childId]
	}

	return result
}

func (g *MappedGraph[T]) NextParant(child T, parent *T) bool {
	parantId := g.byObject[*parent]

	result := g.graph.NextParant(g.byObject[child], &parantId)
	if result {
		*parent = g.objects[parantId]
	}

	return result
}
