package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net/url"
	"os"
	"regexp"
	"time"

	"html"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"golang.org/x/net/publicsuffix"

)

func newApp() (*App, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %v", err)
	}

	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize screen: %v", err)
	}

	return &App{
		screen:         screen,
		selectedIdx:    0,
		expandedItems:  make(map[int]bool),
		showImportance: make(map[int]bool),
		currentPage:    0,
		itemsPerPage:   17, // 17,
		mode:           "main",
		detailIdx:      -1,
	}, nil
}

func (a *App) loadSources() error {
	/* This syntax is a method in go <https://go.dev/tour/methods/8>
	the point is to pass a pointer
	so that you can avoid passing values around
	yet still be able to modify them
	while having somewhat terser syntax than a funcion that takes
	a pointer.
	At the same time, you could achieve a similar thing with a normal
	function.

	On top of that, you can define an interface, as a type that implements
	some method. <https://go.dev/tour/methods/10>
	*/
	// fmt.Printf("Getting sources...")
	a.drawLines([]string{"Getting sources..."})
	// drawText(a.screen, 0, 0, 0, tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite),"Getting sources...")
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_POOL_URL"))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// rows, err := conn.Query(ctx, "SELECT id, title, link, date, summary, importance_bool, importance_reasoning, created_at, processed FROM sources WHERE processed = false AND EXTRACT('week' from date) = 22 ORDER BY date ASC, id ASC") // AND DATE_PART('doy', date) < 34
	rows, err := conn.Query(ctx, "SELECT id, title, link, date, summary, importance_bool, importance_reasoning, created_at, processed FROM sources WHERE processed = false ORDER BY date ASC, id ASC") 
	// AND DATE_PART('doy', date) < 35
	// AND date < '2025-09-08'
	// date '+%j'
	if err != nil {
		return fmt.Errorf("failed to query sources: %v", err)
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var s Source
		err := rows.Scan(&s.ID, &s.Title, &s.Link, &s.Date, &s.Summary, &s.ImportanceBool, &s.ImportanceReasoning, &s.CreatedAt, &s.Processed)
		if err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}
		// Clean HTML entities and tags
		s.Title = stripHTML(html.UnescapeString(s.Title))
		s.Summary = stripHTML(html.UnescapeString(s.Summary))
		sources = append(sources, s)
	}

	filtered_sources, err := filterSources(sources)
	if err != nil {
		return nil
	}
	reordered_sources, err := reorderSources(filtered_sources)
	if err != nil {
		return nil
	}
	unsimilar_sources, err := skipSourcesWithSimilarityMetric(reordered_sources)
	if err != nil {
		return nil
	}
	a.sources = unsimilar_sources

	// drawText(a.screen, 0, 2, 0, tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite),"Clustering sources")
	// fmt.Printf("Clustering sources...")
	a.drawLines([]string{"Getting sources...", "Clustering sources..."})
	// Add clustering
	err = a.clusterSources()
	if err != nil {
		fmt.Printf("Warning: clustering failed: %v\n", err)
		// Continue without clustering
	}

	// Initialize cluster styles after clustering is done
	if len(a.clusters) > 0 {
		a.clusterStyles = generateClusterStyles(len(a.clusters))
	}

	return nil
}

