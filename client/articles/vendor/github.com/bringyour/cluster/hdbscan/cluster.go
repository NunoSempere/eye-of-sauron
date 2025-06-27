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
	"log"
	"math"
	"runtime"
	"sync"

	"gonum.org/v1/gonum/stat/distuv"
)

var (
	// VarianceScore will select an optimal clustering
	// that minimizes the generalized variance across each cluster.
	VarianceScore = "variance_score"
	// StabilityScore will select an optimal clustering that
	// maximized the stability across all clusters.
	StabilityScore = "stability_score"
)

type cluster struct {
	id int

	parent   *int
	children []int

	score    float64
	delta    int
	size     float64
	variance float64
	lMin     float64

	distanceDistribution *distuv.Normal
	largestDistance      float64

	//public
	Centroid []float64
	Points   []int
	Outliers Outliers
}

type clusters []*cluster

// Outlier struct is used to provide information
// about an outlier data point.
type Outlier struct {
	Index              int
	NormalizedDistance float64
}

// Outliers is an array of outlier points
// for a given cluster.
type Outliers []Outlier

// Clustering struct which holds
// all final results.
type Clustering struct {
	data [][]float64

	// settings
	mcs          int
	minTree      bool
	verbose      bool
	randomSample bool
	subSample    bool
	voronoi      bool
	nn           bool
	od           bool
	oc           bool
	sampleBound  int
	distanceFunc DistanceFunc

	// minimum spanning tree
	mst *tree

	// optimal-clustering
	score    string
	Clusters clusters

	semaphore chan bool
	wg        *sync.WaitGroup
}

// NewClustering creates (a pointer to) a new clustering struct.
// This function does not automatically start the clustering
// process. The `Run` method needs to be called to do that.
// Make sure to apply all options *before* calling `Run`.
func NewClustering(data [][]float64, minimumClusterSize int) (*Clustering, error) {
	if minimumClusterSize < 1 {
		return &Clustering{}, ErrMCS
	}

	if len(data) < minimumClusterSize {
		return &Clustering{}, ErrDataLen
	}

	dataLength := len(data[0])

	for _, row := range data {
		if len(row) != dataLength {
			return &Clustering{}, ErrRowLength
		}
	}

	return &Clustering{
		data:      data,
		mcs:       minimumClusterSize,
		mst:       newTree(),
		semaphore: make(chan bool, runtime.NumCPU()),
		wg:        &sync.WaitGroup{},
	}, nil
}

// Run will run the clustering.
func (c *Clustering) Run(distanceFunc DistanceFunc, score string, mst bool) error {
	c.distanceFunc = distanceFunc
	c.minTree = mst
	if c.verbose && !c.minTree {
		log.Println("not using minimum spanning tree")
	}

	c.sample()
	c.buildClusters(c.buildDendrogram(c.mutualReachabilityGraph()))
	c.scoreClusters(score)
	c.selectOptimalClustering(score)
	c.clusterCentroids()
	c.outliersAndVoronoi()
	c.outlierClustering()

	return nil
}

// the clusters hierarchy will not contain clusters that are smaller than the minimum cluster size
// every leaf-cluster is unique subset of points.
func (c *Clustering) buildClusters(links []*link) {
	if c.verbose {
		log.Println("building clusters")
	}

	var clusters clusters

	for i, link := range links {
		if len(link.points) >= c.mcs {
			var children []int
			for _, childLink := range link.children {
				children = append(children, childLink.id)
			}

			newCluster := &cluster{
				id:       i,
				Points:   link.points,
				Outliers: make(Outliers, 0),
				children: children,
			}

			if link.parent == nil {
				newCluster.parent = nil
			} else {
				id := link.parent.id
				newCluster.parent = &id
			}

			clusters = append(clusters, newCluster)
		}
	}

	for _, cluster := range clusters {
		var newChildren []int
		for _, child := range cluster.children {
			if clusters.getClusterByID(child) != nil {
				newChildren = append(newChildren, child)
			}
		}
		cluster.children = newChildren
	}

	c.Clusters = clusters

	if c.verbose {
		log.Println("finished building clusters")
	}
}

func (c *Clustering) clusterCentroids() {
	if c.verbose {
		log.Println("calculating cluster centroids")
	}

	for i, cluster := range c.Clusters {
		avg := make([]float64, len(c.data[0]), len(c.data[0]))
		for _, index := range cluster.Points {
			vec := c.data[index]
			if len(vec) == len(avg) {
				for j, v := range vec {
					avg[j] += v
				}
			}
		}

		for k, v := range avg {
			v /= float64(len(cluster.Points))
			avg[k] = v
		}

		cluster.Centroid = avg
		c.Clusters[i] = cluster
	}

	if c.verbose {
		log.Println("finished calculating cluster centroids")
	}
}

