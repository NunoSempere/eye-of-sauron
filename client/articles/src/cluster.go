package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bringyour/cluster/hdbscan"
	openai "github.com/sashabaranov/go-openai"
)

func getEmbeddings(texts []string, token string) ([][]float64, error) {
	client := openai.NewClient(token)

	// Create an EmbeddingRequest for the user query
	queryReq := openai.EmbeddingRequest{
		Input: texts,
		Model: openai.SmallEmbedding3,
	}

	// Create an embedding for the user query
	queryResponse, err := client.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		return [][]float64{}, err
	}

	es := [][]float64{}
	for _, e := range queryResponse.Data {
		f32s := e.Embedding
		f64s := make([]float64, len(f32s))
		for i, v := range f32s {
			f64s[i] = float64(v)
		}
		es = append(es, f64s)
	}

	return es, nil
}

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

func extractClusters(clusters hdbscan.Clustering) []Cluster {
	cs := []Cluster{}

	for i, cluster := range clusters.Clusters {
		ps := []int{}
		os := []int{}

		for _, p := range cluster.Points {
			ps = append(ps, p)
		}
		for _, o := range cluster.Outliers {
			os = append(os, o.Index)
		}
		ps = unique(ps)

		c := Cluster{ID: i, Points: ps, Outliers: os}
		cs = append(cs, c)
	}
	return cs
}

func getClusters(data [][]float64) []Cluster {
	minimumClusterSize := 2
	minimumSpanningTree := false

	// create
	clustering, err := hdbscan.NewClustering(data, minimumClusterSize)
	if err != nil {
		fmt.Printf("Error creating clustering: %v\n", err)
		return []Cluster{}
	}

	clustering = clustering.OutlierDetection()
	clustering.Run(hdbscan.EuclideanDistance, hdbscan.VarianceScore, minimumSpanningTree)

	result := extractClusters(*clustering)
	return result
}

func cleanTitleForEmbedding(title string) string {
	clean := strings.ReplaceAll(title, "<b>", "")
	clean = strings.ReplaceAll(clean, "</b>", "")
	clean = strings.ReplaceAll(clean, "'", "'")
	return clean
}

func (a *App) clusterSources() error {
	if len(a.sources) < 2 {
		return nil // Not enough sources to cluster
	}

	openaiKey := os.Getenv("OPENAI_KEY")
	if openaiKey == "" {
		fmt.Println("OPENAI_KEY not found, skipping clustering")
		return nil
	}

	// Extract titles for embedding
	titles := make([]string, len(a.sources))
	for i, source := range a.sources {
		titles[i] = cleanTitleForEmbedding(source.Title)
	}

	// Get embeddings
	a.drawLines([]string{"Getting sources...", "Clustering sources...", "Getting embeddings..."})
	// fmt.Printf("\nGetting embeddings...")
	embeddings, err := getEmbeddings(titles, openaiKey)
	if err != nil {
		fmt.Printf("Error getting embeddings: %v\n", err)
		return err
	}

	// Get clusters
	// fmt.Printf("\nCalculating clusters...")
	a.drawLines([]string{"Getting sources...", "Clustering sources...", "Getting embeddings...", "Calculating clusters... [this may take a while]"})
	clusters := getClusters(embeddings)

	// Assign cluster information to sources
	for i := range a.sources {
		a.sources[i].ClusterID = -1 // Default: no cluster
		a.sources[i].IsClusterCentral = false
	}

	for _, cluster := range clusters {
		// Mark central points
		for _, pointIdx := range cluster.Points {
			if pointIdx < len(a.sources) {
				a.sources[pointIdx].ClusterID = cluster.ID
				a.sources[pointIdx].IsClusterCentral = true
			}
		}
		// Mark outliers
		for _, outlierIdx := range cluster.Outliers {
			if outlierIdx < len(a.sources) {
				a.sources[outlierIdx].ClusterID = cluster.ID
				a.sources[outlierIdx].IsClusterCentral = false
			}
		}
	}

	// Sort sources by cluster
	a.sortSourcesByCluster()

	return nil
}

func (a *App) sortSourcesByCluster() {
	// Create a map to group sources by cluster
	clusterGroups := make(map[int][]Source)
	unclusteredSources := []Source{}

	// Group sources by cluster ID
	for _, source := range a.sources {
		if source.ClusterID >= 0 {
			clusterGroups[source.ClusterID] = append(clusterGroups[source.ClusterID], source)
		} else {
			unclusteredSources = append(unclusteredSources, source)
		}
	}

	// Sort within each cluster: central points first, then outliers
	for clusterID := range clusterGroups {
		cluster := clusterGroups[clusterID]
		centralPoints := []Source{}
		outliers := []Source{}

		for _, source := range cluster {
			if source.IsClusterCentral {
				centralPoints = append(centralPoints, source)
			} else {
				outliers = append(outliers, source)
			}
		}

		// Combine central points first, then outliers
		clusterGroups[clusterID] = append(centralPoints, outliers...)
	}

	// Rebuild the sources slice: unclustered first, then clusters in order
	newSources := []Source{}
	newSources = append(newSources, unclusteredSources...)

	// Add clusters in order (0, 1, 2, ...)
	maxClusterID := -1
	for clusterID := range clusterGroups {
		if clusterID > maxClusterID {
			maxClusterID = clusterID
		}
	}

	for clusterID := 0; clusterID <= maxClusterID; clusterID++ {
		if cluster, exists := clusterGroups[clusterID]; exists {
			newSources = append(newSources, cluster...)
		}
	}

	a.sources = newSources
}