func (a *App) draw() {
	a.screen.Clear()
	width, height := a.screen.Size()
	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)
	selectedStyle := tcell.StyleDefault.Background(tcell.Color24).Foreground(tcell.ColorWhite)
	summaryStyle := style.Foreground(tcell.Color248)
	importanceStyle := style.Foreground(tcell.ColorYellow)

	if a.mode == "detail" {
		a.drawDetailView(width, height, style, summaryStyle, importanceStyle)
		return
	} 

	if a.mode == "help" {
		a.drawHelpView()
		return
	}

	if a.mode == "search" {
		a.drawSearchView(width, height, style)
		return
	}

	startIdx := a.currentPage * a.itemsPerPage
	endIdx := startIdx + a.itemsPerPage
	if endIdx > len(a.sources) {
		endIdx = len(a.sources)
	}

	lineIdx := 0
	for idx := startIdx; idx < endIdx; idx++ {
		source := a.sources[idx]
		// Calculate total height needed for this item
		itemHeight := 1 // Title line
		if a.expandedItems[idx] && source.Summary != "" {
			summaryLines := (len(source.Summary) + width - 3) / (width - 2)
			itemHeight += summaryLines + 1
		}
		if a.showImportance[idx] && source.ImportanceReasoning != "" {
			importanceLines := (len(source.ImportanceReasoning) + width - 3) / (width - 2)
			itemHeight += importanceLines + 1
		}

		// Check if there's enough room for the entire item
		if lineIdx+itemHeight >= height-1 {
			break
		}

		currentStyle := style
		if idx == a.selectedIdx {
			currentStyle = selectedStyle
		}

		// Display title, domain & date
		processedMark := " "
		if source.Processed {
			processedMark = "x"
		}
		
		clusterMark := " "
		distanceInfo := ""
		if source.ClusterID >= 0 {
			if source.IsClusterCentral {
				clusterMark = fmt.Sprintf("C%d", source.ClusterID)
			} else {
				clusterMark = fmt.Sprintf("O%d", source.ClusterID)
			}
			
			// Calculate and display distance to centroid
			if len(a.embeddings) > idx && len(a.clusters) > source.ClusterID && source.ClusterID >= 0 {
				cluster := a.clusters[source.ClusterID]
				if cluster.Centroid != nil {
					_ = calculateDistance(a.embeddings[idx], cluster.Centroid)
					// distance := calculateDistance(a.embeddings[idx], cluster.Centroid)
					// distanceInfo = fmt.Sprintf("-%.3f", distance)
				}
			}
		}
	
		host := ""
		parsedURL, err := url.Parse(source.Link)
		if err != nil {
			host = ""
		} else {
			host = parsedURL.Host
			shorthost, err  := publicsuffix.EffectiveTLDPlusOne(host)
			if err == nil {
				host = shorthost
			}
		}

		// Build title with colored cluster section
		titleParts := []string{}
		titleStyles := []tcell.Style{}
		
		// Processed mark
		// titleParts = append(titleParts, fmt.Sprintf("[%s]", processedMark))
		// titleStyles = append(titleStyles, currentStyle)
		
		// Cluster mark with color
		if source.ClusterID >= 0 && source.ClusterID < len(a.clusterStyles) {
			clusterStyle := a.clusterStyles[source.ClusterID]
			if idx == a.selectedIdx {
				// Keep selected background but use cluster foreground color
				_, bg, attrs := selectedStyle.Decompose()
				fg, _, _ := clusterStyle.Decompose()
				clusterStyle = tcell.StyleDefault.Foreground(fg).Background(bg).Attributes(attrs)
			}
			titleParts = append(titleParts, fmt.Sprintf("[%s] %s%s ", processedMark, clusterMark, distanceInfo))
			titleStyles = append(titleStyles, clusterStyle)
		} else {
			titleParts = append(titleParts, fmt.Sprintf("[%s] %s%s ", processedMark, clusterMark, distanceInfo))
			titleStyles = append(titleStyles, currentStyle)
		}
		
		// Rest of title
		titleParts = append(titleParts, fmt.Sprintf(" %s | %s | %s", source.Title, host, source.Date.Format("01-02")))
		titleStyles = append(titleStyles, currentStyle)
		
		// Draw title with overflow handling
		lineIdx = drawTitleWithOverflow(a.screen, 0, lineIdx, width, titleParts, titleStyles)

		// If this is the selected item and we're in expanded mode, show the summary
		if a.expandedItems[idx] && source.Summary != "" {
			lineIdx++
			if lineIdx < height {
				lineIdx = drawText(a.screen, 2, lineIdx, width-2, summaryStyle, source.Summary)
			}
		}

		// Add importance reasoning display
		if a.showImportance[idx] && source.ImportanceReasoning != "" {
			lineIdx++
			if lineIdx < height {
				lineIdx = drawText(a.screen, 2, lineIdx, width-2, importanceStyle, "Importance: "+source.ImportanceReasoning)
			}
		}
		lineIdx++
	}

	// Draw help text at the bottom
	current_item := a.selectedIdx
	num_items := len(a.sources)
	num_pages := int(math.Ceil(float64(num_items) / float64(a.itemsPerPage)))
	helpText := fmt.Sprintf("^/v: Navigate (%d/%d) | <>: Change Page (%d/%d) | Enter: View Details | H: Help", current_item+1, num_items, a.currentPage+1, num_pages)
	if a.statusMessage != "" {
		helpText =  a.statusMessage
	} else if a.failureMark {
		helpText = "[database F]"
	}
	if height > 0 {
		drawText(a.screen, 0, height-1, width, style, helpText)
	}

	a.screen.Show()
}

