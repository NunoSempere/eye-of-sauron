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
	"sort"
)

func (c *Clustering) selectOptimalClustering(score string) {
	if c.verbose {
		log.Println("selecting optimal clusters")
	}

	switch score {
	case VarianceScore:
		c.setVarianceDeltas()
	default:
		// setStabilityDelta(hierarchy)
	}

	var finalClusters clusters
	for _, cluster := range c.Clusters {
		if cluster.delta == 1 {
			finalClusters = append(finalClusters, cluster)
		}
	}

	c.Clusters = finalClusters

	if c.verbose {
		log.Println("finished selecting optimal clusters")
	}
}

func (c *Clustering) setVarianceDeltas() {
	// sort clusters by size
	sort.Sort(c.Clusters)

	for _, cluster := range c.Clusters {
		// calculate average childrens scores
		var avgScore float64
		for _, child := range cluster.children {
			avgScore += c.Clusters.getClusterByID(child).score
		}
		avgScore /= float64(len(cluster.children))

		if cluster.score <= avgScore && len(cluster.children) > 0 {
			cluster.delta = 0
		} else {
			cluster.delta = 1

			// set sub-clusters to 0
			subClusters := c.Clusters.subTree(cluster.id)
			for _, subCluster := range subClusters {
				subCluster.delta = 0
			}
		}
	}
}

func (c clusters) leaves() clusters {
	var leaves clusters
	for _, cluster := range c {
		if len(cluster.children) == 0 {
			leaves = append(leaves, cluster)
		}
	}
	return leaves
}

func (c clusters) allParents(clstr *cluster) clusters {
	var parents clusters

	if clstr.parent != nil {
		parentCluster := c.getClusterByID(*clstr.parent)
		allParents := c.allParents(parentCluster)
		parents = append(parents, allParents...)
	}

	return parents
}

func (c clusters) subTree(id int) clusters {
	var subTree clusters
	for _, cluster := range c {
		if cluster.parent != nil {
			if *cluster.parent == id {
				subTree = append(subTree, cluster)
				childTree := c.subTree(cluster.id)
				subTree = append(subTree, childTree...)
			}
		}
	}

	return subTree
}
