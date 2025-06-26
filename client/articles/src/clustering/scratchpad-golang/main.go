package main

import(
    "github.com/bringyour/cluster/hdbscan"
    "fmt"
)

func printClusters(clusters hdbscan.Clustering, data [][]float64) {
    for i, cluster := range clusters.Clusters {
        fmt.Printf("Cluster %d:\n", i)
        for _, point := range cluster.Points {
            // fmt.Println(point) // Assuming cluster.Points contains the points in the cluster
            fmt.Println(point)
            // fmt.Println(data[point])
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
    }
    minimumClusterSize := 2
    minimumSpanningTree :=  false

    // create
    clustering, err := hdbscan.NewClustering(data, minimumClusterSize)
    if err != nil {
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
    //  fmt.Printf("%+v\n", cs)
    printClusters(*cs, data)
    // fmt.Printf(*cs.)
    // Returns 
    // hdbscan.clusters{(*hdbscan.cluster)(0xc000138000)}


}

// THIS SHIT IS A) NONDETERMINISTIC, B) NONCOMPREHENSIVE, C) WRONG