func (a *App) drawDetailView(width, height int, style, summaryStyle, importanceStyle tcell.Style) {
	if a.detailIdx < 0 || a.detailIdx >= len(a.sources) {
		return
	}

	source := a.sources[a.detailIdx]
	lineIdx := 0

	// Title
	titleStyle := style.Bold(true)
	lineIdx = drawText(a.screen, 0, lineIdx, width, titleStyle, source.Title)
	lineIdx++

	// URL and Date
	host := ""
	parsedURL, err := url.Parse(source.Link)
	if err == nil {
		host = parsedURL.Host
	}
	metaInfo := fmt.Sprintf("Source: %s | Date: %s", host, source.Date.Format("2006-01-02 15:04"))
	lineIdx = drawText(a.screen, 0, lineIdx, width, style, metaInfo)
	lineIdx++
	lineIdx++

	// Summary
	if source.Summary != "" {
		lineIdx = drawText(a.screen, 0, lineIdx, width, summaryStyle.Bold(true), "Summary:")
		lineIdx++
		lineIdx = drawText(a.screen, 0, lineIdx, width, summaryStyle, source.Summary)
		lineIdx++
		lineIdx++
	}

	// Importance Reasoning
	if source.ImportanceReasoning != "" {
		lineIdx = drawText(a.screen, 0, lineIdx, width, importanceStyle.Bold(true), "Importance Reasoning:")
		lineIdx++
		lineIdx = drawText(a.screen, 0, lineIdx, width, importanceStyle, source.ImportanceReasoning)
		lineIdx++
		lineIdx++
	}

	// Full URL
	lineIdx = drawText(a.screen, 0, lineIdx, width, style.Bold(true), "URL:")
	lineIdx++
	lineIdx = drawText(a.screen, 0, lineIdx, width, style, source.Link)

	// Help text at bottom
	helpText := "ESC/Backspace: Back to list | O: Open in Browser | M: Toggle mark | S: Save | W: Web Search | Q: Quit"
	if height > 0 {
		drawText(a.screen, 0, height-1, width, style, helpText)
	}

	a.screen.Show()
}

func (a *App) getInput(prompt string) string {
	// Clear bottom of screen
	width, height := a.screen.Size()
	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)
	for i := 0; i < width; i++ {
		a.screen.SetContent(i, height-1, ' ', nil, style)
	}

	// Show prompt
	for i, r := range prompt {
		a.screen.SetContent(i, height-1, r, nil, style)
	}
	a.screen.Show()

	// Get input
	var input strings.Builder
	cursorPos := len(prompt)
	for {
		ev := a.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEnter:
				return input.String()
			case tcell.KeyEscape:
				return ""
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if input.Len() > 0 {
					str := input.String()
					input.Reset()
					input.WriteString(str[:len(str)-1])
					a.screen.SetContent(cursorPos-1, height-1, ' ', nil, style)
					cursorPos--
				}
			case tcell.KeyRune:
				input.WriteRune(ev.Rune())
				a.screen.SetContent(cursorPos, height-1, ev.Rune(), nil, style)
				cursorPos++
			}
			a.screen.Show()
		}
	}
}

func (a *App) drawHelpView() {
	width, _ := a.screen.Size()
	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)

	lines := []string{
		"^/v: Navigate",
		"<>: Change Page",
		"Enter: View Details",
		"I: Show Importance",
		"O: Open in Browser",
		"M: Toggle mark",
		"S: Save",
		"W: Web Search",
		"C: Mark cluster centrals as processed", 
		"Q: Quit",
		"[C#/O#]: Cluster Central/Outlier",

	}
	lineIdx := 0
	for _, line := range lines {
		lineIdx = drawText(a.screen, 0, lineIdx, width, style, line)
		lineIdx++
	}
	a.screen.Show()
}

