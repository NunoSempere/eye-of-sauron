package search

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"

	"github.com/gdamore/tcell/v2"
)

func (s *Search) Start() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := screen.Init(); err != nil {
		return err
	}
	defer screen.Fini()

	// Set up screen
	screen.SetStyle(tcell.StyleDefault)
	screen.Clear()

	// Initial display
	s.displayResults(screen)
	screen.Show()

	for {
		switch ev := screen.PollEvent().(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyCtrlC:
				return nil
			case tcell.KeyUp:
				if s.selected > 0 {
					s.selected--
					s.displayResults(screen)
					screen.Show()
				}
			case tcell.KeyDown:
				if s.selected < len(s.results)-1 {
					s.selected++
					s.displayResults(screen)
					screen.Show()
				}
			case tcell.KeyEnter:
				if len(s.results) > 0 {
					if err := openURL(s.results[s.selected].URL); err != nil {
						return err
					}
				}
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q', 'Q':
					screen.Fini()
					os.Exit(0)
					// return nil
				case 'o', 'O', 'r', 'R':
					if len(s.results) > 0 {
						if err := openURL(s.results[s.selected].URL); err != nil {
							return err
						}
					}
				case 's':
					clip_bash_cmd := fmt.Sprintf("echo \"%s\" | /usr/bin/xclip -sel clip", s.results[s.selected].URL)
					cmd := exec.Command("bash", "-c", clip_bash_cmd)
					cmd.Run()
				}
			}
		}
	}
}

func (s *Search) displayResults(screen tcell.Screen) {
	screen.Clear()
	writeString(screen, 0, 0, "Search results for: "+s.query)
	writeString(screen, 0, 1, "----------------------------------------")

	for i, result := range s.results {
		style := tcell.StyleDefault
		if i == s.selected {
			style = style.Reverse(true)
		}

		host := ""
		parsedURL, err := url.Parse(result.URL)
		if err == nil {
			host = parsedURL.Host
		}

		writeStringStyle(screen, 2, i+2, "["+host+"] "+result.Title, style)
	}

	writeString(screen, 0, len(s.results)+3, "Use arrow keys to navigate, Enter to open, Ctrl+C to quit")
}

func writeString(screen tcell.Screen, x, y int, str string) {
	writeStringStyle(screen, x, y, str, tcell.StyleDefault)
}

func writeStringStyle(screen tcell.Screen, x, y int, str string, style tcell.Style) {
	for i, r := range str {
		screen.SetContent(x+i, y, r, nil, style)
	}
}
