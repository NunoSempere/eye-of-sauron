package types

import (
	"time"
)

type Source struct {
	Title  string
	Link   string
	Date   time.Time
	Origin string
}

type CacheChecker func(string) (bool, error)
type CacheAdder func(string) error

type ProspectorInput struct {
	Article           Source
	Prospector_type   string
	Openai_token      string
	Postmark_token    string
	LinkCacheChecker  CacheChecker
	LinkCacheAdder    CacheAdder
	TitleCacheChecker CacheChecker
	TitleCacheAdder   CacheAdder
}

type ExpandedSource struct {
	Title               string
	Link                string
	Date                time.Time
	Summary             string
	ImportanceBool      bool
	ImportanceReasoning string
	Origin              string
}

type Filter func(ExpandedSource) (ExpandedSource, bool)