func (a *App) drawSearchView(width, height int, style tcell.Style) {
	if a.searchInstance == nil {
		return
	}

	selectedStyle := tcell.StyleDefault.Background(tcell.Color24).Foreground(tcell.ColorWhite)

	lineIdx := 0
	// Title
	lineIdx = drawText(a.screen, 0, lineIdx, width, style.Bold(true), "Search results for: "+a.searchInstance.GetQuery())
	lineIdx++
	lineIdx = drawText(a.screen, 0, lineIdx, width, style, "----------------------------------------")
	lineIdx++

	// Results
	results := a.searchInstance.GetResults()
	selected := a.searchInstance.GetSelected()
	for i, result := range results {
		currentStyle := style
		if i == selected {
			currentStyle = selectedStyle
		}

		// Extract host from URL
		host := ""
		if parsedURL, err := url.Parse(result.URL); err == nil {
			host = parsedURL.Host
		}

		text := fmt.Sprintf("[%s] %s", host, result.Title)
		lineIdx = drawText(a.screen, 2, lineIdx, width-2, currentStyle, text)
		lineIdx++
	}

	// Help text
	helpText := "^/v: Navigate | Enter/O: Open in Browser | S: Save to Minutes | ESC: Back"
	if height > 0 {
		drawText(a.screen, 0, height-1, width, style, helpText)
	}

	a.screen.Show()
}

func (a *App) drawLines(lines []string) {
	width, _ := a.screen.Size()
	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)

	lineIdx := 0
	for _, line := range lines {
		lineIdx = drawText(a.screen, 0, lineIdx, width, style, line)
		lineIdx++
	}
	a.screen.Show()
}

func (a *App) confirmQuit() bool {
	width, height := a.screen.Size()
	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)

	// Clear bottom line
	for i := 0; i < width; i++ {
		a.screen.SetContent(i, height-1, ' ', nil, style)
	}

	// Show prompt
	prompt := "Quit? (y/n) "
	for i, r := range prompt {
		a.screen.SetContent(i, height-1, r, nil, style)
	}
	a.screen.Show()

	// Wait for y/n response
	for {
		ev := a.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'y', 'Y':
					return true
				case 'n', 'N':
					return false
				}
			case tcell.KeyEscape:
				return false
			}
		}
	}
}

