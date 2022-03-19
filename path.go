package geopath

import (
	"container/heap"
	"errors"
	"math"

	"golang.org/x/exp/slices"
)

const earthRadius = 6371008.8

type path = [2]longLat

type longLat = [2]float64

var ErrNoPath = errors.New("no path")
var ErrGeoJSONParse = errors.New("geojson parse")

// FindShortestPath finds the shortest path between fromLongLat and toLongLat using
// LineString paths found in the provided io.Reader's GeoJSON contents.
//
// Input coordinates are expected to be in ESPG:4326 format.
//
// Precision determines the fuzziness of linking together the start and end
// points, and the start and ends of the line segments. If precision is unset
// no rounding will occur, and start/end points are required to perfectly match
// across the line segments and the start and end points.
//
// If a shortest path is found, the return path will include a set of points,
// starting with startLongLat and ending in endLongLat along with the total path
// distance.
//
// If no path is found, an error of type ErrNoPath is returned.
//
func FindShortestPath(
	paths []path,
	startLongLat [2]float64,
	endLongLat [2]float64,
	precision float64,
) (path [][2]float64, dist float64, err error) {
	// Maybe round points
	round := func(x [2]float64) [2]float64 {
		x[0] = math.Round(x[0]/precision) * precision
		x[1] = math.Round(x[1]/precision) * precision
		return x
	}
	if precision != 0 {
		startLongLat = round(startLongLat)
		endLongLat = round(endLongLat)
		for i := range paths {
			paths[i][0] = round(paths[i][0])
			paths[i][1] = round(paths[i][1])
		}
	}

	// Remove duplicates (if any)
	slices.SortFunc(paths, func(a, b [2][2]float64) bool {
		if a[0][0] == b[1][0] {
			return a[0][1] < b[1][1]
		}
		return a[0][0] < b[1][0]
	})
	slices.Compact(paths)

	// Give each point a sequential index
	pointIdx := make(map[[2]float64]int, len(paths)*2)
	points := make([][2]float64, 0, len(paths)*2)
	addPoint := func(p [2]float64) {
		if _, exists := pointIdx[p]; exists {
			return
		}
		pointIdx[p] = len(points)
		points = append(points, p)
	}
	for _, path := range paths {
		addPoint(path[0])
		addPoint(path[1])
	}
	npoint := len(points)

	// Find closest point to start/end
	startIdx, startDist := -1, math.MaxFloat64
	endIdx, endDist := -1, math.MaxFloat64
	for i, p := range points {
		if d := calcDist(p, startLongLat); d < startDist {
			startDist = d
			startIdx = i
		}
		if d := calcDist(p, endLongLat); d < endDist {
			endDist = d
			endIdx = i
		}
	}

	// Create adj list
	adj := make([][]int, npoint)
	adjDist := make([][]float64, npoint)
	for _, path := range paths {
		p1, p2 := pointIdx[path[0]], pointIdx[path[1]]
		d := calcDist(path[0], path[1])
		adjDist[p1] = append(adjDist[p1], d)
		adjDist[p2] = append(adjDist[p2], d)
		adj[p1] = append(adj[p1], p2)
		adj[p2] = append(adj[p2], p1)
	}

	// Initialize distance vector
	dists := make([]float64, npoint)
	for i := range dists {
		dists[i] = math.MaxFloat64
	}

	// Initialize parent path
	parent := make([]int, npoint)
	for i := range parent {
		parent[i] = i
	}

	// Perform Dijkstra's
	start := edge{
		from: startIdx,
		to:   startIdx,
		dist: 0,
	}
	h := visitHeap{start}
	for len(h) > 0 {
		x := heap.Pop(&h).(edge)
		if x.dist > dists[x.to] {
			continue
		}
		parent[x.to] = x.from // link points together
		if x.to == endIdx {
			break
		}
		for i, nei := range adj[x.to] {
			d := x.dist + adjDist[x.to][i]
			if d >= dists[nei] {
				continue
			}
			dists[nei] = d
			heap.Push(&h, edge{
				from: x.to,
				to:   nei,
				dist: d,
			})
		}
	}

	// Return error if there was no path
	if dists[endIdx] == math.MaxFloat64 {
		return nil, math.MaxFloat64, ErrNoPath
	}

	// Gather path (backwards) and reverse the list
	shortestPath := make([][2]float64, 1, npoint)
	shortestPath[0] = endLongLat
	for i := endIdx; i != startIdx; i = parent[i] {
		shortestPath = append(shortestPath, points[i])
	}
	shortestPath = append(shortestPath, points[startIdx], startLongLat)
	for l, r := 0, len(shortestPath)-1; l < r; l, r = l+1, r-1 {
		shortestPath[l], shortestPath[r] = shortestPath[r], shortestPath[l]
	}

	totalDist := startDist + dists[endIdx] + endDist
	return shortestPath, totalDist, nil
}

type edge struct {
	from int
	to   int
	dist float64
}

type visitHeap []edge

func (h visitHeap) Len() int { return len(h) }
func (h visitHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
func (h visitHeap) Less(i, j int) bool {
	return h[i].dist < h[j].dist
}
func (h *visitHeap) Push(x interface{}) {
	*h = append(*h, x.(edge))
}
func (h *visitHeap) Pop() interface{} {
	n := len(*h)
	it := (*h)[n-1]
	*h = (*h)[:n-1]
	return it
}

func calcDist(p1, p2 [2]float64) float64 {
	radian := func(x float64) float64 {
		return x * math.Pi / 180
	}
	dLong := radian(p2[0] - p1[0])
	dLat := radian(p2[1] - p1[1])
	lat1 := radian(p1[1])
	lat2 := radian(p2[1])
	a := math.Pow(math.Sin(dLat/2), 2)
	b := math.Pow(math.Sin(dLong/2), 2) * math.Cos(lat1) * math.Cos(lat2)
	c := a + b
	d := 2 * math.Atan2(math.Sqrt(c), math.Sqrt(1-c))
	return d * earthRadius
}
