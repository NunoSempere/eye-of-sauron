package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/types"
)

type GKGNode struct {
	Title    string
	Link     string
	GKG_Date string
}

func cut(s string, delimiter string, n int) (string, error) {
	parts := strings.Split(s, delimiter)
	if len(parts) < n {
		err := fmt.Errorf("Expected at least %d parts, got: %d", n, len(parts))
		log.Println(err)
		return "", err
	}
	return parts[n-1], nil
}

func findGKGNodeTitle(data string) (string, error) {
	// Compile a regular expression that captures the content between <PAGE_TITLE> and </PAGE_TITLE>
	re, err := regexp.Compile("<PAGE_TITLE>(.*?)</PAGE_TITLE>")
	if err != nil {
		log.Printf("Error in findGKGNodeTitle: %v", err)
		return "", err
	}
	matches := re.FindStringSubmatch(data)
	if len(matches) < 2 {
		error_msg := "In findGKGNodeTitle, title not found"
		log.Printf("%s", error_msg)
		return "", fmt.Errorf(error_msg)
	}
	return matches[1], nil
}

func processGKGLines(r io.Reader) ([]GKGNode, error) {
	var nodes []GKGNode
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024) // otherwise some lines are too long
	scanner.Buffer(buf, 1024*1024)

	i := 0
	for scanner.Scan() {
		if i > 10000 {
			break
		}
		i++
		line := scanner.Text()

		date, err := cut(line, "\t", 2)
		if err != nil {
			return nil, err
		}

		link, err := cut(line, "\t", 5)
		if err != nil {
			return nil, err
		}

		counts, err := cut(line, "\t", 6)
		if err != nil {
			return nil, err
		}
		xml, err := cut(line, "\t", 27)
		if err != nil {
			return nil, err
		}
		title, err := findGKGNodeTitle(xml)
		if err != nil {
			title = "New GKG node with > 100 deaths or > 1K wounded; though GKG can be mistaken"
		}

		report := false
		for _, count := range strings.Split(counts, ";") {
			count_parts := strings.Split(count, "#")
			if len(count_parts) < 2 {
				break
			}
			count_type := count_parts[0]
			count_num, err := strconv.Atoi(count_parts[1])
			// fmt.Printf("\t\t%d\n", count_num)
			if err != nil {
				// Not an error; some items don't have to have counts.
				break
			}
			if count_type == "KILL" && count_num > 100 {
				// fmt.Printf("counts")
				report = true
			}
			if count_type == "WOUND" && count_num > 1000 {
				// fmt.Printf("counts")
				report = true
			}
		}
		if report {
			/*
				if err == nil {
					title = "New GKG node with > 100 killed or > 1000 wounded—though note that GKG can be unreliable"
					fmt.Printf("Title not found in line: %s", line)
				}
			*/
			new_node := GKGNode{Title: title, Link: link, GKG_Date: date}
			nodes = append(nodes, new_node)
			// log.Printf("Node %v\n", new_node)
			// fmt.Printf("%s\n", link)
			// fmt.Printf("%s\n\n", line)
			// fmt.Printf("\t%s\n", counts)
		}

	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading content: %s\n", err)
	}

	return nodes, nil
}

func SearchGKG() ([]types.Source, error) {
	// Fetch last update
	resp, err := http.Get("http://data.gdeltproject.org/gdeltv2/lastupdate.txt")
	if err != nil {
		return nil, fmt.Errorf("fetching lastupdate.txt: %w", err)
	}

	// Extract link of zipfile
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	var gkg_link string
	line_count := 0
	for scanner.Scan() {
		if line_count == 2 {
			line := scanner.Text()
			gkg_link, err = cut(line, " ", 3)
			if err != nil {
				return nil, err
			}
			break
		}
		line_count++
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	} else if line_count < 2 {
		return nil, fmt.Errorf("lastupdate.txt doesn't have three lines")
	}

	// Download zipfile
	time.Sleep(30 * time.Second)
	log.Printf("gkg link: %v", gkg_link)
	gkg_resp, err := http.Get(gkg_link)
	log.Printf("http response: %v", gkg_resp)

	if err != nil {
		return nil, fmt.Errorf("downloading file: %w", err)
	}

	defer gkg_resp.Body.Close()

	// read zipfile
	gkg_data, err := io.ReadAll(gkg_resp.Body)

	if err != nil {
		return nil, fmt.Errorf("reading download response body: %w", err)
	}

	gkg_reader := bytes.NewReader(gkg_data)

	// Opening zip content
	zip_reader, err := zip.NewReader(gkg_reader, int64(len(gkg_data)))

	if err != nil {
		gkg_reader.Seek(0, 0)
		return nil, fmt.Errorf("Error reading zip content: %w\ndata: %v", err, gkg_data)
	}

	if len(zip_reader.File) == 0 {
		return nil, fmt.Errorf("Zip file is empty")
	}

	zipped_file, err := zip_reader.File[0].Open()

	if err != nil {
		return nil, fmt.Errorf("opening zipped file: %w", err)
	}

	defer zipped_file.Close()

	// return processGKGLines(zipped_file)
	var sources []types.Source
	nodes, err := processGKGLines(zipped_file)
	if err != nil {
		return sources, nil
	}
	for i, _ := range nodes {
		date, err := time.Parse("20060102150405", nodes[i].GKG_Date)
		if err != nil {
			date = time.Now()
		}
		sources = append(sources, types.Source{Title: nodes[i].Title, Link: nodes[i].Link, Date: date})
	}
	return sources, nil
}