func (a *App) run() error {
	err := a.loadSources()
	if err != nil {
		return err
	}

	for {
		a.draw()

		switch ev := a.screen.PollEvent().(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				if a.mode == "detail"{
					a.mode = "main"
					a.detailIdx = -1
				} else if a.mode == "search" {
					a.mode = "main"
					a.searchInstance = nil
				} else if a.mode == "help" {
					a.mode = "main"
				} else {
					if a.confirmQuit() {
						a.waitgroup.Wait()
						return nil
					}
				}
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if a.mode == "detail" {
					a.mode = "main"
					a.detailIdx = -1
				} else if a.mode == "search" {
					a.mode = "main"
					a.searchInstance = nil
				}
			case tcell.KeyRight:
				if a.mode == "main" {
					if (a.currentPage+1)*a.itemsPerPage < len(a.sources) {
						a.currentPage++
						a.selectedIdx = a.currentPage * a.itemsPerPage
					}
				}
			case tcell.KeyLeft:
				if a.mode == "main" {
					if a.currentPage > 0 {
						a.currentPage--
						a.selectedIdx = a.currentPage * a.itemsPerPage
					}
				}
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q', 'Q':
						if a.mode == "main" {
							if a.confirmQuit() {
								a.waitgroup.Wait()
								return nil
							}
						} else {
							a.mode = "main"
						}
				case 'o', 'O':
					if a.mode == "search" && a.searchInstance != nil {
						if result := a.searchInstance.GetSelectedResult(); result != nil {
							openBrowser(result.URL)
						}
					} else if len(a.sources) > 0 {
						idx := a.selectedIdx
						if a.mode == "detail" {
							idx = a.detailIdx
						}
						openBrowser(a.sources[idx].Link)
					}
				case 'n', 'N':
					if (a.mode == "main") && a.selectedIdx < len(a.sources)-1 {
						a.selectedIdx++
					}
				case 'M':
						a.openFile()
				case 'm', 'x':
					if len(a.sources) > 0 {
						idx := a.selectedIdx
						if a.mode == "detail" {
							idx = a.detailIdx
						}
						a.markProcessed(idx, a.sources[idx])
						if a.mode == "main" {
							if a.selectedIdx < len(a.sources)-1 && (a.selectedIdx+1) < (a.currentPage+1)*a.itemsPerPage {
								a.selectedIdx++
							} else if (a.currentPage+1)*a.itemsPerPage < len(a.sources) {
								a.currentPage++
								a.selectedIdx = a.currentPage * a.itemsPerPage
							}
						}
					}
				case 'X', 'p':
					if a.mode == "main" {
						startIdx := a.currentPage * a.itemsPerPage
						endIdx := startIdx + a.itemsPerPage
						if endIdx > len(a.sources) {
							endIdx = len(a.sources)
						}
						for idx := startIdx; idx < endIdx; idx++ {
							a.markProcessed(idx, a.sources[a.selectedIdx])
						}
						if (a.currentPage+1)*a.itemsPerPage < len(a.sources) {
							a.currentPage++
							a.selectedIdx = a.currentPage * a.itemsPerPage
						}
					}
				case 'r':
					if a.mode == "main" {
						a.screen.Clear()
						a.screen.Show()
						a.currentPage = 0
						a.selectedIdx = 0
						for i := range a.expandedItems {
							a.expandedItems[i] = false
							a.showImportance[i] = false
						}

						a.sources = filterSourcesForUnread(a.sources)
					}
				case 'R':
					if a.mode == "main" {
						a.screen.Clear()
						a.screen.Show()
						a.currentPage = 0
						a.selectedIdx = 0
						a.loadSources()
					}
				case 's', 'S':
					if a.mode == "search" && a.searchInstance != nil {
						// Save search result with original source description
						if result := a.searchInstance.GetSelectedResult(); result != nil {
							// Use searchOriginIdx to get the correct source that initiated the search
							idx := a.searchOriginIdx
							if err := a.saveSearchResult(result, a.sources[idx]); err != nil {
								a.statusMessage = fmt.Sprintf("Save error: %v", err)
							} else {
								a.statusMessage = "Saved to minutes file"
								// Clear status message after 2 seconds
								go func() {
									time.Sleep(2 * time.Second)
									a.statusMessage = ""
									a.screen.Sync()
								}()
								a.markRelevantPerHumanCheck(RELEVANT_PER_HUMAN_CHECK_YES, a.searchOriginIdx)
								// Go back to main mode
								a.mode = "main"
								a.searchInstance = nil
							}
						}
					} else if len(a.sources) > 0 && (a.mode == "main" || a.mode == "detail") {
						idx := a.selectedIdx
						if a.mode == "detail" {
							idx = a.detailIdx
						}
						a.saveToFile(a.sources[idx])
						a.markRelevantPerHumanCheck(RELEVANT_PER_HUMAN_CHECK_YES, idx)
					}
				case 'w', 'W':
					if len(a.sources) > 0 && (a.mode == "main" || a.mode == "detail") {
						idx := a.selectedIdx
						if a.mode == "detail" {
							idx = a.detailIdx
						}
						if err := a.webSearch(a.sources[idx], idx); err != nil {
							a.statusMessage = fmt.Sprintf("Search error: %v", err)
						}
					}
				case 'f', 'F':
					if a.mode == "main" {
						// Add new filter
						filter_input := a.getInput("Enter filter keyword: ")
						if filter_input != "" {
							a.statusMessage = "Filtering items..."
							a.draw()
							regex_with_lookaheads := strings.ReplaceAll(filter_input, "ANY(", "(?=.*")
							// ANY(x)ANY(y) will match sentences that contain both x and y in any order, per o3
							filterRegex, err := regexp.Compile("(?i)" + regex_with_lookaheads)
							if err != nil {
								log.Printf("Error compiling regex: %v", err)
								continue
							}

							// Filter items locally and mark them in server
							var remaining_sources []Source
							for _, source := range a.sources {
								if filterRegex.MatchString(source.Title) {
									go markProcessedInServer(true, source.ID, source)
								} else {
									remaining_sources = append(remaining_sources, source)
								}
							}
							a.sources = remaining_sources

							// Reset page if needed
							if a.selectedIdx >= len(a.sources) {
								a.selectedIdx = len(a.sources) - 1
							}
							if a.selectedIdx < 0 {
								a.selectedIdx = 0
							}
							a.currentPage = a.selectedIdx / a.itemsPerPage

							// Clear status message
							a.statusMessage = ""
							a.draw()
						}
					}
				case 'i', 'I':
					if a.mode == "main" && len(a.sources) > 0 {
						a.showImportance[a.selectedIdx] = !a.showImportance[a.selectedIdx]
					}
				case 'c', 'C':
					if a.mode == "main" && len(a.sources) > 0 {
						a.markClusterPartsAsProcessed(a.selectedIdx)

						currentCluster := a.sources[a.selectedIdx].ClusterID
						clusterType := a.sources[a.selectedIdx].IsClusterCentral
						for {
							if a.sources[a.selectedIdx].ClusterID != currentCluster || a.sources[a.selectedIdx].IsClusterCentral != clusterType { // we're on a different cluster => stop
								break
							} else if a.selectedIdx < len(a.sources)-1 { // we are in the same cluster & haven't reached the end => continue
								a.selectedIdx++
								if (a.selectedIdx >= (a.currentPage+1)*a.itemsPerPage) && ((a.currentPage+1)*a.itemsPerPage < len(a.sources)) { // we are on a different page => increase page
									a.currentPage++
								}
							} else { // we're in the same cluster, but have reached the end. => stop
								break
							}
						}
					}
				case 'h', 'H':
					if a.mode == "main" {
						a.mode = "help"
					}
				}
			case tcell.KeyUp:
				if a.mode == "main" && a.selectedIdx > 0 {
					a.selectedIdx--
				} else if a.mode == "search" && a.searchInstance != nil {
					a.searchInstance.SelectPrevious()
				}
			case tcell.KeyDown:
				if a.mode == "main" && a.selectedIdx < len(a.sources)-1 && (a.selectedIdx+1) < (a.currentPage+1)*a.itemsPerPage {
					a.selectedIdx++
				} else if a.mode == "search" && a.searchInstance != nil {
					a.searchInstance.SelectNext()
				}
			case tcell.KeyEnter:
				if a.mode == "main" && len(a.sources) > 0 {
					a.mode = "detail" 
					a.detailIdx = a.selectedIdx
					a.draw()
				} else if a.mode == "detail" && len(a.sources) > 0 {
					a.mode = "main" 
				} else if a.mode == "help" {
					a.mode = "main"
				} else if a.mode == "search" && a.searchInstance != nil {
					if result := a.searchInstance.GetSelectedResult(); result != nil {
						openBrowser(result.URL)
					}
				}
			}
		case *tcell.EventResize:
			a.screen.Sync()
		}
	}
}

