package strata

// vertexIndex returns the index for the given x-y vertex.
//
// This is the core of how the graph operates. It maps a 2d number space to 1d.
//
// E.g.
//
//	+   1  2  3  4  5  .  .  .
// 
//	1   0  1  4  9 16
//	2   3  2  5 10 17
//	3   8  7  6 11 18
//	4  15 14 13 12 19
//	5  24 23 22 21 20
//	.                 .
//	.                   .
//	.                     .
func GraphIndex(x, y int) int {
	if x > y {
		return x*x + y
	} else {
		return y*(y+2) - x
	}
}

type Graph []byte

func (g *Graph) Get(x, y int) bool {
	i := x * 
}

func (g *Graph) Set(x, y int, v bool) {

}
