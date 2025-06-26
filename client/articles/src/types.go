package main

import (
	"sync"
	"time"
		"github.com/gdamore/tcell/v2"

)

type Source struct {
	ID                    int
	Title                 string
	Link                  string
	Date                  time.Time
	Summary               string
	ImportanceBool        bool
	ImportanceReasoning   string
	CreatedAt             time.Time
	Processed             bool
	RelevantPerHumanCheck string
}

var RELEVANT_PER_HUMAN_CHECK_NO = "no"
var RELEVANT_PER_HUMAN_CHECK_YES = "yes"
var RELEVANT_PER_HUMAN_CHECK_DEFAULT = "maybe"

type App struct {
	screen         tcell.Screen
	sources        []Source
	selectedIdx    int
	expandedItems  map[int]bool
	showImportance map[int]bool
	currentPage    int
	itemsPerPage   int
	failureMark    bool
	waitgroup      sync.WaitGroup
	statusMessage  string
	detailMode     bool  // New field to track if we're in detail view
	detailIdx      int   // Index of item being viewed in detail
}

type Topic struct {
	name     string
	keywords []string
}


