package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"runtime"
	"time"

	"html"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

type Source struct {
	ID        string
	Text      string
	Author    string
	Url       string
	CreatedAt time.Time
	Processed bool
}

type App struct {
	screen       tcell.Screen
	sources      []Source
	selectedIdx  int
	currentPage  int
	itemsPerPage int
	failureMark  bool
}

func newApp() (*App, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %v", err)
	}

	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize screen: %v", err)
	}

	return &App{
		screen:       screen,
		selectedIdx:  0,
		currentPage:  0,
		itemsPerPage: 10,
	}, nil
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
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "SELECT tweetid, text, author, url, created_at, processed FROM dwarkesh_tweets WHERE processed = false ORDER BY created_at ASC, tweetid ASC")
	if err != nil {
		return fmt.Errorf("failed to query sources: %v", err)
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var s Source
		err := rows.Scan(&s.ID, &s.Text, &s.Author, &s.Url, &s.CreatedAt, &s.Processed)
		if err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}
		// Clean HTML entities and tags
		s.Text = stripHTML(html.UnescapeString(s.Text))
		sources = append(sources, s)
	}

	a.sources = sources

	return nil
}

func padStringWithWhitespace(s string, n int) string {
	if len(s) > n {
		return s
	}
	padding := strings.Repeat(" ", n-len(s))
	return s + padding

}

func (a *App) draw() {
	a.screen.Clear()
	width, height := a.screen.Size()
	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)
	selectedStyle := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
	// selectedStyle := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	// summaryStyle := style.Foreground(tcell.ColorDimGray)

	startIdx := a.currentPage * a.itemsPerPage
	endIdx := startIdx + a.itemsPerPage
	if endIdx > len(a.sources) {
		endIdx = len(a.sources)
	}

	lineIdx := 0
	for idx := startIdx; idx < endIdx; idx++ {
		source := a.sources[idx]
		// Calculate total height needed for this item
		itemHeight := 1 // Text line

		// Check if there's enough room for the entire item
		if lineIdx+itemHeight >= height-1 { // Leave room for help text
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

		// title := fmt.Sprintf("[%s] %s | %s | %s", processedMark, padStringWithWhitespace(source.Text, 85), padStringWithWhitespace(host, 30), source.Date.Format("2006-01-02")) // why isn't the padding here working???
		title := fmt.Sprintf("[%s] %s | @%s | %s", processedMark, source.Text, source.Author, source.CreatedAt.Format("2006-01-02")) // why isn't the padding here working???
		// title := "[" + processedMark + "] " + padStringWithWhitespace(source.Text, 85) + " | " + padStringWithWhitespace(host, 30) + " | " + source.Date.Format("2006-01-02")
		lineIdx = drawText(a.screen, 0, lineIdx, width, currentStyle, title)

		lineIdx++
	}

	// Draw help text at the bottom
	current_item := a.selectedIdx
	num_items := len(a.sources)
	num_pages := int(math.Ceil(float64(num_items) / float64(a.itemsPerPage)))
	helpText := fmt.Sprintf("^/v: Navigate (%d/%d) | <>: Change Page (%d/%d) | Enter: Expand/Collapse", current_item+1, num_items, a.currentPage+1, num_pages)
	helpText2 := "O: Open in Browser \n | M: Toggle mark | S: Save | Q: Quit"
	if a.failureMark {
		helpText2 = helpText2 + " [database F]"
	}
	if height > 0 {
		drawText(a.screen, 0, height-2, width, style, helpText)
		drawText(a.screen, 0, height-1, width, style, helpText2)
	}

	a.screen.Show()
}

func markProcessedInSever(state bool, id string) error {

	flag := true
	// This function might not work correctly if there are too many items to skip over
	// the answer so far has been to add a connection pool on the database end,
	// over at digitalocean
	if flag {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_POOL_URL"))
		if err != nil {
			log.Printf("failed to connect to database: %v", err)
			return err
		}
		defer conn.Close(ctx)

		_, err = conn.Exec(ctx, "UPDATE dwarkesh_tweets SET processed = $1 WHERE tweetid = $2", state, id)
		if err != nil {
			log.Printf("failed to mark source as processed: %v", err)
			// Revert UI state if database update fails
			return err
		}
	}
	return nil
}

