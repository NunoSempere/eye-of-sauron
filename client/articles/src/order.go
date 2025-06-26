package main

import (
  "os"
    "log"
    "regexp"
  "bufio"
  "strings"
  "sort"
)

func readTopicsFromFile(filepath string) ([]Topic, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var topics []Topic
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines and comments
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		keywords := strings.Split(parts[1], ",")
		// Clean up each keyword
		var cleanKeywords []string
		for _, k := range keywords {
			k = strings.TrimSpace(k)
			if k != "" {
				cleanKeywords = append(cleanKeywords, k)
			}
		}
		if len(cleanKeywords) > 0 {
			topics = append(topics, Topic{name: name, keywords: cleanKeywords})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}

func reorderSources(sources []Source) ([]Source, error) {
    var reordered_sources []Source
    remaining_sources := sources

    topics, err := readTopicsFromFile("data/topics.txt")
    if err != nil {
        log.Printf("Error loading topics: %v", err)
        return sources, err
    }

    for _, topic := range topics {
        var topic_regexes []*regexp.Regexp
        for _, regex_string := range topic.keywords {
            regex, err := regexp.Compile("(?i)" + regex_string)
            if err != nil {
                log.Printf("Regex error: %v", err)
                return nil, err
            }
            topic_regexes = append(topic_regexes, regex)
        }

        var new_remaining_sources []Source
        var topic_sources []Source
        for _, source := range remaining_sources {
            match := testStringAgainstRegexes(topic_regexes, source.Title)
            if match {
                topic_sources = append(topic_sources, source)
            } else {
                new_remaining_sources = append(new_remaining_sources, source)
            }
        }

        // Sort the topic_sources alphabetically by title
        sort.Slice(topic_sources, func(i, j int) bool {
            return topic_sources[i].Title < topic_sources[j].Title
        })
        reordered_sources = append(reordered_sources, topic_sources...)
        remaining_sources = new_remaining_sources
    }

    // Append remaining sources that didn't fit into any topic, sorted alphabetically
    sort.Slice(remaining_sources, func(i, j int) bool {
        return remaining_sources[i].Title < remaining_sources[j].Title
    })
    // Show sources that don't fit neatly into a topic first
    reordered_sources = append(remaining_sources, reordered_sources...)

    return reordered_sources, nil
}
