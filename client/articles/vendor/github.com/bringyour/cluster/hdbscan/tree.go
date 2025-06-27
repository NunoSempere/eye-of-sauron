// Copyright 2020 Humility AI Incorporated, All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hdbscan

import (
	"math"
	"sync"
)

type edge struct {
	p1   int
	p2   int
	dist float64
}

type edges []edge

type tree struct {
	vertices []int
	edges    edges
	*sync.Mutex
}

func newTree() *tree {
	return &tree{
		vertices: make([]int, 0),
		edges:    make(edges, 0),
		Mutex:    &sync.Mutex{},
	}
}

func (t *tree) addVertice(vertice int) {
	t.vertices = append(t.vertices, vertice)
}

func (t *tree) addEdge(e edge) {
	t.edges = append(t.edges, e)
}

// the edges in the minimum spanning tree will be the minimum
// mutual-reachability distances between any two connected points.
// where no two points are connected by an edge unless they are closest
// to each other relative to other points.
func (c *Clustering) addRowToMinSpanningTree(row int, distances []float64) {
	c.mst.Lock()
	defer c.mst.Unlock()

	if len(c.mst.vertices) == 0 {
		c.mst.addVertice(row)
	}

	newEdge := c.mst.nearestVertice(row, distances)
	if newEdge.p1 == newEdge.p2 {
		return
	}

	// add new point and new edge to mst
	c.mst.addVertice(newEdge.p2)
	c.mst.addEdge(newEdge)
}

// nearestVertice will find the next smallest edge that is not already
// in the MST. Where "smallness" is found by finding the smallest
// mutual-reachability between two points that is not already an edge in the
// tree.
func (t *tree) nearestVertice(row int, distances []float64) edge {
	minDist := math.MaxFloat64
	p1Index := 0
	p2Index := 0

	if !containsInt(t.vertices, row) {
		// check distance between point_i and all points already in MST
		for _, j := range t.vertices {
			// if distance between point_i & vertice_j is the smallest-distance
			// left in graph then this will become a new edge in the MST.
			if minDist > distances[j] {
				minDist = distances[j]
				p1Index = j
				p2Index = row
			}
		}
	}

	// create smallest-distance edge between mst vertice_j and point_i
	return edge{p1: p1Index, p2: p2Index, dist: distances[p1Index]}
}

// Len ...
func (e edges) Len() int {
	return len(e)
}

// Swap ...
func (e edges) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

// Less ...
func (e edges) Less(i, j int) bool {
	return e[j].dist > e[i].dist
}