func (a *App) markProcessed(i int) error {
	if len(a.sources) == 0 {
		return nil
	}

	// Toggle processed state in UI immediately
	newState := !a.sources[i].Processed
	a.sources[i].Processed = newState

	// Update database asynchronously
	go func() {
		err := markProcessedInSever(newState, a.sources[i].ID)
		if err != nil {
			go func() {
				a.failureMark = true
				time.Sleep(30)
				a.failureMark = false
			}()
			a.sources[i].Processed = !newState
		}
	}()

	return nil
}

func (a *App) saveToFile(source Source) error {

	targetFile := os.Getenv("NEWSLETTER_FILE")

	f, err := os.OpenFile(targetFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	data := fmt.Sprintf("\n> %s—[@%s](%s)\n", source.Text, source.Author, source.Url)
	if _, err := f.Write([]byte(data)); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	nvim_cmd := fmt.Sprintf("nvim +'$-1' %s", targetFile)
	cmd := exec.Command("/usr/bin/tmux", "new-window", nvim_cmd)
	cmd.Run()

	return nil
}

func cleanText(s string) string {
	n := max(10, len(s)/2)
	if len(s) < n {
		return s
	} else if pos := strings.LastIndex(s[len(s)-n:], "-"); pos != -1 {
		return s[:len(s)-n+pos]
	} else {
		return s
	}
}

func (a *App) webSearch(source Source) {
	clean_title := cleanText(source.Text)
	web_search_bash_cmd := fmt.Sprintf("bash -i -c \"websearch \\\"%s\\\"\"", clean_title)
	cmd := exec.Command("/usr/bin/tmux", "new-window", web_search_bash_cmd)
	// log.Printf(web_search_bash_cmd)
	// cmd := exec.Command("/usr/bin/tmux", "new-window", "bash -c -i 'websearch 1 && bash'") //web_search_bash_cmd)
	cmd.Run()
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

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
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
				if a.confirmQuit() {
					return nil
				}
			case tcell.KeyRight:
				if (a.currentPage+1)*a.itemsPerPage < len(a.sources) {
					a.currentPage++
					a.selectedIdx = a.currentPage * a.itemsPerPage
				}
			case tcell.KeyLeft:
				if a.currentPage > 0 {
					a.currentPage--
					a.selectedIdx = a.currentPage * a.itemsPerPage
				}
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q', 'Q':
					if a.confirmQuit() {
						return nil
					}
				case 'o', 'O':
					if len(a.sources) > 0 {
						openBrowser(a.sources[a.selectedIdx].Url)
					}
				case 'n', 'N':
					if a.selectedIdx < len(a.sources)-1 {
						a.selectedIdx++
					}
				case 'm', 'M', 'x':
					if len(a.sources) > 0 {
						a.markProcessed(a.selectedIdx)
						if a.selectedIdx < len(a.sources)-1 && (a.selectedIdx+1) < (a.currentPage+1)*a.itemsPerPage {
							a.selectedIdx++
						} else if (a.currentPage+1)*a.itemsPerPage < len(a.sources) {
							a.currentPage++
							a.selectedIdx = a.currentPage * a.itemsPerPage
						}
					}
				case 'X':
					startIdx := a.currentPage * a.itemsPerPage
					endIdx := startIdx + a.itemsPerPage
					if endIdx > len(a.sources) {
						endIdx = len(a.sources)
					}
					for idx := startIdx; idx < endIdx; idx++ {
						a.markProcessed(idx)
					}
				case 'r':
					a.screen.Clear()
					a.screen.Show()
					a.currentPage = 0
					a.selectedIdx = 0
					a.sources = filterSourcesForUnread(a.sources)
				case 'R':
					a.screen.Clear()
					a.screen.Show()
					a.currentPage = 0
					a.selectedIdx = 0
					a.loadSources()
				case 's', 'S':
					if len(a.sources) > 0 {
						a.saveToFile(a.sources[a.selectedIdx])
					}
				case 'w', 'W':
					a.webSearch(a.sources[a.selectedIdx])
				}
			case tcell.KeyUp:
				if a.selectedIdx > 0 {
					a.selectedIdx--
				}
			case tcell.KeyDown:
				if a.selectedIdx < len(a.sources)-1 && (a.selectedIdx+1) < (a.currentPage+1)*a.itemsPerPage {
					a.selectedIdx++
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

func main() {
	logFile, err := os.OpenFile("src/client.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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
