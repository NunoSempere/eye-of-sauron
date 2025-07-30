package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"log"
	"github.com/adrg/strutil/metrics"
)

// Filtering
// Could eventually move to a new file
func readRegexesFromFile(filepath string) ([]*regexp.Regexp, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var regexes []*regexp.Regexp
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		regexStr := scanner.Text()
		if len(regexStr) > 1 && regexStr[0] != '#' {
			regex, err := regexp.Compile("(?i)" + regexStr) // make case insensitive
			if err != nil {
				return nil, err // exit at first failure
			}
			regexes = append(regexes, regex)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return regexes, nil
}

func testStringAgainstRegexes(rs []*regexp.Regexp, s string) bool {
	for _, r := range rs {
		if r.MatchString(s) {
			return true
		}
	}
	return false
}

func isSourceRepeat(i int, sources []Source) bool {
	// TODO: maybe make this a bit more sophisticated, and mark as repeat even if they are not exactly the same but
	// also if they are pretty near according to some measure of distance
	for j, _ := range sources {
		if i != j && strings.ToUpper(cleanTitle(sources[i].Title)) == strings.ToUpper(cleanTitle(sources[j].Title)) {
			return true
		}
	}
	return false
}

func filterSources(sources []Source) ([]Source, error) {
	var filtered_sources []Source
	regexes, err := readRegexesFromFile("data/filters.txt")
	if err != nil {
		log.Printf("Error loading regexes: %v", err)
		return filtered_sources, err
	}

	for i, source := range sources {
		match := testStringAgainstRegexes(regexes, source.Title)
		is_repeat := isSourceRepeat(i, sources) // TODO: maybe extract this into own loop
		if !match && !is_repeat {
			filtered_sources = append(filtered_sources, source)
		} else {
			log.Printf("Skipped over: %s", source.Title)
			go markProcessedInServer(true, source.ID, source)
		}
	}
	return filtered_sources, nil
}

func filterSourcesForUnread(sources []Source) []Source {
	var unread_sources []Source
	for _, source := range sources {
		if !source.Processed {
			unread_sources = append(unread_sources, source)
		} else {
		}
	}
	return unread_sources
}


func skipSourcesWithSimilarityMetric(sources []Source) ([]Source, error) {
	if len(sources) < 2 {
		return sources, nil
	}

	new_sources := []Source{sources[0]}

	last_title := sources[0].Title
	for i := 1; i < len(sources); i++ {
		title_i := sources[i].Title
		if len(title_i) > 30 && len(last_title) > 30 {
			hamming := metrics.NewHamming()
			distance := hamming.Distance(title_i[:30], last_title[:30])
			if distance <= 4 {
				go markProcessedInServer(true, sources[i].ID, sources[i])
				continue
			} 
			last_title = title_i
		} 
		new_sources = append(new_sources, sources[i])
	}
	return new_sources, nil
}
