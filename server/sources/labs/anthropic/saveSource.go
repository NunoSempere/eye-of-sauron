package main

import (
	"git.nunosempere.com/NunoSempere/news/lib/pgx"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

// SaveSource saves to AI database always, and to main database if passes_filters is true
func SaveSource(source types.ExpandedSource, passes_filters bool) {
	pgx.SaveSourceConditional(source, passes_filters)
}