func (a *App) markClusterPartsAsProcessed(selectedIdx int) {
	if selectedIdx >= len(a.sources) {
		return
	}
	
	selectedSource := a.sources[selectedIdx]
	partType := selectedSource.IsClusterCentral
	partTypeName := "central"
	if !partType {
		partTypeName = "outlier"
	}
	if selectedSource.ClusterID < 0 {
		return // Not in a cluster
	}
	
	clusterID := selectedSource.ClusterID
	markedCount := 0
	
	// Mark all central members of this cluster as processed
	for i := range a.sources {
		if a.sources[i].ClusterID == clusterID && a.sources[i].IsClusterCentral == partType {
			a.markProcessed(i, a.sources[i])
			markedCount++
		}
	}
	
	if markedCount > 0 {
		a.statusMessage = fmt.Sprintf("Marked %d %s cluster members as processed", markedCount, partTypeName)
		// Clear status message after 2 seconds
		go func() {
			time.Sleep(500 * time.Millisecond)
			a.statusMessage = ""
			a.screen.Sync()
		}()
	}
}

func generateClusterStyles(numClusters int) []tcell.Style {
	if numClusters == 0 {
		return []tcell.Style{}
	}
	
	styles := make([]tcell.Style, numClusters)
	
	// Predefined colors that work well in terminals
	colors := []tcell.Color{
		tcell.ColorRed,
		tcell.ColorGreen, 
		tcell.ColorBlue,
		tcell.ColorYellow,
		// tcell.ColorMagenta,
		// tcell.ColorCyan,
		tcell.ColorOrange,
		tcell.ColorPurple,
		tcell.ColorLime,
		tcell.ColorPink,
		tcell.ColorTeal,
		tcell.ColorSilver,
		tcell.ColorGold,
		tcell.ColorCoral,
		tcell.ColorSkyblue,
		tcell.ColorViolet,
	}
	
	for i := 0; i < numClusters; i++ {
		// Cycle through predefined colors
		colorIndex := i % len(colors)
		styles[i] = tcell.StyleDefault.Foreground(colors[colorIndex])
	}
	
	return styles
}

