package strata

const (
	wordShift = 6
	wordSize  = 1 << wordShift
	wordMask  = wordSize - 1
)

// result * wordSize >= dividend
func roundUpWordDivision(dividend int) (result int) {
	if dividend&wordMask > 0 {
		return (dividend >> wordShift) + 1
	}
	return dividend >> wordShift
}

// graphVertexCount returns the number of possible verticies between n nodes.
// That is, n squared after accounting for the fact that indicies start at zero.
func graphVertexCount(n int) int {
	return (n + 1) * (n + 1)
}

// Graphs do not need to be initialised.
//
// MakeGraph circumvents repeated allocation of larger arrays if the graph size is known ahead of time.
func MakeGraph(len, cap int) Graph {
	return Graph{
		len: len,
		array: make(
			[]uint64,
			roundUpWordDivision(len*len),
			roundUpWordDivision(cap*cap),
		),
	}
}

// Graph implements a Graph datastructure.
// It allows self-referencing nodes and
type Graph struct {
	len   int
	array []uint64
}

// vertexIndex returns the index for the given x-y vertex.
func (g Graph) vertexIndex(x, y int) int {
	if x > g.len || y > g.len {
		panic("index out of bounds")
	}

	if x > y {
		return x*x + y
	} else {
		return y*(y+2) - x
	}
}

// Len returns the number of nodes in the graph.
func (g Graph) Len() int {
	return g.len
}

// Set creates a vertex from node x to node y.
func (g Graph) Set(x, y int) {
	i := g.vertexIndex(x, y)

	g.array[i>>wordShift] |= 1 << (i & wordMask)
}

// Clear removes a vertex from node x to node y.
func (g Graph) Clear(x, y int) {
	i := g.vertexIndex(x, y)

	g.array[i>>wordShift] &^= 1 << (i & wordMask)
}

// Toggle toggles the vertex from node x to node y.
func (g Graph) Toggle(x, y int) {
	i := g.vertexIndex(x, y)

	g.array[i>>wordShift] ^= 1 << (i & wordMask)
}

// Get returns true if a vertex from node x to node y exists.
func (g Graph) Get(x, y int) bool {
	i := g.vertexIndex(x, y)

	return g.array[i>>wordShift]&1<<(i&wordMask) != 0
}

// Append adds n new nodes to the graph,
// returning the id of the first added node.
// Subsequent nodes have incrementing ids from the first.
func (g *Graph) Append(n int) int {
	defer g.grow(n)
	return g.len
}

// grow increases the size of the graph by n nodes.
func (g *Graph) grow(n int) {
	newLen := roundUpWordDivision(graphVertexCount(g.len + n))

	if cap(g.array) >= newLen {
		g.array = g.array[:newLen]
		return
	}

	newArray := make([]uint64, newLen, newLen+(cap(g.array)*2))
	copy(newArray, g.array)
	g.array = newArray
}

// NextChild assigns *child to the id of the next child of parant.
//
// E.g. to iterate over all children of parentNode;
//
//	 for childNode := -1; NextChild(parentNode, &childNode); {
//		    // Do something with child node
//	 }
func (g Graph) NextChild(parent int, child *int) bool {
	for ; *child < g.len; *child++ {
		if g.Get(parent, *child) {
			return true
		}
	}

	return false
}

// NextParent assigns *parent to the id of the next parent of child.
//
// E.g. to iterate over all parents of childNode;
//
//	 for parentNode := -1; NextParant(childNode, &parantNode); {
//		    // Do something with parant node
//	 }
func (g Graph) NextParant(child int, parent *int) bool {
	for ; *parent < g.len; *parent++ {
		if g.Get(*parent, child) {
			return true
		}
	}

	return false
}
