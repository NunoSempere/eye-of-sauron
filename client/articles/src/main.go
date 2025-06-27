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

	"html"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
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
		detailMode:     false,
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
	fmt.Printf("Getting sources...")
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_POOL_URL"))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// rows, err := conn.Query(ctx, "SELECT id, title, link, date, summary, importance_bool, importance_reasoning, created_at, processed FROM sources WHERE processed = false AND EXTRACT('week' from date) = 22 ORDER BY date ASC, id ASC") // AND DATE_PART('doy', date) < 34
	rows, err := conn.Query(ctx, "SELECT id, title, link, date, summary, importance_bool, importance_reasoning, created_at, processed FROM sources WHERE processed = false ORDER BY date ASC, id ASC") // AND DATE_PART('doy', date) < 34
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

	fmt.Printf("\nClustering surces...")
	// Add clustering
	err = a.clusterSources()
	if err != nil {
		fmt.Printf("Warning: clustering failed: %v\n", err)
		// Continue without clustering
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

	if a.detailMode {
		a.drawDetailView(width, height, style, summaryStyle, importanceStyle)
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
		if source.ClusterID >= 0 {
			if source.IsClusterCentral {
				clusterMark = fmt.Sprintf("C%d", source.ClusterID)
			} else {
				clusterMark = fmt.Sprintf("O%d", source.ClusterID)
			}
		}
		
		host := ""
		parsedURL, err := url.Parse(source.Link)
		if err != nil {
			host = ""
		} else {
			host = parsedURL.Host
		}

		title := fmt.Sprintf("[%s][%s] %s | %s | %s", processedMark, clusterMark, source.Title, host, source.Date.Format("2006-01-02"))
		lineIdx = drawText(a.screen, 0, lineIdx, width, currentStyle, title)

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
	helpText := fmt.Sprintf("^/v: Navigate (%d/%d) | <>: Change Page (%d/%d) | Enter: View Details | I: Show Importance", current_item+1, num_items, a.currentPage+1, num_pages)
	helpText2 := "O: Open in Browser | M: Toggle mark | S: Save | Q: Quit | [C#/O#]: Cluster Central/Outlier"
	if a.statusMessage != "" {
		helpText2 = fmt.Sprintf("%s | %s", helpText2, a.statusMessage)
	} else if a.failureMark {
		helpText2 = helpText2 + " [database F]"
	}
	if height > 0 {
		drawText(a.screen, 0, height-2, width, style, helpText)
		drawText(a.screen, 0, height-1, width, style, helpText2)
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
	helpText := "ESC/Backspace: Back to list | O: Open in Browser | M: Toggle mark | S: Save | Q: Quit"
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
				if a.detailMode {
					a.detailMode = false
					a.detailIdx = -1
				} else {
					return nil
				}
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if a.detailMode {
					a.detailMode = false
					a.detailIdx = -1
				}
			case tcell.KeyRight:
				if !a.detailMode {
					if (a.currentPage+1)*a.itemsPerPage < len(a.sources) {
						a.currentPage++
						a.selectedIdx = a.currentPage * a.itemsPerPage
					}
				}
			case tcell.KeyLeft:
				if !a.detailMode {
					if a.currentPage > 0 {
						a.currentPage--
						a.selectedIdx = a.currentPage * a.itemsPerPage
					}
				}
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q', 'Q':
					a.waitgroup.Wait() // gracefully wait for all goroutines to finish
					return nil
				case 'o', 'O':
					if len(a.sources) > 0 {
						idx := a.selectedIdx
						if a.detailMode {
							idx = a.detailIdx
						}
						openBrowser(a.sources[idx].Link)
					}
				case 'n', 'N':
					if !a.detailMode && a.selectedIdx < len(a.sources)-1 {
						a.selectedIdx++
					}
				case 'm', 'M', 'x':
					if len(a.sources) > 0 {
						idx := a.selectedIdx
						if a.detailMode {
							idx = a.detailIdx
						}
						a.markProcessed(idx, a.sources[idx])
						if !a.detailMode {
							if a.selectedIdx < len(a.sources)-1 && (a.selectedIdx+1) < (a.currentPage+1)*a.itemsPerPage {
								a.selectedIdx++
							} else if (a.currentPage+1)*a.itemsPerPage < len(a.sources) {
								a.currentPage++
								a.selectedIdx = a.currentPage * a.itemsPerPage
							}
						}
					}
				case 'X', 'p':
					if !a.detailMode {
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
					if !a.detailMode {
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
					if !a.detailMode {
						a.screen.Clear()
						a.screen.Show()
						a.currentPage = 0
						a.selectedIdx = 0
						a.loadSources()
					}
				case 's', 'S':
					if len(a.sources) > 0 {
						idx := a.selectedIdx
						if a.detailMode {
							idx = a.detailIdx
						}
						a.saveToFile(a.sources[idx])
						a.markRelevantPerHumanCheck(RELEVANT_PER_HUMAN_CHECK_YES, idx)
					}
				case 'w', 'W':
					if len(a.sources) > 0 {
						idx := a.selectedIdx
						if a.detailMode {
							idx = a.detailIdx
						}
						a.webSearch(a.sources[idx])
					}
				case 'f', 'F':
					if !a.detailMode {
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
					if !a.detailMode && len(a.sources) > 0 {
						a.showImportance[a.selectedIdx] = !a.showImportance[a.selectedIdx]
					}
				}
			case tcell.KeyUp:
				if !a.detailMode && a.selectedIdx > 0 {
					a.selectedIdx--
				}
			case tcell.KeyDown:
				if !a.detailMode && a.selectedIdx < len(a.sources)-1 && (a.selectedIdx+1) < (a.currentPage+1)*a.itemsPerPage {
					a.selectedIdx++
				}
			case tcell.KeyEnter:
				if !a.detailMode && len(a.sources) > 0 {
					a.detailMode = true
					a.detailIdx = a.selectedIdx
				} else if a.detailMode && len(a.sources) > 0 {
					a.detailMode = false
				}
			}
		case *tcell.EventResize:
			a.screen.Sync()
		}
	}
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
