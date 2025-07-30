package main

import (
  "os"
  "time"
  "fmt"
	"os/exec"
	"path/filepath"
	"runtime"
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


func (a *App) saveToFile(source Source) error {

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
	data := fmt.Sprintf("\n%s\n%s\n%s\n", source.Title, source.Summary, source.Link)
	if _, err := f.Write([]byte(data)); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	nvim_cmd := fmt.Sprintf("nvim +'$-2' %s", targetFile)
	cmd := exec.Command("/usr/bin/tmux", "new-window", nvim_cmd)
	cmd.Run()

	return nil
}


func (a *App) webSearch(source Source) {
	clean_title := cleanTitle(source.Title)
	web_search_bash_cmd := fmt.Sprintf("bash -i -c \"websearch \\\"%s\\\"\"", clean_title)
	cmd := exec.Command("/usr/bin/tmux", "new-window", web_search_bash_cmd)
	// log.Printf(web_search_bash_cmd)
	// cmd := exec.Command("/usr/bin/tmux", "new-window", "bash -c -i 'websearch 1 && bash'") //web_search_bash_cmd)
	cmd.Run()
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
