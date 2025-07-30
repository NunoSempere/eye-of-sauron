package main

import (
	"strings"
)

func padStringWithWhitespace(s string, n int) string {
	if len(s) > n {
		return s
	}
	padding := strings.Repeat(" ", n-len(s))
	return s + padding
}

func cleanTitle(s string) string {
  s2 := strings.ReplaceAll(s, "<b>", "")
	s3 := strings.ReplaceAll(s2, "</b>", "")
	s4 := strings.ReplaceAll(s3, "&#39;", "'")
	return s4
	/*
	n := 10
	if len(s) < n {
		return s
	} 

	if pos := strings.LastIndex(s[len(s)-n:], " – "); pos != -1 {
		if pos > n {
			return s[:len(s)-n+pos]
		}
	} 
	return s
	*/
}


func stripHTML(s string) string {
	var result strings.Builder
	var inTag bool
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

