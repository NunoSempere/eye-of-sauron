package main

import(
    "github.com/bringyour/cluster/hdbscan"
    "fmt"
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

func extractClusters(clusters hdbscan.Clustering) [][]int{
    ns := [][]int{}

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
        ns = append(ns, ps)
        ns = append(ns, os)
    }
    return ns
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

func main() {
    data := [][]float64{
        []float64{1,2,3},
        []float64{1,2.2,3},
        []float64{3.2,2,1},
        []float64{3.1,2,1},
        []float64{5,5,5},
        []float64{5.2,5,5},
        []float64{7, -1, 0},
    }
    minimumClusterSize := 2
    minimumSpanningTree :=  false

    // create
    clustering, err := hdbscan.NewClustering(data, minimumClusterSize)
    if err != nil {
        fmt.Println(err)
        panic(err)
    }

    // options
    // clustering = clustering.Verbose().OutlierDetection()
    clustering = clustering.OutlierDetection()

    //run
    clustering.Run(hdbscan.EuclideanDistance, hdbscan.VarianceScore, minimumSpanningTree)

    // If using sampling, then can use the Assign() method afterwards on the total dataset.
    cs := (clustering)
    // fmt.Printf("Number of clusters: %d\n", len(cs))
    fmt.Printf("\n%+v\n\n", *cs)
    // fmt.Printf("%+v\n", *cs.Clusters)
    printClusters(*cs, data)

}
