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

// OutlierDetection will track all unassigned
// points as outliers of their nearest cluster.
// It provides a `NormalizedDistance` value for
// each outlier which can be interpreted as the
// probability of the point being an outlier
// (relative to all other outliers).
func (c *Clustering) OutlierDetection() *Clustering {
	c.od = true
	return c
}

// Verbose will set verbosity to true for clustering process
// and the internals of a clustering run will be logged to stdout.
func (c *Clustering) Verbose() *Clustering {
	c.verbose = true
	return c
}

// Voronoi will set voronoi-clustering to true, and
// after density clustering is performed,
// all points not assigned to a cluster will be placed
// into their nearest cluster (by centroid distance).
func (c *Clustering) Voronoi() *Clustering {
	c.voronoi = true
	return c
}

// NearestNeighbor specifies if nearest-neighbor
// distances should be used for outlier detection
// and for voronoi clustering instead of centroid-based
// distances.
// NearestNeighbor will find the closest assigned data
// point to an unassigned data point and consider the
// unassigned data point to be of that same cluster (as an outlier and/or a point).
func (c *Clustering) NearestNeighbor() *Clustering {
	c.nn = true
	return c
}

// OutlierClustering is an option to group the outliers of a cluster
// into a new cluster if there are at least a minimum-cluster-size
// number of them.
// This option will automatically perform outlier detection on the clustering
// as well.
func (c *Clustering) OutlierClustering() *Clustering {
	c.od = true
	c.oc = true
	return c
}

// Subsample will take the first 'n' data points and perform clustering on
// those. 'n' is a provided argument and should be between 0 and the total data size.
// Voronoi clustering will be performed after the clusters have been found for all points
// that are not in the subsample.
func (c *Clustering) Subsample(n int) *Clustering {
	if n < 0 || n > len(c.data) {
		n = len(c.data)
	}
	c.sampleBound = n
	c.subSample = true

	return c
}

// randomSampling ...  (DO NOT USE!)
// will process a random sample of the total provided data.
// The amount of data in the sample is based on the percentage (0-1) argument provided.
// Randomly sampled data will also automatically perform a Voronoi clustering for
// all data points by using the centroids of the clusters generated from the sample.
// In other words, the final result may not be as of high quality as on a different sample
// or the entire dataset.
func (c *Clustering) randomSampling(percentage float64) *Clustering {
	// if percentage is not valid: process total data
	if percentage < 0 || percentage > 1 {
		percentage = 1
	}

	bound := int(percentage * 100)
	c.sampleBound = bound

	c.randomSample = true

	return c
}
