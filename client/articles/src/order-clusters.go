package main

import (
  "os"
  "log"
  "regexp"
  "bufio"
  "strings"
)

func readTopicsFromFileForClusters(filepath string) ([]Topic, error) {
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

func  testSourceListAgainstRegexes(rs []*regexp.Regexp, ss []Source) bool {
	for _, source := range ss {
		if source.IsClusterCentral {
			title := source.Title
			for _, r := range rs {
				if r.MatchString(title) {
					return true
				}
			}
		} 
	}
	return false
}


func reorderClusters(sss [][]Source) ([][]Source, error) {
    var reordered_source_lists [][]Source
    remaining_source_lists := sss

    topics, err := readTopicsFromFile("data/topics.txt")
    if err != nil {
        log.Printf("Error loading topics: %v", err)
        return sss, err
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

        var new_remaining_source_lists [][]Source
        var topic_source_lists [][]Source
        for _, source_list := range remaining_source_lists {
            match := testSourceListAgainstRegexes(topic_regexes, source_list)
            if match {
                topic_source_lists = append(topic_source_lists, source_list)
            } else {
                new_remaining_source_lists = append(new_remaining_source_lists, source_list)
            }
        }

        reordered_source_lists = append(reordered_source_lists, topic_source_lists...)
        remaining_source_lists = new_remaining_source_lists
    }

  	// fmt.Printf("%v", remaining_source_lists)
  	// fmt.Printf("Press enter to continue")
    // bufio.NewReader(os.Stdin).ReadBytes('\n') // wait for keyword

    // Append remaining sources that didn't fit into any topic, sorted alphabetically
    // Show sources that don't fit neatly into a topic first

    reordered_source_lists = append(reordered_source_lists, remaining_source_lists...)

    return reordered_source_lists, nil
}
