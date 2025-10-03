package main

import (
  "os"
  "time"
  "fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"nunosempere.com/eye-of-sauron/client/src/search"
)

func (a *App) openFile() error {

	basePath := os.Getenv("MINUTES_FOLDER")

	now := time.Now()
	year, week := now.ISOWeek()
	dirName := fmt.Sprintf("%d-%02d", year, week)

	targetDir := filepath.Join(basePath, dirName)

	_, err := os.Stat(targetDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(targetDir, 0755)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	targetFile := filepath.Join(targetDir, "own.md")

	nvim_cmd := fmt.Sprintf("nvim +'$-2' %s", targetFile)
	cmd := exec.Command("/usr/bin/tmux", "new-window", nvim_cmd)
	cmd.Run()

	return nil
}


// Helper function to append content to minutes file
func (a *App) appendToMinutesFile(title, summary, link string) error {
	basePath := os.Getenv("MINUTES_FOLDER")

	now := time.Now()
	year, week := now.ISOWeek()
	dirName := fmt.Sprintf("%d-%02d", year, week)

	targetDir := filepath.Join(basePath, dirName)

	_, err := os.Stat(targetDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(targetDir, 0755)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	targetFile := filepath.Join(targetDir, "own.md")

	f, err := os.OpenFile(targetFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data := fmt.Sprintf("\n%s\n%s\n%s\n", title, summary, link)
	if _, err := f.Write([]byte(data)); err != nil {
		return err
	}

	nvim_cmd := fmt.Sprintf("nvim +'$-2' %s", targetFile)
	cmd := exec.Command("/usr/bin/tmux", "new-window", nvim_cmd)
	cmd.Run()

	return nil
}

func (a *App) saveToFile(source Source) error {
	return a.appendToMinutesFile(source.Title, source.Summary, source.Link)
}

func (a *App) webSearch(source Source, sourceIdx int) error {
	clean_title := cleanTitle(source.Title)
	searchInstance, err := search.New(clean_title)
	if err != nil {
		return err
	}
	a.searchInstance = searchInstance
	a.searchOriginIdx = sourceIdx
	a.mode = "search"
	return nil
}

// Save search result to minutes file along with original source description
func (a *App) saveSearchResult(searchResult *search.Result, originalSource Source) error {
	return a.appendToMinutesFile(searchResult.Title, originalSource.Summary, searchResult.URL)
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