func drawTitleWithOverflow(screen tcell.Screen, x, y, maxWidth int, titleParts []string, titleStyles []tcell.Style) int {
	if len(titleParts) == 0 {
		return y
	}
	
	// First, draw the fixed parts (processed mark and cluster mark) on the first line
	currentX := x
	currentY := y
	fixedPartsCount := 1 // processed mark and cluster mark
	
	// Draw fixed parts that should always be on the first line
	for i := 0; i < fixedPartsCount && i < len(titleParts); i++ {
		part := titleParts[i]
		style := titleStyles[i]
		
		for j, r := range part {
			if currentX+j < maxWidth {
				screen.SetContent(currentX+j, currentY, r, nil, style)
			}
		}
		currentX += len(part)
	}
	
	// Handle the remaining parts (title, host, date) with word wrapping
	if len(titleParts) > fixedPartsCount {
		remainingText := titleParts[fixedPartsCount]
		remainingStyle := titleStyles[fixedPartsCount]
		
		// Calculate remaining width on first line
		remainingWidth := maxWidth - currentX
		
		// Split the remaining text into words
		words := strings.Fields(remainingText)
		if len(words) == 0 {
			return currentY
		}
		
		currentLine := ""
		
		for _, word := range words {
			// Check if adding this word would exceed the line width
			testLine := currentLine
			if testLine != "" {
				testLine += " "
			}
			testLine += word
			
			if len(testLine) <= remainingWidth {
				// Word fits on current line
				currentLine = testLine
			} else {
				// Word doesn't fit, draw current line and start new one
				if currentLine != "" {
					// Draw current line
					for i, r := range currentLine {
						if currentX+i < maxWidth {
							screen.SetContent(currentX+i, currentY, r, nil, remainingStyle)
						}
					}
					currentY++
					currentX = x + 2 // Indent continuation lines
					remainingWidth = maxWidth - currentX
				}
				currentLine = word
			}
		}
		
		// Draw final line
		if currentLine != "" {
			for i, r := range currentLine {
				if currentX+i < maxWidth {
					screen.SetContent(currentX+i, currentY, r, nil, remainingStyle)
				}
			}
		}
	}
	
	return currentY
}

func drawText(screen tcell.Screen, x, y, maxWidth int, style tcell.Style, text string) int {
	words := strings.Fields(text)
	if len(words) == 0 {
		return y
	}

	currentLine := words[0]
	currentY := y

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= maxWidth {
			currentLine += " " + word
		} else {
			// Draw current line
			for i, r := range currentLine {
				screen.SetContent(x+i, currentY, r, nil, style)
			}
			currentY++
			currentLine = word
		}
	}

	// Draw final line
	for i, r := range currentLine {
		screen.SetContent(x+i, currentY, r, nil, style)
	}

	return currentY
}

func main() {
	logFile, err := os.OpenFile("data/client.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	mw := io.Writer(logFile)
	log.SetOutput(mw)

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	app, err := newApp()
	if err != nil {
		log.Fatalf("Could not create app: %v", err)
	}

	if err := app.run(); err != nil {
		app.screen.Fini()
		log.Fatalf("Error running app: %v", err)
	}

	app.screen.Fini()
}
