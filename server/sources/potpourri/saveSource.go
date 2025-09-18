package main

import (
	"git.nunosempere.com/NunoSempere/news/lib/pgx"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

func SaveSource(source types.ExpandedSource) {
	pgx.SaveSource(source)
}
