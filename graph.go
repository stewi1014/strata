package strata

import "math"

// graphIndex returns the index for the given x-y vertex
// in a byte array.
//
// This is the core of how the graph operates.
// It maps a 2d number space to 1d. AKA a pair function.
//
// All pair functions I've found online have had significant flaws for this use case,
// so as far as I know it doesn't have a name.
//
// It follows this structure;
// .    0  1  2  3  4  .  .  .
//
// 0    0  1  4  9 16
// 1    2  3  6 11 18
// 2    5  7  8 13 20
// 3   10 12 14 15 22
// 4   17 19 21 23 24
// .                   .
// .                      .
// .                         .
//
// But instead of returning the index directly,
// it returns values for addressing a single bit in an array of bytes.
func graphIndex(x, y int) (i int, mask byte) {
	if x >= y {
		i = x*x + 2*y
	} else {
		i = y*y + 2*x + 1
	}

	return i / 8, 1 << (i & 7)
}

func MakeGraph(len int) Graph {
	return make(Graph, (len*len+7)/8)
}

// Graph contains information about verticies between nodes, where
// each node can have a vertex to any other node, with
// the verticies being directional.
//
// Graph is an alias for []byte,
// and can be manipulated like a slice.
// Keeping in mind that there is a square relationship between
// the number of nodes and slice dimensions.
// i.e. node count = √slice len
//
// Generally, it is highly optimised,
// but because it uses a single bit in memory for every vertex,
// graphs with more than 2^16 nodes start to use non-insignificant amounts of memory.
type Graph []byte

// Len returns the number of nodes in the graph.
//
// this is equivilent to the
func (g Graph) Len() int {
	// floating point square root is done in hardware on almost all systems,
	// so this horrible ~s***~ function is actually the most performant.
	return int(math.Sqrt(float64(len(g) * 8)))
}

func (g Graph) Cap() int {
	return int(math.Sqrt(float64(cap(g) * 8)))
}

func (g Graph) Get(x, y int) bool {
	i, mask := graphIndex(x, y)
	return g[i]&mask > 1
}

func (g Graph) Set(x, y int) {
	i, mask := graphIndex(x, y)
	g[i] |= mask
}

func (g Graph) Unset(x, y int) {
	i, mask := graphIndex(x, y)
	g[i] &^= mask
}
