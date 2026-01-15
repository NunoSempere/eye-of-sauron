package main

import (
	"context"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
	"github.com/tiktoken-go/tokenizer"

	"github.com/NunoSempere/hdbscan/hdbscan"
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

func getEmbeddingsStaggered(texts []string, token string) ([][]float64, error) {
	//  max 300000 tokens per request, using 280000 as threshold for safety
	maxTokensPerBatch := 280000
	log.Printf("[CLUSTERING] Starting getEmbeddingsStaggered with %d texts (max tokens per batch: %d)", len(texts), maxTokensPerBatch)

	enc, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
    log.Printf("Error getting the tokenizer")
    log.Printf("%v", err)
  }

	es := [][]float64{}
	n := 0
	last_not_in_batch := 0
	batchCount := 0
	for i, text := range texts {
		m, err := enc.Count(text)
		if err != nil {
			log.Printf("Error counting tokens: %v", err)
				return [][]float64{}, err
		}
		if n < maxTokensPerBatch && (n+m > maxTokensPerBatch) {
			batchCount++
			log.Printf("[CLUSTERING] Batch %d: Getting embeddings for texts %d to %d (token count: %d)", batchCount, last_not_in_batch, i, n)
			es_batch, err := getEmbeddings(texts[last_not_in_batch:i], token)
			if err != nil {
				log.Printf("Error getting embeddings: %v", err)
				return [][]float64{}, err
			}
			log.Printf("[CLUSTERING] Batch %d: Received %d embeddings", batchCount, len(es_batch))
			es = append(es, es_batch...)
			last_not_in_batch = i
			n = 0
		} else {
			n = n + m
		}
	}

	// Append last batch
	batchCount++
	log.Printf("[CLUSTERING] Final batch %d: Getting embeddings for texts %d to %d (token count: %d)", batchCount, last_not_in_batch, len(texts), n)
	es_batch, err := getEmbeddings(texts[last_not_in_batch:len(texts)], token)
	if err != nil {
		log.Printf("Error getting embeddings: %v", err)
		return [][]float64{}, err
	}
	log.Printf("[CLUSTERING] Final batch %d: Received %d embeddings", batchCount, len(es_batch))
	es = append(es, es_batch...)

	log.Printf("[CLUSTERING] Total embeddings generated: %d (expected: %d)", len(es), len(texts))
	if len(es) > 0 {
		log.Printf("[CLUSTERING] Embedding dimension: %d", len(es[0]))
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

// calculateCentroid computes the arithmetic mean of all points in a cluster
func calculateCentroid(points []int, data [][]float64) []float64 {
	if len(points) == 0 {
		return nil
	}
	
	dimensions := len(data[0])
	centroid := make([]float64, dimensions)
	
	// Sum all coordinates
	for _, pointIdx := range points {
		for dim := 0; dim < dimensions; dim++ {
			centroid[dim] += data[pointIdx][dim]
		}
	}
	
	// Calculate mean
	for dim := 0; dim < dimensions; dim++ {
		centroid[dim] /= float64(len(points))
	}
	
	return centroid
}

// calculateDistance computes Euclidean distance between two points
func calculateDistance(point1, point2 []float64) float64 {
	if len(point1) != len(point2) {
		return 0
	}
	
	sum := 0.0
	for i := 0; i < len(point1); i++ {
		diff := point1[i] - point2[i]
		sum += diff * diff
	}
	
	return math.Sqrt(sum)
}

func getClusters(data [][]float64) []Cluster {
	log.Printf("[CLUSTERING] getClusters called with %d data points", len(data))
	if len(data) > 0 {
		log.Printf("[CLUSTERING] Data dimension: %d", len(data[0]))
	}

	minimumClusterSize := 3 //
	minimumSpanningTree := false
	log.Printf("[CLUSTERING] Parameters: minimumClusterSize=%d, minimumSpanningTree=%v", minimumClusterSize, minimumSpanningTree)

	// create
	log.Printf("[CLUSTERING] Creating HDBSCAN clustering object...")
	clustering, err := hdbscan.NewClustering(data, minimumClusterSize)
	if err != nil {
		log.Printf("Error creating clustering: %v\n", err)
		return []Cluster{}
	}
	log.Printf("[CLUSTERING] HDBSCAN clustering object created successfully")

	log.Printf("[CLUSTERING] Running outlier detection...")
	clustering = clustering.OutlierDetection()
	log.Printf("[CLUSTERING] Running HDBSCAN algorithm...")
	clustering.Run(hdbscan.EuclideanDistance, hdbscan.VarianceScore, minimumSpanningTree)
	log.Printf("[CLUSTERING] HDBSCAN algorithm completed")

	log.Printf("[CLUSTERING] Extracting clusters...")
	result := extractClusters(*clustering)
	log.Printf("[CLUSTERING] Extracted %d clusters", len(result))

	// Log cluster details
	for i, cluster := range result {
		log.Printf("[CLUSTERING] Cluster %d: %d central points, %d outliers", i, len(cluster.Points), len(cluster.Outliers))
	}

	// Calculate centroids for each cluster
	log.Printf("[CLUSTERING] Calculating centroids...")
	for i := range result {
		result[i].Centroid = calculateCentroid(result[i].Points, data)
		if result[i].Centroid != nil {
			log.Printf("[CLUSTERING] Cluster %d centroid calculated (dimension: %d)", i, len(result[i].Centroid))
		}
	}

	return result
}

func cleanTitleForEmbedding(title string) string {
	clean := strings.ReplaceAll(title, "<b>", "")
	clean = strings.ReplaceAll(clean, "</b>", "")
	clean = strings.ReplaceAll(clean, "'", "'")
	return clean
}

// assignRandomClusters creates fake clusters and distances for testing
func assignRandomClusters(sources []Source) ([]Cluster, [][]float64) {
	rand.Seed(time.Now().UnixNano())
	
	numSources := len(sources)
	if numSources == 0 {
		return []Cluster{}, [][]float64{}
	}
	
	// Create 3-5 random clusters
	numClusters := 3 + rand.Intn(3)
	clusters := make([]Cluster, numClusters)
	
	// Initialize clusters
	for i := 0; i < numClusters; i++ {
		clusters[i] = Cluster{
			ID:       i,
			Points:   []int{},
			Outliers: []int{},
			Centroid: make([]float64, 512), // Fake embedding dimension
		}
		
		// Random centroid
		for j := 0; j < 512; j++ {
			clusters[i].Centroid[j] = rand.Float64()*2 - 1 // Random values between -1 and 1
		}
	}
	
	// Assign sources to clusters randomly
	for i := 0; i < numSources; i++ {
		clusterID := rand.Intn(numClusters)
		
		// 70% chance of being central, 30% chance of being outlier
		if rand.Float64() < 0.7 {
			clusters[clusterID].Points = append(clusters[clusterID].Points, i)
		} else {
			clusters[clusterID].Outliers = append(clusters[clusterID].Outliers, i)
		}
	}
	
	// Create fake embeddings (needed for distance calculations)
	embeddings := make([][]float64, numSources)
	for i := 0; i < numSources; i++ {
		embeddings[i] = make([]float64, 512)
		for j := 0; j < 512; j++ {
			embeddings[i][j] = rand.Float64()*2 - 1
		}
	}
	
	return clusters, embeddings
}

func (a *App) clusterSources() error {
	log.Printf("[CLUSTERING] clusterSources called with %d sources", len(a.sources))
	if len(a.sources) < 2 {
		log.Printf("[CLUSTERING] Not enough sources to cluster (need at least 2, have %d)", len(a.sources))
		return nil // Not enough sources to cluster
	}

	// Check for random clustering flag
	if os.Getenv("RANDOM_CLUSTERS") == "true" {
		log.Println("[CLUSTERING] Using random clustering for testing")
		clusters, embeddings := assignRandomClusters(a.sources)
		
		a.embeddings = embeddings
		a.clusters = clusters
		
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

	openaiKey := os.Getenv("OPENAI_KEY")
	if openaiKey == "" {
		log.Println("[CLUSTERING] OPENAI_KEY not found, skipping clustering")
		return nil
	}
	log.Printf("[CLUSTERING] OPENAI_KEY found, proceeding with real clustering")

	// Extract titles for embedding
	texts := make([]string, len(a.sources))
	for i, source := range a.sources {
		texts[i] = cleanTitleForEmbedding(source.Title) + "\n" + source.Summary
	}
	log.Printf("[CLUSTERING] Prepared %d texts for embedding", len(texts))

	// Get embeddings
	a.drawLines([]string{"Getting sources...", "Clustering sources...", "  Getting embeddings..."})
	log.Printf("[CLUSTERING] Calling getEmbeddingsStaggered...")
	embeddings, err := getEmbeddingsStaggered(texts, openaiKey)
	if err != nil {
		log.Printf("[CLUSTERING] Error getting embeddings: %v\n", err)
		return err
	}
	log.Printf("[CLUSTERING] Successfully retrieved %d embeddings", len(embeddings))

	// Store embeddings in App for distance calculations
	a.embeddings = embeddings

	// Get clusters
	a.drawLines([]string{"Getting sources...", "Clustering sources...", "  Getting embeddings...", "  Calculating clusters... [this may take a while]"})
	log.Printf("[CLUSTERING] Calling getClusters with %d embeddings...", len(embeddings))
	clusters := getClusters(embeddings)
	log.Printf("[CLUSTERING] getClusters returned %d clusters", len(clusters))

	// Store clusters in App
	a.clusters = clusters

	// Assign cluster information to sources
	log.Printf("[CLUSTERING] Initializing all sources with ClusterID=-1 and IsClusterCentral=false")
	for i := range a.sources {
		a.sources[i].ClusterID = -1 // Default: no cluster
		a.sources[i].IsClusterCentral = false
	}

	log.Printf("[CLUSTERING] Assigning cluster information to sources...")
	centralCount := 0
	outlierCount := 0
	for _, cluster := range clusters {
		// Mark central points
		for _, pointIdx := range cluster.Points {
			if pointIdx < len(a.sources) {
				a.sources[pointIdx].ClusterID = cluster.ID
				a.sources[pointIdx].IsClusterCentral = true
				centralCount++
			} else {
				log.Printf("[CLUSTERING] WARNING: Central point index %d is out of bounds (sources length: %d)", pointIdx, len(a.sources))
			}
		}
		// Mark outliers
		for _, outlierIdx := range cluster.Outliers {
			if outlierIdx < len(a.sources) {
				a.sources[outlierIdx].ClusterID = cluster.ID
				a.sources[outlierIdx].IsClusterCentral = false
				outlierCount++
			} else {
				log.Printf("[CLUSTERING] WARNING: Outlier index %d is out of bounds (sources length: %d)", outlierIdx, len(a.sources))
			}
		}
	}
	log.Printf("[CLUSTERING] Assigned %d central points and %d outliers", centralCount, outlierCount)

	// Count unclustered sources
	unclusteredCount := 0
	for _, source := range a.sources {
		if source.ClusterID == -1 {
			unclusteredCount++
		}
	}
	log.Printf("[CLUSTERING] %d sources remain unclustered", unclusteredCount)

	// Sort sources by cluster
	a.sortSourcesByCluster()

	return nil
}

func (a *App) sortSourcesByCluster() {
	log.Printf("[CLUSTERING] sortSourcesByCluster called")
	// Create a map to group sources by cluster
	clusterGroups := make(map[int][]Source)
	unclusteredSources := []Source{}

	// Group sources by cluster ID
	log.Printf("[CLUSTERING] Grouping sources by cluster ID...")
	for _, source := range a.sources {
		if source.ClusterID >= 0 {
			clusterGroups[source.ClusterID] = append(clusterGroups[source.ClusterID], source)
		} else {
			unclusteredSources = append(unclusteredSources, source)
		}
	}
	log.Printf("[CLUSTERING] Found %d unclustered sources and %d cluster groups", len(unclusteredSources), len(clusterGroups))

	// Sort within each cluster: central points first, then outliers
	log.Printf("[CLUSTERING] Sorting within each cluster (central points first, then outliers)...")
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

		log.Printf("[CLUSTERING] Cluster %d: %d central, %d outliers", clusterID, len(centralPoints), len(outliers))
		// Combine central points first, then outliers
		clusterGroups[clusterID] = append(centralPoints, outliers...)
	}

	// Sort the cluster group themselves
	a.drawLines([]string{"Getting sources...", "Clustering sources...", "  Getting embeddings...", "  Calculating clusters... [this may take a while]", "  Grouping clusters"})
	log.Printf("[CLUSTERING] Preparing to reorder clusters by topic...")
	sss := [][]Source{}
	for _, ss := range clusterGroups {
		sss = append(sss, ss)
	}
	log.Printf("[CLUSTERING] Calling reorderClusters with %d cluster groups...", len(sss))
	sss_ordered_by_cluster, err := reorderClusters(sss)
	if err != nil {
		log.Printf("[CLUSTERING] Error sorting clusters by topic: %v", err)
	} else {
		log.Printf("[CLUSTERING] Successfully reordered clusters")
	}


	// Rebuild the sources slice: unclustered first, then clusters in order
	log.Printf("[CLUSTERING] Rebuilding sources slice...")
	newSources := []Source{}
	newSources = append(newSources, unclusteredSources...)
	log.Printf("[CLUSTERING] Added %d unclustered sources", len(unclusteredSources))

	totalClusteredSources := 0
	for i, ss := range sss_ordered_by_cluster {
		log.Printf("[CLUSTERING] Adding cluster group %d with %d sources", i, len(ss))
		newSources = append(newSources, ss...)
		totalClusteredSources += len(ss)
	}
	log.Printf("[CLUSTERING] Added %d clustered sources across %d cluster groups", totalClusteredSources, len(sss_ordered_by_cluster))

	// Add clusters in order (0, 1, 2, ...)
	/*
	maxClusterID := -1
	for clusterID := range clusterGroups {
		if clusterID > maxClusterID {
			maxClusterID = clusterID
		}
	}

	for clusterID := 0; clusterID <= maxClusterID; clusterID++ {
		if clusterPoints, exists := clusterGroups[clusterID]; exists {
			newSources = append(newSources, clusterPoints...)
		}
	}
	*/

	a.sources = newSources
	log.Printf("[CLUSTERING] sortSourcesByCluster completed. Final sources count: %d (was: %d)", len(newSources), len(a.sources))
}