func (c *Clustering) outliersAndVoronoi() {
	if !c.od && !c.voronoi {
		return
	}

	if len(c.Clusters) == 0 {
		return
	}

	if c.verbose {
		if c.od {
			log.Println("finding outliers")
		}

		if c.voronoi {
			log.Println("starting voronoi clustering")
		}
	}

	for i, v := range c.data {
		var exists bool
		for _, cluster := range c.Clusters {
			for _, point := range cluster.Points {
				if point == i {
					exists = true
					break
				}
			}

			if exists {
				break
			}
		}

		if !exists {
			// calculate nearest cluster
			minDistance := math.MaxFloat64
			var nearestClusterIndex int
			for i, cluster := range c.Clusters {
				if c.nn {
					for _, p := range cluster.Points {
						distance := c.distanceFunc(c.data[p], v)
						if distance < minDistance {
							minDistance = distance
							nearestClusterIndex = i
						}
					}
				} else {
					distance := c.distanceFunc(cluster.Centroid, v)
					if distance < minDistance {
						minDistance = distance
						nearestClusterIndex = i
					}
				}
			}

			// voronoi cluster
			if c.voronoi {
				c.Clusters[nearestClusterIndex].Points = append(c.Clusters[nearestClusterIndex].Points, i)
			}

			// outlier detection
			if c.od {
				c.Clusters[nearestClusterIndex].Outliers = append(c.Clusters[nearestClusterIndex].Outliers, Outlier{
					Index:              i,
					NormalizedDistance: minDistance,
				})
			}
		}
	}

	// normalize outlier distances
	if c.od {
		c.distanceDistributions()

		for _, cluster := range c.Clusters {
			for j, outlier := range cluster.Outliers {
				outlier.NormalizedDistance = isNum(cluster.distanceDistribution.CDF(outlier.NormalizedDistance))
				cluster.Outliers[j] = outlier
			}
		}
	}

	if c.verbose {
		if c.od {
			log.Println("finished finding outliers")
		}

		if c.voronoi {
			log.Println("finished voronoi clustering")
		}
	}
}

func (c *Clustering) distanceDistributions() {
	for i, cluster := range c.Clusters {
		// distance distribution
		ld := float64(math.MinInt64)
		var distances []float64
		for j1, p1 := range cluster.Points {
			if c.nn {
				minDistance := math.MaxFloat64
				for j2, p2 := range cluster.Points {
					if j1 != j2 {
						distance := c.distanceFunc(c.data[p1], c.data[p2])
						if distance < minDistance {
							minDistance = distance
						}
					}
				}
				distances = append(distances, minDistance)

				// largest NN-distance
				if minDistance > ld {
					ld = minDistance
				}
			} else {
				distance := c.distanceFunc(c.data[p1], cluster.Centroid)
				distances = append(distances, distance)
				if distance > ld {
					ld = distance
				}
			}
		}
		dd := &distuv.Normal{}
		dd.Fit(distances, nil)
		cluster.distanceDistribution = dd
		cluster.largestDistance = ld

		c.Clusters[i] = cluster
	}
}

func (c clusters) getClusterByID(id int) *cluster {
	for _, cluster := range c {
		if cluster.id == id {
			return cluster
		}
	}

	return nil
}

// Len ...
func (c clusters) Len() int {
	return len(c)
}

// Swap ...
func (c clusters) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Less ...
func (c clusters) Less(i, j int) bool {
	return len(c[i].Points) < len(c[j].Points)
}

func (c clusters) maxID() int {
	var maxID int
	for _, clust := range c {
		if clust.id > maxID {
			maxID = clust.id
		}
	}

	return maxID
}

// MinProb ...
func (o Outliers) MinProb() Outlier {
	minProb := float64(1)
	var ol Outlier

	for _, v := range o {
		if v.NormalizedDistance <= minProb {
			minProb = v.NormalizedDistance
			ol = v
		}
	}

	return ol
}

func (c *Clustering) outlierClustering() {
	if !c.oc {
		return
	}

	maxID := c.Clusters.maxID()
	var newClusters clusters
	for i, clust := range c.Clusters {
		if len(clust.Outliers) >= c.mcs {
			newCluster := &cluster{
				id:     i + maxID + 1,
				Points: make([]int, 0),
			}

			for _, o := range clust.Outliers {
				newCluster.Points = append(newCluster.Points, o.Index)
			}

			newClusters = append(newClusters, newCluster)
		}
	}

	c.Clusters = append(c.Clusters, newClusters...)
}
