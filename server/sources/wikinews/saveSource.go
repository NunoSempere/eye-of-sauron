package main

import (
	"git.nunosempere.com/NunoSempere/news/lib/pgx"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

// SaveSource connects to the database and inserts the expanded source.
func SaveSource(source types.ExpandedSource) {
	pgx.SaveSource(source)
}
