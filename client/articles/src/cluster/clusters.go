package cluster

import (
	"fmt"
	"github.com/bringyour/cluster/hdbscan"
)

func unique[A comparable](input []A) []A {
	seen := make(map[A]bool)
	var result []A
	for _, v := range input {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

type Cluster struct {
	points []int
	outliers []int
}

func extractClusters(clusters hdbscan.Clustering) []Cluster {
	cs := []Cluster{}

	for _, cluster := range clusters.Clusters {
		// fmt.Printf("Cluster %d: %+v\n", i, cluster)
		ps := []int{}
		os := []int{}

		for _, p := range cluster.Points {
			ps = append(ps, p)
		}
		for _, o := range cluster.Outliers {
			// ps = append(ps, o.Index)
			os = append(os, o.Index)
		}
		ps = unique(ps)

		c := Cluster{points: ps, outliers: os}
		cs = append(cs, c)
	}
	return cs
}

func printClusters(clusters hdbscan.Clustering, data [][]float64) {
	for _, cluster := range clusters.Clusters {
		// fmt.Printf("Cluster %d: %+v\n", i, cluster)
		ps := []int{}

		for _, p := range cluster.Points {
			ps = append(ps, p)
		}
		for _, o := range cluster.Outliers {
			ps = append(ps, o.Index)
		}
		ps = unique(ps)
		for _, p := range ps {
			fmt.Printf("%v, ", data[p])
		}
		fmt.Println()
	}
}

func GetClusters(data [][]float64) []Cluster {
	/*data := [][]float64{
	    []float64{1,2,3},
	    []float64{1,2.2,3},
	    []float64{3.2,2,1},
	    []float64{3.1,2,1},
	    []float64{5,5,5},
	    []float64{5.2,5,5},
	    []float64{7, -1, 0},
	}*/
	minimumClusterSize := 2
	minimumSpanningTree := false

	// create
	clustering, err := hdbscan.NewClustering(data, minimumClusterSize)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// clustering = clustering.Verbose().OutlierDetection()
	clustering = clustering.OutlierDetection()
	clustering.Run(hdbscan.EuclideanDistance, hdbscan.VarianceScore, minimumSpanningTree)

	// fmt.Printf("\n%+v\n\n", *cs)
	// printClusters(*cs, data)
	result := extractClusters(*clustering)
	return result

}
