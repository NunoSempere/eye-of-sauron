package main

import (
	"git.nunosempere.com/NunoSempere/news/lib/filters"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

func FilterAndExpandSource(source types.Source, openai_key string, database_url string) (types.ExpandedSource, bool) {
	return filters.StandardProcessingPipeline(source, openai_key, database_url)
}
